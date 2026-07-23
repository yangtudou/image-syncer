package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/opencontainers/go-digest"

	"github.com/containers/image/v5/docker/reference"
)

const (
	DockerHubURL = "docker.io"
)

type RepoURL struct {
	ref reference.Reference

	registry    string
	repo        string
	tagOrDigest string
}

// GenerateRepoURLs creates RepoURL list.
// 支持：
// image:tag
// image@digest
// image:tag1,tag2
// image:/regex/
// image(no tag) -> externalTags
func GenerateRepoURLs(
	url string,
	externalTagsOrDigest func(
		registry,
		repository string,
	) ([]string, error),
) ([]*RepoURL, error) {

	var result []*RepoURL

	var tagsOrDigest []string
	var urlWithoutTagOrDigest string

	ref, err := reference.ParseNormalizedNamed(url)

	// digest
	if err == nil {

		if canonicalRef, ok := ref.(reference.Canonical); ok {

			tagsOrDigest = append(
				tagsOrDigest,
				canonicalRef.Digest().String(),
			)

			urlWithoutTagOrDigest = canonicalRef.Name()

		} else if taggedRef, ok := ref.(reference.NamedTagged); ok {

			tagsOrDigest = append(
				tagsOrDigest,
				taggedRef.Tag(),
			)

			urlWithoutTagOrDigest = taggedRef.Name()

		} else {

			// 没有 tag
			registry, repo :=
				getRegistryAndRepositoryFromURLWithoutTagOrDigest(url)

			tags, err :=
				externalTagsOrDigest(
					registry,
					repo,
				)

			if err != nil {
				return nil,
					fmt.Errorf(
						"failed to get tags: %v",
						err,
					)
			}

			urlWithoutTagOrDigest = url

			tagsOrDigest = append(
				tagsOrDigest,
				tags...,
			)
		}

	} else {

		// regex tag
		if strings.Contains(url, ":/") {

			slice :=
				strings.SplitN(
					url,
					":/",
					2,
				)

			if len(slice) != 2 ||
				!strings.HasSuffix(slice[1], "/") {

				return nil,
					fmt.Errorf(
						"invalid tag regex format",
					)
			}

			_, err :=
				reference.ParseNormalizedNamed(
					slice[0],
				)

			if err != nil {
				return nil,
					err
			}

			urlWithoutTagOrDigest = slice[0]

			regexStr :=
				strings.TrimSuffix(
					slice[1],
					"/",
				)

			regex, err :=
				regexp.Compile(
					regexStr,
				)

			if err != nil {
				return nil,
					fmt.Errorf(
						"invalid regex: %v",
						err,
					)
			}

			registry, repo :=
				getRegistryAndRepositoryFromURLWithoutTagOrDigest(
					urlWithoutTagOrDigest,
				)

			allTags, err :=
				externalTagsOrDigest(
					registry,
					repo,
				)

			if err != nil {
				return nil,
					err
			}

			for _, tag := range allTags {

				if regex.MatchString(tag) {

					tagsOrDigest =
						append(
							tagsOrDigest,
							tag,
						)
				}
			}

		} else {

			// 多 tag
			items :=
				strings.Split(
					url,
					",",
				)

			if len(items) == 0 {
				return nil,
					fmt.Errorf(
						"invalid image url",
					)
			}

			first, err :=
				reference.ParseNormalizedNamed(
					items[0],
				)

			if err != nil {
				return nil,
					err
			}

			tagged, ok :=
				first.(reference.NamedTagged)

			if !ok {
				return nil,
					fmt.Errorf(
						"invalid tag image: %s",
						items[0],
					)
			}

			urlWithoutTagOrDigest =
				tagged.Name()

			tagsOrDigest =
				append(
					tagsOrDigest,
					tagged.Tag(),
				)

			tagsOrDigest =
				append(
					tagsOrDigest,
					items[1:]...,
				)
		}
	}

	registry, repo :=
		getRegistryAndRepositoryFromURLWithoutTagOrDigest(
			urlWithoutTagOrDigest,
		)

	for _, tag := range tagsOrDigest {

		newURL :=
			registry +
				"/" +
				repo +
				AttachConnectorToTagOrDigest(tag)

		ref, err :=
			reference.ParseNormalizedNamed(
				newURL,
			)

		if err != nil {
			return nil,
				fmt.Errorf(
					"invalid canonical url: %s",
					newURL,
				)
		}

		result =
			append(
				result,
				&RepoURL{
					ref:         ref,
					registry:    registry,
					repo:        repo,
					tagOrDigest: tag,
				},
			)
	}

	return result, nil
}

func (r *RepoURL) String() string {
	return r.ref.String()
}

func (r *RepoURL) GetRegistry() string {
	return r.registry
}

func (r *RepoURL) GetRepo() string {
	return r.repo
}

func (r *RepoURL) GetTagOrDigest() string {
	return r.tagOrDigest
}

func (r *RepoURL) GetRepoWithTagOrDigest() string {

	if r.tagOrDigest == "" {
		return r.repo
	}

	return r.repo +
		AttachConnectorToTagOrDigest(
			r.tagOrDigest,
		)
}

func (r *RepoURL) HasDigest() bool {

	_, ok :=
		r.ref.(reference.Canonical)

	return ok
}

func (r *RepoURL) GetURLWithoutTagOrDigest() string {

	return r.registry +
		"/" +
		r.repo
}

func AttachConnectorToTagOrDigest(
	tagOrDigest string,
) string {

	if tagOrDigest == "" {
		return ""
	}

	d := digest.Digest(tagOrDigest)

	if err := d.Validate(); err != nil {

		return ":" + tagOrDigest
	}

	return "@" + tagOrDigest
}

func getRegistryAndRepositoryFromURLWithoutTagOrDigest(
	url string,
) (
	registry string,
	repo string,
) {

	items :=
		strings.SplitN(
			url,
			"/",
			2,
		)

	if len(items) == 1 {

		registry = DockerHubURL
		repo = items[0]

	} else {

		registry = items[0]
		repo = items[1]
	}

	return
}
