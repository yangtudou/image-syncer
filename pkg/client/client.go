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
	"github.com/AliyunContainerService/image-syncer/pkg/utils/types"
)

// Client describes a synchronization client
type Client struct {
	taskList       *concurrent.List
	failedTaskList *concurrent.List

	taskCounter       *concurrent.Counter
	failedTaskCounter *concurrent.Counter

	successImagesList                     *concurrent.ImageList
	successImagesFile, outputImagesFormat string

	config *Config

	routineNum int
	retries    int
	logger     *logrus.Logger

	forceUpdate bool
}

// NewSyncClient creates synchronization client
func NewSyncClient(
	cfg *appconfig.Config,
	logFile string,
	successImagesFile string,
	outputImagesFormat string,
	routineNum int,
	retries int,
	osFilterList []string,
	archFilterList []string,
	forceUpdate bool,
) (*Client, error) {

	logger := NewFileLogger(logFile)

	logger.Info("initializing sync client")

	config, err := NewSyncConfigFromModel(
		cfg,
		osFilterList,
		archFilterList,
	)

	if err != nil {

		logger.WithError(err).
			Error("generate sync config failed")

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

		successImagesFile: successImagesFile,

		outputImagesFormat: outputImagesFormat,

		config: config,

		routineNum: routineNum,

		retries: retries,

		logger: logger,

		forceUpdate: forceUpdate,
	}, nil
}

// Run starts synchronization
func (c *Client) Run() error {

	start := time.Now()

	c.logger.WithFields(logrus.Fields{
		"workers": c.routineNum,
		"retry":   c.retries,
	}).Info("sync started")

	for source, dest := range c.config.ImageList {

		c.logger.WithFields(logrus.Fields{
			"source": source,
			"dest":   dest,
		}).Debug("processing sync rule")

		destList := []string{}

		switch value := dest.(type) {

		case string:

			if value != "" {
				destList = append(
					destList,
					value,
				)
			}

		case []string:

			destList = append(
				destList,
				value...,
			)

		case []interface{}:

			for _, item := range value {

				destList = append(
					destList,
					fmt.Sprintf("%v", item),
				)
			}
		}

		for _, destination := range destList {

			ruleTask, err := task.NewRuleTask(
				c.logger,
				source,
				destination,
				c.config.osFilterList,
				c.config.archFilterList,
				func(repository string) types.Auth {

					auth, exist := c.config.GetAuth(
						repository,
					)

					if !exist {

						c.logger.WithField(
							"repository",
							repository,
						).Debug(
							"auth not found, using anonymous access",
						)
					}

					return auth
				},
				c.forceUpdate,
			)

			if err != nil {

				return fmt.Errorf(
					"failed to generate rule task for %s -> %s: %v",
					source,
					destination,
					err,
				)
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
		"initial tasks created",
	)

	pool, err := ants.NewPoolWithFunc(
		c.routineNum,
		func(i interface{}) {

			tTask, ok := i.(task.Task)

			if !ok {

				c.logger.Errorf(
					"invalid task type %T",
					i,
				)

				return
			}

			c.logger.WithFields(logrus.Fields{
				"type": tTask.Type(),
				"task": tTask.String(),
			}).Debug(
				"task started",
			)

			nextTasks, message, err := tTask.Run()

			count, total := c.taskCounter.Increase()

			progress := fmt.Sprintf(
				"%d/%d",
				count,
				total,
			)

			if err != nil {

				c.failedTaskList.PushBack(
					tTask,
				)

				c.failedTaskCounter.IncreaseTotal()

				c.logger.WithFields(logrus.Fields{
					"task":     tTask.String(),
					"progress": progress,
				}).WithError(err).
					Error(
						"task failed",
					)

			} else {

				if tTask.Type() == task.ManifestType {

					c.successImagesList.Add(
						tTask.GetSource().String(),
						tTask.GetDestination().String(),
					)
				}

				c.logger.WithFields(logrus.Fields{
					"task":     tTask.String(),
					"message":  message,
					"progress": progress,
				}).Info(
					"task finished",
				)
			}

			for _, next := range nextTasks {

				c.taskList.PushFront(
					next,
				)

				c.taskCounter.IncreaseTotal()
			}
		},
	)

	if err != nil {
		return err
	}

	defer pool.Release()

	if err := c.handleTasks(pool); err != nil {

		c.logger.WithError(err).
			Error(
				"handle tasks failed",
			)
	}

	for i := 0; i < c.retries; i++ {

		_, failed := c.failedTaskCounter.Value()

		if failed == 0 {
			break
		}

		c.logger.WithField(
			"retry",
			i+1,
		).Info(
			"retry failed tasks",
		)

		c.taskCounter,
			c.failedTaskCounter =
			c.failedTaskCounter,
			concurrent.NewCounter(
				0,
				0,
			)

		if c.failedTaskList.Len() > 0 {

			c.taskList.PushBackList(
				c.failedTaskList,
			)

			c.failedTaskList.Reset()
		}

		if c.taskList.Len() > 0 {

			if err := c.handleTasks(pool); err != nil {

				c.logger.WithError(err).
					Error(
						"retry tasks failed",
					)
			}
		}
	}

	c.logger.WithFields(logrus.Fields{
		"failed":   c.failedTaskList.Len(),
		"duration": time.Since(start),
	}).Info(
		"sync finished",
	)

	if c.successImagesFile != "" {

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

			return json.NewEncoder(
				file,
			).Encode(
				c.successImagesList.Content(),
			)
		}

		return yaml.NewEncoder(
			file,
		).Encode(
			c.successImagesList.Content(),
		)
	}

	_, failed := c.failedTaskCounter.Value()

	if failed != 0 {

		return fmt.Errorf(
			"failed tasks exist",
		)
	}

	return nil
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
				time.Second,
			)

			continue
		}

		if err := pool.Invoke(
			item,
		); err != nil {
			return err
		}
	}

	return nil
}
