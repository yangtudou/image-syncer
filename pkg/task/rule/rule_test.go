package rule

import (
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/AliyunContainerService/image-syncer/pkg/config"
	"github.com/AliyunContainerService/image-syncer/pkg/utils/types"
)

// 测试：通过真实项目配置文件 examples/sync.yaml 加载并生成 RuleTask
func TestGenerateRuleTasksFromRealExampleFile(t *testing.T) {
	logger := logrus.New()
	mockAuthFunc := func(repository string) types.Auth {
		return types.Auth{Username: "test-user", Password: "test-password"}
	}

	// 相对路径指向项目根目录下的 examples/sync.yaml
	examplePath := filepath.Join("..", "..", "..", "examples", "sync.yaml")

	cfg, err := config.Load(examplePath)
	assert.NoError(t, err, "读取真实的 examples/sync.yaml 失败")

	tasks, err := GenerateRuleTasksFromConfig(
		cfg,
		logger,
		[]string{"linux"},
		[]string{"amd64"},
		mockAuthFunc,
		false,
		nil,
	)

	assert.NoError(t, err)
	assert.NotEmpty(t, tasks, "解析 examples/sync.yaml 应当成功生成 RuleTask")

	t.Logf("成功从 examples/sync.yaml 加载并生成了 %d 个任务：", len(tasks))
	for _, task := range tasks {
		t.Logf("  - %s -> %s", task.source, task.destination)
	}
}

// 基础参数校验测试
func TestNewRuleTaskValidation(t *testing.T) {
	logger := logrus.New()
	mockAuthFunc := func(repository string) types.Auth {
		return types.Auth{}
	}

	t.Run("为空的 source 应该报错", func(t *testing.T) {
		_, err := NewRuleTask(logger, "", "dest:latest", nil, nil, mockAuthFunc, false, nil)
		assert.Error(t, err)
	})

	t.Run("为空的 destination 应该报错", func(t *testing.T) {
		_, err := NewRuleTask(logger, "src:latest", "", nil, nil, mockAuthFunc, false, nil)
		assert.Error(t, err)
	})
}
