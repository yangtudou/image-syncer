package task

import (
	"github.com/AliyunContainerService/image-syncer/pkg/task/rule"
)

// GenerateRulePlan 从生成好的 RuleTask 列表中构造 SyncPlan 状态追踪器
func GenerateRulePlan(tasks []*rule.RuleTask) *SyncPlan {
	plan := NewSyncPlan(
		"image-sync",
		"registry",
	)

	for _, t := range tasks {
		srcAuth := t.SourceAuth()
		dstAuth := t.DestinationAuth()

		plan.AddImage(
			t.Source(),
			t.Destination(),
			srcAuth.Username+":"+srcAuth.Password,
			dstAuth.Username+":"+dstAuth.Password,
		)
	}

	return plan
}
