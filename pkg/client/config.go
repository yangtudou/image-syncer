package client

import (
	"fmt"
	"os"
	"strings"

	appconfig "github.com/AliyunContainerService/image-syncer/pkg/config"
	"github.com/AliyunContainerService/image-syncer/pkg/utils"
	"github.com/AliyunContainerService/image-syncer/pkg/utils/types"
)

// Config information of sync client
type Config struct {
	AuthList map[string]types.Auth

	ImageList map[string]interface{}

	osFilterList   []string
	archFilterList []string
}

// NewSyncConfigFromModel converts yaml config model into client config
func NewSyncConfigFromModel(
	cfg *appconfig.Config,
	osFilterList []string,
	archFilterList []string,
) (*Config, error) {

	config := &Config{
		AuthList:       make(map[string]types.Auth),
		ImageList:      make(map[string]interface{}),
		osFilterList:   osFilterList,
		archFilterList: archFilterList,
	}

	if cfg.Dest.Auth != nil {
		config.AuthList[cfg.Dest.Registry] = types.Auth{
			Username: cfg.Dest.Auth.Username,
			Password: cfg.Dest.Auth.Password,
		}
	}

	for registry, source := range cfg.Sources {

		if source.Auth != nil {
			config.AuthList[registry] = types.Auth{
				Username: source.Auth.Username,
				Password: source.Auth.Password,
			}
		}

		for image, value := range source.Images {

			sourceBase := buildSourceImage(
				registry,
				image,
			)

			destinationBase := buildDestinationImage(
				cfg.Dest,
				image,
			)

			config.ImageList[sourceBase] =
				convertTagsWithDestination(
					destinationBase,
					value,
				)
		}
	}

	return config, nil
}

func buildSourceImage(
	registry string,
	image string,
) string {

	return strings.TrimSuffix(
		registry,
		"/",
	) + "/" + strings.TrimPrefix(
		image,
		"/",
	)
}

func buildDestinationImage(
	dest appconfig.DestConfig,
	image string,
) string {

	image = strings.Trim(
		image,
		"/",
	)

	name := image

	if dest.Flatten {

		parts := strings.Split(
			image,
			"/",
		)

		name = parts[len(parts)-1]
	}

	result := strings.TrimSuffix(
		dest.Registry,
		"/",
	)

	if dest.Namespace != "" {

		result += "/" +
			strings.Trim(
				dest.Namespace,
				"/",
			)
	}

	return result + "/" + name
}

func convertTagsWithDestination(
	destination string,
	value any,
) any {

	switch v := value.(type) {

	case nil:

		return destination + ":latest"

	case string:

		tag := os.ExpandEnv(v)

		if tag == "" {
			tag = "latest"
		}

		return destination + ":" + tag

	case []interface{}:

		result := make(
			[]string,
			0,
			len(v),
		)

		for _, item := range v {

			tag := os.ExpandEnv(
				fmt.Sprintf(
					"%v",
					item,
				),
			)

			if tag == "" {
				tag = "latest"
			}

			result = append(
				result,
				destination+":"+tag,
			)
		}

		return result

	case []string:

		result := make(
			[]string,
			0,
			len(v),
		)

		for _, item := range v {

			tag := os.ExpandEnv(
				item,
			)

			if tag == "" {
				tag = "latest"
			}

			result = append(
				result,
				destination+":"+tag,
			)
		}

		return result
	}

	return destination + ":latest"
}

func (c *Config) GetAuth(repository string) (types.Auth, bool) {

	auth := types.Auth{}

	prefixLen := 0

	exist := false

	for key, value := range c.AuthList {

		if utils.RepoMathPrefix(
			repository,
			key,
		) {

			if len(key) > prefixLen {

				auth = value
				prefixLen = len(key)
				exist = true
			}
		}
	}

	return auth, exist
}
