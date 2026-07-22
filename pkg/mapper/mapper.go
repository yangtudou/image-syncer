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

	for registry, images := range cfg.Images {
		for image, tags := range images {
			sourceImages := expandTags(registry, image, tags)

			for _, source := range sourceImages {
				mappings = append(mappings, Mapping{
					Source: source,
					Target: buildTarget(source, cfg.Target),
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

		hasAppend := false

		for _, item := range tags {
			tag := fmt.Sprintf("%v", item)

			if strings.HasPrefix(tag, "+") {
				hasAppend = true
				tag = strings.TrimPrefix(tag, "+")
			}

			result = append(result, prefix+":"+tag)
		}

		if hasAppend {
			result = append(
				[]string{prefix + ":latest"},
				result...,
			)
		}

		return result
	}

	return nil
}

func buildTarget(source string, target config.TargetConfig) string {
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
