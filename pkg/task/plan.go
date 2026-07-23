package task

import (
	"fmt"
	"sync"
	"time"
)

type SyncPlan struct {
	Source      string
	Destination string

	Images []ImagePlan

	StartTime time.Time
	EndTime   time.Time

	mu sync.RWMutex
}

type ImagePlan struct {
	Source      string
	Destination string

	SourceAuth      string
	DestinationAuth string

	Status string
	Error  string

	StartTime time.Time
	EndTime   time.Time
}

func NewSyncPlan(
	source string,
	destination string,
) *SyncPlan {

	return &SyncPlan{
		Source:      source,
		Destination: destination,
		Images:      make([]ImagePlan, 0),
		StartTime:   time.Now(),
	}
}

func (p *SyncPlan) AddImage(
	source string,
	destination string,
	sourceAuth string,
	destinationAuth string,
) {

	p.mu.Lock()
	defer p.mu.Unlock()

	p.Images = append(
		p.Images,
		ImagePlan{
			Source:          source,
			Destination:     destination,
			SourceAuth:      sourceAuth,
			DestinationAuth: destinationAuth,
			Status:          "pending",
			StartTime:       time.Now(),
		},
	)
}

func (p *SyncPlan) Start() {

	p.mu.Lock()
	defer p.mu.Unlock()

	p.StartTime = time.Now()
}

func (p *SyncPlan) Finish() {

	p.mu.Lock()
	defer p.mu.Unlock()

	p.EndTime = time.Now()
}

func (p *SyncPlan) MarkRunning(
	source string,
	destination string,
) {

	p.mu.Lock()
	defer p.mu.Unlock()

	for i := range p.Images {

		item := &p.Images[i]

		if item.Source == source &&
			item.Destination == destination {

			item.Status = "running"
			item.StartTime = time.Now()

			return
		}
	}
}

func (p *SyncPlan) MarkSuccess(
	source string,
	destination string,
) {

	p.mu.Lock()
	defer p.mu.Unlock()

	for i := range p.Images {

		item := &p.Images[i]

		if item.Source == source &&
			item.Destination == destination {

			item.Status = "success"
			item.EndTime = time.Now()

			return
		}
	}
}

func (p *SyncPlan) MarkFailed(
	source string,
	destination string,
	err error,
) {

	p.mu.Lock()
	defer p.mu.Unlock()

	for i := range p.Images {

		item := &p.Images[i]

		if item.Source == source &&
			item.Destination == destination {

			item.Status = "failed"

			if err != nil {
				item.Error = err.Error()
			}

			item.EndTime = time.Now()

			return
		}
	}
}

func (p *SyncPlan) Summary() map[string]int {

	p.mu.RLock()
	defer p.mu.RUnlock()

	result := map[string]int{
		"total": len(p.Images),
	}

	for _, item := range p.Images {

		result[item.Status]++
	}

	return result
}

func (p *SyncPlan) Snapshot() SyncPlan {

	p.mu.RLock()
	defer p.mu.RUnlock()

	result := SyncPlan{
		Source:      p.Source,
		Destination: p.Destination,
		StartTime:   p.StartTime,
		EndTime:     p.EndTime,
		Images:      make([]ImagePlan, len(p.Images)),
	}

	copy(
		result.Images,
		p.Images,
	)

	return result
}

func (p *SyncPlan) Print() {

	p.mu.RLock()
	defer p.mu.RUnlock()

	fmt.Println(
		"========== Sync Plan ==========",
	)

	fmt.Printf(
		"SOURCE: %s\n",
		p.Source,
	)

	fmt.Printf(
		"DESTINATION: %s\n",
		p.Destination,
	)

	fmt.Println()

	fmt.Println(
		"IMAGES:",
	)

	for i, item := range p.Images {

		fmt.Printf(
			"[%d]\n",
			i+1,
		)

		fmt.Printf(
			"  %s\n",
			item.Source,
		)

		fmt.Println(
			"       ->",
		)

		fmt.Printf(
			"  %s\n",
			item.Destination,
		)

		fmt.Printf(
			"  status: %s\n",
			item.Status,
		)

		if item.Error != "" {

			fmt.Printf(
				"  error: %s\n",
				item.Error,
			)
		}

		fmt.Printf(
			"  start: %s\n",
			item.StartTime.Format(time.RFC3339),
		)

		if !item.EndTime.IsZero() {

			fmt.Printf(
				"  end: %s\n",
				item.EndTime.Format(time.RFC3339),
			)
		}

		fmt.Println()
	}

	fmt.Println(
		"SUMMARY:",
		p.Summary(),
	)

	fmt.Println(
		"DURATION:",
		p.EndTime.Sub(p.StartTime),
	)

	fmt.Println(
		"===============================",
	)
}
