package rule

import (
	"fmt"

	"github.com/AliyunContainerService/image-syncer/pkg/sync"
	"github.com/AliyunContainerService/image-syncer/pkg/utils"
	"github.com/AliyunContainerService/image-syncer/pkg/utils/types"
)

// URLTaskFactory 定义一个创建 Task 的工厂函数签名，避免直接 import pkg/task
type URLTaskFactory func(
	sourceImage *sync.ImageSource,
	destinationImage *sync.ImageDestination,
	sourceAuth, destinationAuth types.Auth,
	osFilterList, archFilterList []string,
	forceUpdate bool,
	plan SyncPlan,
) any // 或者替换为你定义的通用 Task 接口

func (r *RuleTask) createSyncImages(
	sourceURL, destinationURL *utils.RepoURL,
) (*sync.ImageSource, *sync.ImageDestination, types.Auth, types.Auth, error) {
	sourceAuth := r.getAuthFunc(sourceURL.GetURLWithoutTagOrDigest())
	destinationAuth := r.getAuthFunc(destinationURL.GetURLWithoutTagOrDigest())

	sourceImage, err := sync.NewImageSource(
		sourceURL.GetRegistry(),
		sourceURL.GetRepo(),
		sourceURL.GetTagOrDigest(),
		sourceAuth.Username,
		sourceAuth.Password,
		sourceAuth.Insecure,
	)
	if err != nil {
		return nil, nil, sourceAuth, destinationAuth, fmt.Errorf(
			"create source image %s failed: %w",
			sourceURL.String(),
			err,
		)
	}

	destinationImage, err := sync.NewImageDestination(
		destinationURL.GetRegistry(),
		destinationURL.GetRepo(),
		destinationURL.GetTagOrDigest(),
		destinationAuth.Username,
		destinationAuth.Password,
		destinationAuth.Insecure,
	)
	if err != nil {
		return nil, nil, sourceAuth, destinationAuth, fmt.Errorf(
			"create destination image %s failed: %w",
			destinationURL.String(),
			err,
		)
	}

	return sourceImage, destinationImage, sourceAuth, destinationAuth, nil
}

// BuildURLTask 保持在这里！只需额外接收一个 factory 函数
func (r *RuleTask) BuildURLTask(
	sourceURL, destinationURL *utils.RepoURL,
	factory URLTaskFactory, // 👈 动态注入创建逻辑，斩断循环依赖
) (any, error) {
	sourceImage, destinationImage, sourceAuth, destinationAuth, err := r.createSyncImages(sourceURL, destinationURL)
	if err != nil {
		return nil, err
	}

	return factory(
		sourceImage,
		destinationImage,
		sourceAuth,
		destinationAuth,
		r.osFilterList,
		r.archFilterList,
		r.forceUpdate,
		r.plan,
	), nil
}
