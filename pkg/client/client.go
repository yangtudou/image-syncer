package client

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/AliyunContainerService/image-syncer/pkg/concurrent"
	appconfig "github.com/AliyunContainerService/image-syncer/pkg/config"
	"github.com/AliyunContainerService/image-syncer/pkg/task"
	"github.com/AliyunContainerService/image-syncer/pkg/task/rule"
	"github.com/AliyunContainerService/image-syncer/pkg/utils/types"
)

type Client struct {
	taskList       *concurrent.List
	failedTaskList *concurrent.List

	taskCounter       *concurrent.Counter
	failedTaskCounter *concurrent.Counter

	successImagesList *concurrent.ImageList

	successImagesFile  string
	outputImagesFormat string

	config *Config

	plan *task.SyncPlan

	routineNum int
	retries    int

	logger *logrus.Logger

	forceUpdate bool
}

func NewSyncClient(
	cfg *appconfig.Config,
	logger *logrus.Logger,
	successImagesFile string,
	outputImagesFormat string,
	routineNum int,
	retries int,
	osFilterList []string,
	archFilterList []string,
	forceUpdate bool,
) (*Client, error) {

	logger.Info("creating sync client")

	config, err := NewSyncConfigFromModel(
		cfg,
		osFilterList,
		archFilterList,
	)

	if err != nil {

		return nil, fmt.Errorf(
			"generate config error: %v",
			err,
		)
	}

	return &Client{

		taskList: concurrent.NewList(),

		failedTaskList: concurrent.NewList(),

		taskCounter: concurrent.NewCounter(
			0,
			0,
		),

		failedTaskCounter: concurrent.NewCounter(
			0,
			0,
		),

		successImagesList: concurrent.NewImageList(),

		successImagesFile:  successImagesFile,
		outputImagesFormat: outputImagesFormat,

		config: config,

		plan: task.NewSyncPlan(
			"image-sync",
			"registry",
		),

		routineNum: routineNum,
		retries:    retries,

		logger: logger,

		forceUpdate: forceUpdate,
	}, nil
}

func (c *Client) Run() error {

	start := time.Now()

	c.plan.Start()

	c.logger.WithFields(logrus.Fields{
		"workers": c.routineNum,
		"retries": c.retries,
	}).Info(
		"sync started",
	)

	for source, dest := range c.config.ImageList {

		destList := normalizeDestinations(dest)

		for _, destination := range destList {

			ruleTask, err := rule.NewRuleTask(
				c.logger,
				source,
				destination,
				c.config.osFilterList,
				c.config.archFilterList,
				func(repository string) types.Auth {

					auth, _ := c.config.GetAuth(
						repository,
					)

					return auth
				},
				c.forceUpdate,
				c.plan,
			)

			if err != nil {

				return err
			}

			c.taskList.PushBack(
				ruleTask,
			)

			c.taskCounter.IncreaseTotal()

			c.logger.WithField(
				"task",
				ruleTask.String(),
			).Debug(
				"task queued",
			)
		}
	}

	_, total := c.taskCounter.Value()

	c.logger.WithField(
		"total",
		total,
	).Info(
		"tasks created",
	)

	pool, err := ants.NewPoolWithFunc(
		c.routineNum,
		c.executeTask,
	)

	if err != nil {
		return err
	}

	defer pool.Release()

	if err := c.handleTasks(pool); err != nil {
		return err
	}

	for retry := 0; retry < c.retries; retry++ {

		_, failed := c.failedTaskCounter.Value()

		if failed == 0 {
			break
		}

		c.logger.WithField(
			"retry",
			retry+1,
		).Info(
			"retry failed tasks",
		)

		oldFailed := c.failedTaskList

		c.failedTaskList = concurrent.NewList()

		c.failedTaskCounter =
			concurrent.NewCounter(
				0,
				0,
			)

		c.taskCounter =
			concurrent.NewCounter(
				0,
				0,
			)

		c.taskList.PushBackList(
			oldFailed,
		)

		if err := c.handleTasks(pool); err != nil {
			return err
		}
	}

	c.plan.Finish()

	c.plan.Print()

	c.logger.WithFields(
		logrus.Fields{
			"summary":  c.plan.Summary(),
			"duration": time.Since(start),
		},
	).Info(
		"sync completed",
	)

	c.plan.Print()

	if err := c.writeSuccessImages(); err != nil {
		return err
	}

	_, failed := c.failedTaskCounter.Value()

	if failed > 0 {

		return fmt.Errorf(
			"sync failed tasks: %d",
			failed,
		)
	}

	return nil
}

func (c *Client) executeTask(
	i interface{},
) {

	t, ok := i.(task.Task)

	if !ok {
		return
	}

	c.logger.WithFields(
		logrus.Fields{
			"type": t.Type(),
			"task": t.String(),
		},
	).Debug(
		"task running",
	)

	if t.Type() == task.URLType {

		c.plan.MarkRunning(
			t.GetSource().String(),
			t.GetDestination().String(),
		)
	}

	next, message, err := t.Run()

	count, total :=
		c.taskCounter.Increase()

	progress :=
		fmt.Sprintf(
			"%d/%d",
			count,
			total,
		)

	if err != nil {

		c.failedTaskList.PushBack(t)

		c.failedTaskCounter.IncreaseTotal()

		if t.Type() == task.URLType {

			c.plan.MarkFailed(
				t.GetSource().String(),
				t.GetDestination().String(),
				err,
			)
		}

		c.logger.WithFields(
			logrus.Fields{
				"task":     t.String(),
				"progress": progress,
			},
		).WithError(err).
			Error(
				"task failed",
			)

		return
	}

	if t.Type() == task.URLType {

		c.successImagesList.Add(
			t.GetSource().String(),
			t.GetDestination().String(),
		)

		c.plan.MarkSuccess(
			t.GetSource().String(),
			t.GetDestination().String(),
		)
	}

	c.logger.WithFields(
		logrus.Fields{
			"task":     t.String(),
			"message":  message,
			"progress": progress,
		},
	).Info(
		"task finished",
	)

	for _, nextTask := range next {

		c.taskList.PushFront(
			nextTask,
		)

		c.taskCounter.IncreaseTotal()
	}
}

func normalizeDestinations(
	value interface{},
) []string {

	result := []string{}

	switch v := value.(type) {

	case string:

		if v != "" {
			result = append(result, v)
		}

	case []string:

		result = append(
			result,
			v...,
		)

	case []interface{}:

		for _, item := range v {

			result = append(
				result,
				fmt.Sprintf("%v", item),
			)
		}
	}

	return result
}

func (c *Client) handleTasks(
	pool *ants.PoolWithFunc,
) error {

	for {

		item := c.taskList.PopFront()

		if item == nil {

			if pool.Running() == 0 {
				break
			}

			time.Sleep(
				200 * time.Millisecond,
			)

			continue
		}

		if err := pool.Invoke(item); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) writeSuccessImages() error {

	if c.successImagesFile == "" {
		return nil
	}

	file, err := os.OpenFile(
		c.successImagesFile,
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		0666,
	)

	if err != nil {
		return err
	}

	defer file.Close()

	if c.outputImagesFormat == "json" {

		return json.NewEncoder(file).
			Encode(
				c.successImagesList.Content(),
			)
	}

	return yaml.NewEncoder(file).
		Encode(
			c.successImagesList.Content(),
		)
}
