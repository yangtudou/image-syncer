package config

type ImageRule struct {

	// source registry
	Registry string

	// repository
	Repository string

	// tags
	Tags []string
}
