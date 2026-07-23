package task

import (
	"github.com/AliyunContainerService/image-syncer/pkg/sync"
)

type Type string

const (
	URLType      = Type("URL")
	ManifestType = Type("Manifest")
	RuleType     = Type("Rule")
	BlobType     = Type("Blob")
)

type Task interface {

	// Run returns next tasks and result message.
	Run() ([]Task, string, error)

	// GetPrimary returns primary task.
	GetPrimary() Task

	// Runnable returns if task can run immediately.
	Runnable() bool

	// ReleaseOnce try to release dependency once.
	ReleaseOnce() bool

	// GetSource returns source image.
	GetSource() *sync.ImageSource

	// GetDestination returns destination image.
	GetDestination() *sync.ImageDestination

	// GetPlan returns sync plan.
	GetPlan() *SyncPlan

	// String returns task description.
	String() string

	// Type returns task type.
	Type() Type
}

type TaskInfo struct {
	Type        Type
	Description string
	Source      string
	Destination string
}

func DescribeTask(t Task) TaskInfo {

	info := TaskInfo{}

	if t == nil {
		return info
	}

	info.Type = t.Type()

	info.Description = t.String()

	if source := t.GetSource(); source != nil {
		info.Source = source.String()
	}

	if destination := t.GetDestination(); destination != nil {
		info.Destination = destination.String()
	}

	return info
}
