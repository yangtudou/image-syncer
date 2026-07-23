package mapper

import (
	"fmt"
	"strings"

	"github.com/AliyunContainerService/image-syncer/pkg/config"
)

type Mapping struct {
	Source string
	Target string
}

func Generate(cfg *config.Config) []Mapping {
	var mappings []Mapping

	for registry, source := range cfg.Sources {
		for image, tags := range source.Images {
			sourceImages := expandTags(registry, image, tags)

			for _, sourceImage := range sourceImages {
				mappings = append(mappings, Mapping{
					Source: sourceImage,
					Target: buildTarget(sourceImage, cfg.Dest),
				})
			}
		}
	}

	return mappings
}

func expandTags(registry, image string, value any) []string {
	prefix := registry + "/" + image

	switch tags := value.(type) {

	case nil:
		return []string{
			prefix + ":latest",
		}

	case string:
		return []string{
			prefix + ":" + tags,
		}

	case []any:
		var result []string

		for _, item := range tags {
			tag := fmt.Sprintf("%v", item)

			result = append(result, prefix+":"+tag)
		}

		return result
	}

	return nil
}

func buildTarget(source string, target config.DestConfig) string {
	imagePath := source

	if target.Flatten {
		imagePath = flattenImage(source)
	}

	return fmt.Sprintf(
		"%s/%s/%s",
		target.Registry,
		target.Namespace,
		imagePath,
	)
}

func flattenImage(image string) string {
	parts := strings.Split(image, "/")

	if len(parts) == 0 {
		return image
	}

	return parts[len(parts)-1]
}
