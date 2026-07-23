package rule

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/AliyunContainerService/image-syncer/pkg/sync"
	"github.com/AliyunContainerService/image-syncer/pkg/utils/types"
)

// RuleTask 顶层规则任务，用于管理和生成具体的镜像同步计划
type RuleTask struct {
	logger *logrus.Logger

	source      string
	destination string

	osFilterList   []string
	archFilterList []string

	getAuthFunc    func(repository string) types.Auth
	urlTaskFactory URLTaskFactory
	forceUpdate    bool
	plan           SyncPlan
}

// NewRuleTask 创建一个新的 RuleTask 实例
func NewRuleTask(
	logger *logrus.Logger,
	source string,
	destination string,
	osFilterList []string,
	archFilterList []string,
	getAuthFunc func(repository string) types.Auth,
	forceUpdate bool,
	plan SyncPlan,
) (*RuleTask, error) {
	if source == "" {
		return nil, fmt.Errorf("source url should not be empty")
	}

	if destination == "" {
		return nil, fmt.Errorf("destination url should not be empty")
	}

	return &RuleTask{
		logger:         logger,
		source:         source,
		destination:    destination,
		osFilterList:   osFilterList,
		archFilterList: archFilterList,
		getAuthFunc:    getAuthFunc,
		forceUpdate:    forceUpdate,
		plan:           plan,
	}, nil
}

func (r *RuleTask) GetPrimary() any {
	return types.Auth{}
}

func (r *RuleTask) Runnable() bool {
	return true
}

func (r *RuleTask) ReleaseOnce() bool {
	return true
}

// GetSource 规则容器任务不直接操作底层的二进制 layer，返回 nil 即可
func (r *RuleTask) GetSource() *sync.ImageSource {
	return nil
}

// GetDestination 规则容器任务不直接操作底层的二进制 layer，返回 nil 即可
func (r *RuleTask) GetDestination() *sync.ImageDestination {
	return nil
}

func (r *RuleTask) GetPlan() SyncPlan {
	return r.plan
}

func (r *RuleTask) String() string {
	return fmt.Sprintf(
		"RuleTask %s -> %s",
		r.source,
		r.destination,
	)
}

func (r *RuleTask) Type() string {
	return "rule"
}





// Getter methods for RuleTask
func (t *RuleTask) Source() string {
	return t.source
}

func (t *RuleTask) Destination() string {
	return t.destination
}

func (t *RuleTask) SourceAuth() types.Auth {
	if t.getAuthFunc != nil {
		return t.getAuthFunc(t.source)
	}
	return types.Auth{}
}

func (t *RuleTask) DestinationAuth() types.Auth {
	if t.getAuthFunc != nil {
		return t.getAuthFunc(t.destination)
	}
	return types.Auth{}
}




