package mapper

import (
	"fmt"
	"os"
	"strings"
)

func WriteImageSyncer(file string, mappings []Mapping) error {
	var builder strings.Builder

	for _, item := range mappings {
		builder.WriteString(
			fmt.Sprintf(
				"%s: %s\n",
				item.Source,
				item.Target,
			),
		)
	}

	return os.WriteFile(
		file,
		[]byte(builder.String()),
		0600,
	)
}
