package task

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/AliyunContainerService/image-syncer/pkg/sync"
	"github.com/AliyunContainerService/image-syncer/pkg/utils"
	"github.com/AliyunContainerService/image-syncer/pkg/utils/types"
)

// RuleTask analyze an image config rule and generates URLTask(s).
type RuleTask struct {
	logger *logrus.Logger

	source      string
	destination string

	osFilterList, archFilterList []string

	getAuthFunc func(repository string) types.Auth

	forceUpdate bool
}

func NewRuleTask(
	logger *logrus.Logger,
	source string,
	destination string,
	osFilterList, archFilterList []string,
	getAuthFunc func(repository string) types.Auth,
	forceUpdate bool,
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
		getAuthFunc:    getAuthFunc,
		osFilterList:   osFilterList,
		archFilterList: archFilterList,
		forceUpdate:    forceUpdate,
	}, nil
}

func (r *RuleTask) Run() ([]Task, string, error) {

	r.logger.WithFields(logrus.Fields{
		"source":      r.source,
		"destination": r.destination,
	}).Debug("analyzing image rule")

	sourceURLs, err := utils.GenerateRepoURLs(
		r.source,
		func(registry, repository string) ([]string, error) {
			return []string{"latest"}, nil
		},
	)

	if err != nil {
		return nil, "", fmt.Errorf(
			"source url %s format error: %v",
			r.source,
			err,
		)
	}

	destinationURLs, err := utils.GenerateRepoURLs(
		r.destination,
		func(registry, repository string) ([]string, error) {

			tags := make([]string, 0, len(sourceURLs))

			for _, item := range sourceURLs {
				tags = append(
					tags,
					item.GetTagOrDigest(),
				)
			}

			return tags, nil
		},
	)

	if err != nil {
		return nil, "", fmt.Errorf(
			"destination url %s format error: %v",
			r.destination,
			err,
		)
	}

	if len(sourceURLs) != len(destinationURLs) {
		return nil, "", fmt.Errorf(
			"source and destination tag count mismatch: source=%d destination=%d",
			len(sourceURLs),
			len(destinationURLs),
		)
	}

	results := make(
		[]Task,
		0,
		len(sourceURLs),
	)

	for i, sourceURL := range sourceURLs {

		destinationURL := destinationURLs[i]

		sourceAuth := r.getAuthFunc(
			sourceURL.GetURLWithoutTagOrDigest(),
		)

		destinationAuth := r.getAuthFunc(
			destinationURL.GetURLWithoutTagOrDigest(),
		)

		sourceImage, err := sync.NewImageSource(
			sourceURL.GetRegistry(),
			sourceURL.GetRepo(),
			sourceURL.GetTagOrDigest(),
			sourceAuth.Username,
			sourceAuth.Password,
			sourceAuth.Insecure,
		)

		if err != nil {
			return nil, "", fmt.Errorf(
				"create source image error: %v",
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
			return nil, "", fmt.Errorf(
				"create destination image error: %v",
				err,
			)
		}

		results = append(
			results,
			NewURLTask(
				sourceImage,
				destinationImage,
				sourceAuth,
				destinationAuth,
				r.osFilterList,
				r.archFilterList,
				r.forceUpdate,
			),
		)
	}

	return results, "", nil
}

func (r *RuleTask) GetPrimary() Task {
	return nil
}

func (r *RuleTask) Runnable() bool {
	return true
}

func (r *RuleTask) ReleaseOnce() bool {
	return true
}

func (r *RuleTask) GetSource() *sync.ImageSource {
	return nil
}

func (r *RuleTask) GetDestination() *sync.ImageDestination {
	return nil
}

func (r *RuleTask) String() string {
	return fmt.Sprintf(
		"RuleTask %s -> %s",
		r.source,
		r.destination,
	)
}

func (r *RuleTask) Type() Type {
	return RuleType
}
