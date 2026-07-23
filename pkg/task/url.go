package task

import (
	"fmt"

	"github.com/AliyunContainerService/image-syncer/pkg/concurrent"
	imagesync "github.com/AliyunContainerService/image-syncer/pkg/sync"
	"github.com/AliyunContainerService/image-syncer/pkg/utils/types"

	"github.com/containers/image/v5/manifest"
)

type URLTask struct {
	source      *imagesync.ImageSource
	destination *imagesync.ImageDestination

	sourceAuth      types.Auth
	destinationAuth types.Auth

	osFilterList   []string
	archFilterList []string

	forceUpdate bool

	manifestTask *ManifestTask

	plan *SyncPlan
}

func NewURLTask(
	source *imagesync.ImageSource,
	destination *imagesync.ImageDestination,
	sourceAuth types.Auth,
	destinationAuth types.Auth,
	osFilterList []string,
	archFilterList []string,
	forceUpdate bool,
	plan *SyncPlan,
) *URLTask {

	return &URLTask{
		source:          source,
		destination:     destination,
		sourceAuth:      sourceAuth,
		destinationAuth: destinationAuth,
		osFilterList:    osFilterList,
		archFilterList:  archFilterList,
		forceUpdate:     forceUpdate,
		plan:            plan,
	}
}

func (u *URLTask) Run() ([]Task, string, error) {

	if u.manifestTask != nil {

		return nil,
			"image sync finished",
			nil
	}

	manifestBytes, manifestType, err :=
		u.source.GetManifest()

	if err != nil {

		return nil,
			"",
			fmt.Errorf(
				"failed to get manifest: %v",
				err,
			)
	}

	obj, bytes, subManifestInfos, err :=
		imagesync.GenerateManifestObj(
			manifestBytes,
			manifestType,
			u.osFilterList,
			u.archFilterList,
			u.source,
			nil,
		)

	if err != nil {

		return nil,
			"",
			fmt.Errorf(
				"failed to generate manifest object: %v",
				err,
			)
	}

	if obj == nil {

		return nil,
			"manifest ignored by platform filter",
			nil
	}

	result := make(
		[]Task,
		0,
	)

	if len(subManifestInfos) > 0 {

		counter :=
			concurrent.NewCounter(
				len(subManifestInfos),
				len(subManifestInfos),
			)

		rootTask :=
			NewManifestTask(
				u,
				u.source,
				u.destination,
				counter,
				bytes,
				nil,
				u.plan,
			)

		u.manifestTask = rootTask

		for _, sub := range subManifestInfos {

			childBlobInfos, err :=
				u.source.GetBlobInfos(
					sub.Obj,
				)

			if err != nil {

				return nil,
					"",
					fmt.Errorf(
						"failed to get child blob infos: %v",
						err,
					)
			}

			childCounter :=
				concurrent.NewCounter(
					len(childBlobInfos),
					len(childBlobInfos),
				)

			childManifestTask :=
				NewManifestTask(
					rootTask,
					u.source,
					u.destination,
					childCounter,
					sub.Bytes,
					sub.Digest,
					u.plan,
				)

			for _, info := range childBlobInfos {

				result = append(
					result,
					NewBlobTask(
						childManifestTask,
						info,
						u.plan,
					),
				)
			}

			if len(childBlobInfos) == 0 {

				result = append(
					result,
					childManifestTask,
				)
			}
		}

		if len(result) == 0 {

			result = append(
				result,
				rootTask,
			)
		}

		return result,
			"start to sync manifest list",
			nil
	}

	mainManifest, ok :=
		obj.(manifest.Manifest)

	if !ok {

		return nil,
			"",
			fmt.Errorf(
				"invalid manifest object type %T",
				obj,
			)
	}

	blobInfos, err :=
		u.source.GetBlobInfos(
			mainManifest,
		)

	if err != nil {

		return nil,
			"",
			fmt.Errorf(
				"failed to get blob infos: %v",
				err,
			)
	}

	counter :=
		concurrent.NewCounter(
			len(blobInfos),
			len(blobInfos),
		)

	manifestTask :=
		NewManifestTask(
			u,
			u.source,
			u.destination,
			counter,
			bytes,
			nil,
			u.plan,
		)

	u.manifestTask = manifestTask

	for _, info := range blobInfos {

		result = append(
			result,
			NewBlobTask(
				manifestTask,
				info,
				u.plan,
			),
		)
	}

	if len(blobInfos) == 0 {

		result = append(
			result,
			manifestTask,
		)
	}

	return result,
		"start to sync image",
		nil
}

func (u *URLTask) GetPrimary() Task {
	return u
}

func (u *URLTask) Runnable() bool {
	return true
}

func (u *URLTask) ReleaseOnce() bool {
	return true
}

func (u *URLTask) GetSource() *imagesync.ImageSource {
	return u.source
}

func (u *URLTask) GetDestination() *imagesync.ImageDestination {
	return u.destination
}

func (u *URLTask) GetPlan() *SyncPlan {
	return u.plan
}

func (u *URLTask) String() string {

	return fmt.Sprintf(
		"copy image %s -> %s",
		u.source.String(),
		u.destination.String(),
	)
}

func (u *URLTask) Type() Type {
	return URLType
}
