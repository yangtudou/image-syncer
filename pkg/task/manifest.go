package task

import (
	"fmt"

	"github.com/AliyunContainerService/image-syncer/pkg/concurrent"
	"github.com/AliyunContainerService/image-syncer/pkg/sync"
	"github.com/AliyunContainerService/image-syncer/pkg/utils"
	"github.com/opencontainers/go-digest"
)

type ManifestTask struct {
	source      *sync.ImageSource
	destination *sync.ImageDestination

	primary Task

	counter *concurrent.Counter

	bytes  []byte
	digest *digest.Digest
}

func NewManifestTask(
	primary Task,
	source *sync.ImageSource,
	destination *sync.ImageDestination,
	counter *concurrent.Counter,
	bytes []byte,
	digest *digest.Digest,
) *ManifestTask {

	return &ManifestTask{
		primary:     primary,
		source:      source,
		destination: destination,
		counter:     counter,
		bytes:       bytes,
		digest:      digest,
	}
}

func (m *ManifestTask) Run() ([]Task, string, error) {

	if m.destination == nil {

		return nil,
			"",
			fmt.Errorf(
				"destination is nil",
			)
	}

	if err := m.destination.PushManifest(
		m.bytes,
		m.digest,
	); err != nil {

		return nil,
			"",
			fmt.Errorf(
				"failed to put manifest: %v",
				err,
			)
	}

	if m.primary == nil {

		return nil,
			"manifest pushed",
			nil
	}

	if m.primary.ReleaseOnce() {

		return []Task{
				m.primary,
			},
			"start to sync parent manifest",
			nil
	}

	return nil,
		"manifest pushed",
		nil
}

func (m *ManifestTask) GetPrimary() Task {
	return m.primary
}

func (m *ManifestTask) Runnable() bool {

	if m.counter == nil {
		return true
	}

	count, _ := m.counter.Value()

	return count == 0
}

func (m *ManifestTask) ReleaseOnce() bool {

	if m.counter == nil {
		return true
	}

	count, _ := m.counter.Decrease()

	return count == 0
}

func (m *ManifestTask) GetSource() *sync.ImageSource {
	return m.source
}

func (m *ManifestTask) GetDestination() *sync.ImageDestination {
	return m.destination
}

func (m *ManifestTask) String() string {

	src := ""
	dst := ""

	if m.source != nil {

		src = m.source.GetTagOrDigest()
	}

	if m.destination != nil {

		dst = m.destination.GetTagOrDigest()
	}

	if m.digest != nil {

		src = m.digest.String()
		dst = m.digest.String()
	}

	sourceRegistry := ""
	sourceRepository := ""

	destinationRegistry := ""
	destinationRepository := ""

	if m.source != nil {

		sourceRegistry = m.source.GetRegistry()
		sourceRepository = m.source.GetRepository()
	}

	if m.destination != nil {

		destinationRegistry = m.destination.GetRegistry()
		destinationRepository = m.destination.GetRepository()
	}

	return fmt.Sprintf(
		"synchronizing manifest from %s/%s%s to %s/%s%s",
		sourceRegistry,
		sourceRepository,
		utils.AttachConnectorToTagOrDigest(src),
		destinationRegistry,
		destinationRepository,
		utils.AttachConnectorToTagOrDigest(dst),
	)
}

func (m *ManifestTask) Type() Type {
	return ManifestType
}
