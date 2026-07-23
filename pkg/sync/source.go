package sync

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/AliyunContainerService/image-syncer/pkg/utils"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/types"
)

type ImageSource struct {
	ref    types.ImageReference
	source types.ImageSource
	ctx    context.Context
	sysctx *types.SystemContext

	registry    string
	repository  string
	tagOrDigest string
}

func NewImageSource(
	registry string,
	repository string,
	tagOrDigest string,
	username string,
	password string,
	insecure bool,
) (*ImageSource, error) {

	if strings.Contains(repository, ":") {
		return nil, fmt.Errorf(
			"repository string should not include ':'",
		)
	}

	ref, err := docker.ParseReference(
		"//" + registry + "/" + repository +
			utils.AttachConnectorToTagOrDigest(tagOrDigest),
	)

	if err != nil {
		return nil, err
	}

	sysctx := &types.SystemContext{}

	if insecure {
		sysctx.DockerInsecureSkipTLSVerify =
			types.OptionalBoolTrue
	}

	if username != "" && password != "" {
		sysctx.DockerAuthConfig =
			&types.DockerAuthConfig{
				Username: username,
				Password: password,
			}
	}

	ctx := context.WithValue(
		context.Background(),
		utils.CTXKey("ImageSource"),
		repository,
	)

	var source types.ImageSource

	if tagOrDigest != "" {

		source, err = ref.NewImageSource(
			ctx,
			sysctx,
		)

		if err != nil {
			return nil, err
		}
	}

	return &ImageSource{
		ref:         ref,
		source:      source,
		ctx:         ctx,
		sysctx:      sysctx,
		registry:    registry,
		repository:  repository,
		tagOrDigest: tagOrDigest,
	}, nil
}

func (i *ImageSource) GetManifest() ([]byte, string, error) {

	if i.source == nil {
		return nil, "",
			fmt.Errorf(
				"cannot get manifest without tag or digest",
			)
	}

	return i.source.GetManifest(
		i.ctx,
		nil,
	)
}

func (i *ImageSource) GetBlobInfos(
	manifestObjSlice ...manifest.Manifest,
) ([]types.BlobInfo, error) {

	if i.source == nil {
		return nil,
			fmt.Errorf(
				"cannot get blobs without tag or digest",
			)
	}

	var result []types.BlobInfo

	for _, obj := range manifestObjSlice {

		for _, layer := range obj.LayerInfos() {
			result = append(
				result,
				layer.BlobInfo,
			)
		}

		config := obj.ConfigInfo()

		if config.Digest != "" {
			result = append(
				result,
				config,
			)
		}
	}

	return result, nil
}

func (i *ImageSource) GetABlob(
	blobInfo types.BlobInfo,
) (io.ReadCloser, int64, error) {

	if i.source == nil {
		return nil, 0,
			fmt.Errorf(
				"cannot get blob without tag or digest",
			)
	}

	return i.source.GetBlob(
		i.ctx,
		types.BlobInfo{
			Digest: blobInfo.Digest,
			URLs:   blobInfo.URLs,
			Size:   -1,
		},
		NoCache,
	)
}

func (i *ImageSource) Close() error {

	if i == nil {
		return nil
	}

	if i.source == nil {
		return nil
	}

	return i.source.Close()
}

func (i *ImageSource) GetRegistry() string {
	return i.registry
}

func (i *ImageSource) GetRepository() string {
	return i.repository
}

func (i *ImageSource) GetTagOrDigest() string {
	return i.tagOrDigest
}

func (i *ImageSource) String() string {

	return i.registry +
		"/" +
		i.repository +
		utils.AttachConnectorToTagOrDigest(
			i.tagOrDigest,
		)
}

func (i *ImageSource) GetSourceRepoTags() ([]string, error) {

	return docker.GetRepositoryTags(
		i.ctx,
		i.sysctx,
		i.ref,
	)
}
