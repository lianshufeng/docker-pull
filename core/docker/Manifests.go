package docker_tools

import (
	"encoding/json"
	"fmt"
)

type Manifest_v1 struct {
	Manifests []struct {
		Annotations map[string]string `json:"annotations"`
		Digest      string            `json:"digest"`
		MediaType   string            `json:"mediaType"`
		Platform    struct {
			Architecture string `json:"architecture"`
			Os           string `json:"os"`
		} `json:"platform"`
		Size int `json:"size"`
	} `json:"manifests"`
	MediaType     string `json:"mediaType"`
	SchemaVersion int    `json:"schemaVersion"`
}

type Manifests struct {
	SchemaVersion int           `json:"schemaVersion"`
	Name          string        `json:"name"`
	Tag           string        `json:"tag"`
	Architecture  string        `json:"architecture"`
	FsLayers      []FsLayer     `json:"fsLayers"`
	History       []HistoryItem `json:"history"`
	Signatures    []Signature   `json:"signatures"`
}

type FsLayer struct {
	BlobSum string `json:"blobSum"`
}

type HistoryItem struct {
	V1Compatibility string `json:"v1Compatibility"`
}

type Signature struct {
	Header    json.RawMessage `json:"header"`
	Signature string          `json:"signature"`
	Protected string          `json:"protected"`
}

// Manifest v2
type Manifest struct {
	SchemaVersion int     `json:"schemaVersion"`
	MediaType     string  `json:"mediaType"`
	Config        Config  `json:"config"`
	Layers        []Layer `json:"layers"`
}

type Config struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}

type Layer struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}

func GetManifests(imageName string, tag string, os string, architecture string, auth_token string, mirror string, proxy string) Manifest {
	header := map[string]string{
		"Authorization": "Bearer " + auth_token,
		"Accept":        "application/vnd.docker.distribution.manifest.v2+json",
	}
	body := Net_Get("registry.hub.docker.com", fmt.Sprintf("v2/%s/manifests/%s", imageName, tag), header, mirror, proxy)
	fmt.Println(string(body))

	var obj map[string]interface{}
	json.Unmarshal(body, &obj)
	if obj["schemaVersion"] == nil {
		fmt.Println("schemaVersion is nil")
		return Manifest{}
	}

	mediaType := obj["mediaType"].(string)

	if mediaType == "application/vnd.docker.distribution.manifest.v2+json" {
		var manifest Manifest
		json.Unmarshal(body, &manifest)
		return manifest
	} else if mediaType == "application/vnd.oci.image.index.v1+json" {
		var manifest_v1 Manifest_v1
		json.Unmarshal(body, &manifest_v1)
		for i := range manifest_v1.Manifests {
			platform := manifest_v1.Manifests[i].Platform
			if platform.Os == os && platform.Architecture == architecture {
				return Manifest{
					SchemaVersion: manifest_v1.SchemaVersion,
					MediaType:     manifest_v1.MediaType,
					Config: Config{
						MediaType: manifest_v1.Manifests[i].MediaType,
						Size:      manifest_v1.Manifests[i].Size,
						Digest:    manifest_v1.Manifests[i].Digest,
					},
				}
			}
		}
	}

	return Manifest{}

	//if schemaVersion == 1 {
	//	var manifests Manifests
	//	json.Unmarshal(body, &manifests)
	//	return manifests
	//} else if schemaVersion == 2 {
	//	var manifests Manifests2
	//	json.Unmarshal(body, &manifests)
	//
	//	for i := range manifests.Manifests {
	//		platform := manifests.Manifests[i].Platform
	//		if platform.Os == os && platform.Architecture == architecture {
	//			return Manifests{
	//				SchemaVersion: manifests.SchemaVersion,
	//				Name:          imageName,
	//				Tag:           tag,
	//				Architecture:  platform.Architecture,
	//				FsLayers: []FsLayer{
	//					{
	//						BlobSum: manifests.Manifests[i].Digest,
	//					},
	//				},
	//			}
	//		}
	//	}
	//}
	//return Manifests{}
}

//var data map[string]interface{}
//err := json.Unmarshal(body, &data)
//fmt.Println(err)

//兼容两种返回结果
//if data["manifests"] != nil {
//for _, v := range m {
//	// 使用 reflect.ValueOf 获取反射值
//	value := reflect.ValueOf(v)
//	// 检查值的类型
//	if value.Kind() == reflect.String {
//		fmt.Println("Value is string:", value.String())
//	} else if value.Kind() == reflect.Int || value.Kind() == reflect.Float64 || value.Kind() == reflect.Bool {
//		// 可以添加更多类型的判断
//		fmt.Println("Value is not a string, but it's a:", value.Kind())
//	}
//}
//}
