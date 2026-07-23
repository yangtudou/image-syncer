package sync

import (
	"fmt"
	"io"
	"strings"

	"github.com/containers/image/v5/manifest"
	"github.com/opencontainers/go-digest"
	specsv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/tidwall/gjson"
)

type ManifestInfo struct {
	Obj manifest.Manifest

	Digest *digest.Digest

	Bytes []byte
}

func GenerateManifestObj(
	manifestBytes []byte,
	manifestType string,
	osFilterList []string,
	archFilterList []string,
	i *ImageSource,
	parent *manifest.Schema2List,
) (interface{}, []byte, []*ManifestInfo, error) {

	switch manifestType {

	case manifest.DockerV2Schema2MediaType:

		obj, err := manifest.Schema2FromManifest(manifestBytes)
		if err != nil {
			return nil, nil, nil, err
		}

		if parent == nil && obj.ConfigInfo().Digest != "" {

			blob, _, err := i.GetABlob(obj.ConfigInfo())
			if err != nil {
				return nil, nil, nil, err
			}

			defer blob.Close()

			data, err := io.ReadAll(blob)
			if err != nil {
				return nil, nil, nil, err
			}

			result := gjson.GetManyBytes(
				data,
				"architecture",
				"os",
			)

			if !platformValidate(
				osFilterList,
				archFilterList,
				&manifest.Schema2PlatformSpec{
					Architecture: result[0].String(),
					OS:           result[1].String(),
				},
			) {
				return nil, nil, nil, nil
			}
		}

		return obj, manifestBytes, nil, nil

	case manifest.DockerV2Schema1MediaType,
		manifest.DockerV2Schema1SignedMediaType:

		obj, err := manifest.Schema1FromManifest(manifestBytes)
		if err != nil {
			return nil, nil, nil, err
		}

		return obj, manifestBytes, nil, nil

	case specsv1.MediaTypeImageManifest:

		obj, err := manifest.OCI1FromManifest(manifestBytes)
		if err != nil {
			return nil, nil, nil, err
		}

		return obj, manifestBytes, nil, nil

	case manifest.DockerV2ListMediaType:

		list, err := manifest.Schema2ListFromManifest(manifestBytes)
		if err != nil {
			return nil, nil, nil, err
		}

		return generateSchema2List(
			list,
			osFilterList,
			archFilterList,
			i,
		)

	case specsv1.MediaTypeImageIndex:

		index, err := manifest.OCI1IndexFromManifest(manifestBytes)
		if err != nil {
			return nil, nil, nil, err
		}

		var children []*ManifestInfo
		var descriptors []specsv1.Descriptor

		for idx, descriptor := range index.Manifests {

			if descriptor.Platform != nil {

				if !platformValidate(
					osFilterList,
					archFilterList,
					&manifest.Schema2PlatformSpec{
						Architecture: descriptor.Platform.Architecture,
						OS:           descriptor.Platform.OS,
					},
				) {
					continue
				}
			}

			descriptors = append(
				descriptors,
				descriptor,
			)

			data, typ, err := i.source.GetManifest(
				i.ctx,
				&descriptor.Digest,
			)

			if err != nil {
				return nil, nil, nil, err
			}

			child, _, _, err := GenerateManifestObj(
				data,
				typ,
				osFilterList,
				archFilterList,
				i,
				nil,
			)

			if err != nil {
				return nil, nil, nil, err
			}

			if child == nil {
				continue
			}

			obj, ok := child.(manifest.Manifest)

			if !ok {
				return nil, nil, nil,
					fmt.Errorf(
						"invalid child manifest type %T",
						child,
					)
			}

			d := index.Manifests[idx].Digest

			children = append(
				children,
				&ManifestInfo{
					Obj:    obj,
					Digest: &d,
					Bytes:  data,
				},
			)
		}

		if len(descriptors) == 0 {
			return nil, nil, nil, nil
		}

		index.Manifests = descriptors

		data, err := index.Serialize()
		if err != nil {
			return nil, nil, nil, err
		}

		return index, data, children, nil
	}

	return nil, nil, nil,
		fmt.Errorf(
			"unsupported manifest type: %s",
			manifestType,
		)
}

func generateSchema2List(
	list *manifest.Schema2List,
	osFilterList []string,
	archFilterList []string,
	i *ImageSource,
) (interface{}, []byte, []*ManifestInfo, error) {

	var children []*ManifestInfo
	var descriptors []manifest.Schema2ManifestDescriptor

	for idx, descriptor := range list.Manifests {

		if !platformValidate(
			osFilterList,
			archFilterList,
			&descriptor.Platform,
		) {
			continue
		}

		descriptors = append(
			descriptors,
			descriptor,
		)

		data, typ, err := i.source.GetManifest(
			i.ctx,
			&descriptor.Digest,
		)

		if err != nil {
			return nil, nil, nil, err
		}

		child, _, _, err := GenerateManifestObj(
			data,
			typ,
			osFilterList,
			archFilterList,
			i,
			list,
		)

		if err != nil {
			return nil, nil, nil, err
		}

		if child == nil {
			continue
		}

		obj, ok := child.(manifest.Manifest)

		if !ok {
			return nil, nil, nil,
				fmt.Errorf(
					"invalid child manifest type %T",
					child,
				)
		}

		d := list.Manifests[idx].Digest

		children = append(
			children,
			&ManifestInfo{
				Obj:    obj,
				Digest: &d,
				Bytes:  data,
			},
		)
	}

	if len(descriptors) == 0 {
		return nil, nil, nil, nil
	}

	list.Manifests = descriptors

	data, err := list.Serialize()
	if err != nil {
		return nil, nil, nil, err
	}

	return list, data, children, nil
}

func colonMatch(
	pat string,
	first string,
	second string,
) bool {

	if strings.Index(
		pat,
		first,
	) != 0 {
		return false
	}

	return len(first) == len(pat) ||
		(pat[len(first)] == ':' &&
			pat[len(first)+1:] == second)
}

func platformValidate(
	osFilterList []string,
	archFilterList []string,
	platform *manifest.Schema2PlatformSpec,
) bool {

	if platform == nil {
		return true
	}

	osMatched := true
	archMatched := true

	if len(osFilterList) != 0 &&
		platform.OS != "" {

		osMatched = false

		for _, o := range osFilterList {

			if colonMatch(
				o,
				platform.OS,
				platform.OSVersion,
			) {
				osMatched = true
			}
		}
	}

	if len(archFilterList) != 0 &&
		platform.Architecture != "" {

		archMatched = false

		for _, a := range archFilterList {

			if colonMatch(
				a,
				platform.Architecture,
				platform.Variant,
			) {
				archMatched = true
			}
		}
	}

	return osMatched && archMatched
}
