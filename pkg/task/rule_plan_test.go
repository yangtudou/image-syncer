package task

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/AliyunContainerService/image-syncer/pkg/task/rule"
	"github.com/AliyunContainerService/image-syncer/pkg/utils/types"
)

func TestGenerateRulePlan(t *testing.T) {
	logger := logrus.New()
	mockAuth := func(repo string) types.Auth {
		return types.Auth{Username: "user", Password: "pwd"}
	}

	// 1. 构造 RuleTask 测试数据
	task1, err := rule.NewRuleTask(logger, "docker.io/library/nginx:1.21", "registry.com/ns/nginx:1.21", nil, nil, mockAuth, false, nil)
	assert.NoError(t, err)

	task2, err := rule.NewRuleTask(logger, "docker.io/library/redis:alpine", "registry.com/ns/redis:alpine", nil, nil, mockAuth, false, nil)
	assert.NoError(t, err)

	// 2. 生成 SyncPlan
	plan := GenerateRulePlan([]*rule.RuleTask{task1, task2})

	// 3. 校验 Plan 状态
	assert.NotNil(t, plan)
	snapshot := plan.Snapshot()
	assert.Len(t, snapshot.Images, 2)

	assert.Equal(t, "docker.io/library/nginx:1.21", snapshot.Images[0].Source)
	assert.Equal(t, "registry.com/ns/nginx:1.21", snapshot.Images[0].Destination)
	assert.Equal(t, "pending", snapshot.Images[0].Status)

	summary := plan.Summary()
	assert.Equal(t, 2, summary["total"])
	assert.Equal(t, 2, summary["pending"])
}
