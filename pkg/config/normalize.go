package config

import (
	"fmt"
	"os"
)

func (c *Config) NormalizeRules() []ImageRule {

	rules := make(
		[]ImageRule,
		0,
	)

	for registry, source := range c.Sources {

		for repository, value := range source.Images {

			rules = append(
				rules,
				ImageRule{
					Registry:   registry,
					Repository: repository,
					Tags:       normalizeTags(value),
				},
			)
		}
	}

	return rules
}

func normalizeTags(
	value interface{},
) []string {

	tags := make(
		[]string,
		0,
	)

	switch v := value.(type) {

	case nil:

		tags = append(
			tags,
			"latest",
		)

	case string:

		tags = append(
			tags,
			expandTag(v),
		)

	case []interface{}:

		for _, item := range v {

			tags = append(
				tags,
				expandTag(
					fmt.Sprintf("%v", item),
				),
			)
		}

	case []string:

		for _, item := range v {

			tags = append(
				tags,
				expandTag(item),
			)
		}
	}

	if len(tags) == 0 {

		tags = append(
			tags,
			"latest",
		)
	}

	return tags
}

func expandTag(
	tag string,
) string {

	tag = os.ExpandEnv(tag)

	if tag == "" {
		return "latest"
	}

	return tag
}
