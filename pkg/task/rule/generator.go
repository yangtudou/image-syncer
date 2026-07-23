package rule

import (
	"fmt"
	"path" // 💡 顺手改用 path，避免 Windows 下 filepath 把 '/' 搞成 '\'

	"github.com/sirupsen/logrus"

	"github.com/AliyunContainerService/image-syncer/pkg/config"
	"github.com/AliyunContainerService/image-syncer/pkg/utils"
	"github.com/AliyunContainerService/image-syncer/pkg/utils/types"
)

// SyncPlan 接口：用于解除对 pkg/task 的直接依赖
type SyncPlan interface {
	AddImage(source, destination, sourceAuth, destinationAuth string)
}

// GenerateRuleTasksFromConfig 直接承接上游 config.Config，批量生成 RuleTask 列表
func GenerateRuleTasksFromConfig(
	cfg *config.Config,
	logger *logrus.Logger,
	osFilterList []string,
	archFilterList []string,
	getAuthFunc func(repository string) types.Auth,
	forceUpdate bool,
	plan SyncPlan, // 👈 这里的 *task.SyncPlan 改成了 interface
) ([]*RuleTask, error) {
	rules := cfg.NormalizeRules()
	tasks := make([]*RuleTask, 0, len(rules))

	for _, rule := range rules {
		for _, tag := range rule.Tags {
			// 1. 拼接完整源镜像地址
			source := fmt.Sprintf(
				"%s/%s:%s",
				rule.Registry,
				rule.Repository,
				tag,
			)

			// 2. 根据 Flatten 策略决定 Repo 名称
			repoName := rule.Repository
			if cfg.Dest.Flatten {
				// 💡 镜像路径固定用 '/'，统一用 path.Base，防 Windows 踩坑
				repoName = path.Base(rule.Repository)
			}

			// 3. 兼容 Namespace 为空的情况，防止拼出双斜杠 '//'
			var destination string
			if cfg.Dest.Namespace != "" {
				destination = fmt.Sprintf(
					"%s/%s/%s:%s",
					cfg.Dest.Registry,
					cfg.Dest.Namespace,
					repoName,
					tag,
				)
			} else {
				destination = fmt.Sprintf(
					"%s/%s:%s",
					cfg.Dest.Registry,
					repoName,
					tag,
				)
			}

			t, err := NewRuleTask(
				logger,
				source,
				destination,
				osFilterList,
				archFilterList,
				getAuthFunc,
				forceUpdate,
				plan,
			)
			if err != nil {
				return nil, fmt.Errorf("create RuleTask failed for %s: %w", source, err)
			}

			tasks = append(tasks, t)
		}
	}

	return tasks, nil
}

// generateRepoURLs 解析并匹配源与目标的仓库 URL 列表
func (r *RuleTask) generateRepoURLs() ([]*utils.RepoURL, []*utils.RepoURL, error) {
	sourceURLs, err := utils.GenerateRepoURLs(
		r.source,
		func(registry, repository string) ([]string, error) {
			return []string{"latest"}, nil
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"parse source image %s failed: %v",
			r.source,
			err,
		)
	}

	destinationURLs, err := utils.GenerateRepoURLs(
		r.destination,
		func(registry, repository string) ([]string, error) {
			tags := make([]string, 0, len(sourceURLs))
			for _, item := range sourceURLs {
				tags = append(tags, item.GetTagOrDigest())
			}
			return tags, nil
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"parse destination image %s failed: %v",
			r.destination,
			err,
		)
	}

	if len(sourceURLs) != len(destinationURLs) {
		return nil, nil, fmt.Errorf(
			"image count mismatch source=%d destination=%d",
			len(sourceURLs),
			len(destinationURLs),
		)
	}

	return sourceURLs, destinationURLs, nil
}
