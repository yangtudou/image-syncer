package task

import (
	"fmt"

	"github.com/docker/go-units"

	"github.com/AliyunContainerService/image-syncer/pkg/sync"
	"github.com/containers/image/v5/types"
)

// BlobTask sync a blob which belongs to the primary ManifestTask.
type BlobTask struct {
	primary Task

	info types.BlobInfo

	plan *SyncPlan
}

func NewBlobTask(
	manifestTask Task,
	info types.BlobInfo,
	plan *SyncPlan,
) *BlobTask {

	return &BlobTask{
		primary: manifestTask,
		info:    info,
		plan:    plan,
	}
}

func (b *BlobTask) Run() ([]Task, string, error) {

	var resultMsg string

	dst := b.primary.GetDestination()
	src := b.primary.GetSource()

	blobExist, err := dst.CheckBlobExist(
		b.info,
	)

	if err != nil {

		return nil,
			resultMsg,
			fmt.Errorf(
				"failed to check blob %s(%v) exist: %v",
				b.info.Digest,
				b.info.Size,
				err,
			)
	}

	if !blobExist {

		blob, size, err := src.GetABlob(
			b.info,
		)

		if err != nil {

			return nil,
				resultMsg,
				fmt.Errorf(
					"failed to get blob %s(%v): %v",
					b.info.Digest,
					size,
					err,
				)
		}

		b.info.Size = size

		if err = dst.PutABlob(
			blob,
			b.info,
		); err != nil {

			return nil,
				resultMsg,
				fmt.Errorf(
					"failed to put blob %s(%v): %v",
					b.info.Digest,
					b.info.Size,
					err,
				)
		}

	} else {

		resultMsg = "ignore exist blob"
	}

	if b.primary.ReleaseOnce() {

		resultMsg = "start to sync manifest"

		return []Task{
				b.primary,
			},
			resultMsg,
			nil
	}

	return nil, resultMsg, nil
}

func (b *BlobTask) GetPrimary() Task {

	return b.primary
}

func (b *BlobTask) Runnable() bool {

	return true
}

func (b *BlobTask) ReleaseOnce() bool {

	return true
}

func (b *BlobTask) GetSource() *sync.ImageSource {

	return b.primary.GetSource()
}

func (b *BlobTask) GetDestination() *sync.ImageDestination {

	return b.primary.GetDestination()
}

func (b *BlobTask) GetPlan() *SyncPlan {

	return b.plan
}

func (b *BlobTask) String() string {

	return fmt.Sprintf(
		"synchronizing blob %s(%v) from %s to %s",
		b.info.Digest,
		units.HumanSize(float64(b.info.Size)),
		b.GetSource().String(),
		b.GetDestination().String(),
	)
}

func (b *BlobTask) Type() Type {

	return BlobType
}
