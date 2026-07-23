package task

import (
	"fmt"
)

type SyncPlan struct {
	Source      string
	Destination string

	Images []ImagePlan
}

type ImagePlan struct {
	Source      string
	Destination string

	SourceAuth      string
	DestinationAuth string
}

func (p *SyncPlan) Print() {
	fmt.Println("========== Sync Plan ==========")

	fmt.Println("SOURCE:")
	fmt.Println(" ", p.Source)

	fmt.Println()

	fmt.Println("DESTINATION:")
	fmt.Println(" ", p.Destination)

	fmt.Println()

	fmt.Println("IMAGES:")

	for i, item := range p.Images {
		fmt.Printf("[%d]\n", i+1)

		fmt.Println("  ", item.Source)
		fmt.Println("       ->")
		fmt.Println("  ", item.Destination)

		fmt.Println("  source auth:", item.SourceAuth)
		fmt.Println("  dest auth:  ", item.DestinationAuth)

		fmt.Println()
	}

	fmt.Println("===============================")
}
