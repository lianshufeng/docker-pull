package docker_tools

import (
	"encoding/json"
	"fmt"
	"os"
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

type TarManifest struct {
	SchemaVersion int     `json:"schemaVersion"`
	MediaType     string  `json:"mediaType"`
	Config        Config  `json:"config"`
	Layers        []Layer `json:"layers"`
}

func GetManifests(imageName string, tag string, accept string, platform_os string, architecture string, auth_token string, mirror string, proxy string) Manifest {
	header := map[string]string{
		"Authorization": "Bearer " + auth_token,
		"Accept":        accept,
	}
	body := Net_Get("registry-1.docker.io", fmt.Sprintf("v2/%s/manifests/%s", imageName, tag), header, mirror, proxy)
	//fmt.Println(string(body))

	var obj map[string]interface{}
	json.Unmarshal(body, &obj)
	if obj["schemaVersion"] == nil {
		//异常
		panic("request manifests error..")
	}

	mediaType := obj["mediaType"].(string)

	if mediaType == "application/vnd.docker.distribution.manifest.v2+json" {
		var manifest Manifest
		json.Unmarshal(body, &manifest)
		return manifest
	} else if mediaType == "application/vnd.oci.image.manifest.v1+json" {
		var manifest Manifest
		json.Unmarshal(body, &manifest)
		return manifest
	} else if mediaType == "application/vnd.oci.image.index.v1+json" { //oci 选择镜像
		var manifest_v1 Manifest_v1
		json.Unmarshal(body, &manifest_v1)
		for i := range manifest_v1.Manifests {
			platform := manifest_v1.Manifests[i].Platform
			if platform.Os == platform_os && platform.Architecture == architecture {
				_accept := manifest_v1.Manifests[i].MediaType
				_auth_token := GetAuthToken(imageName, _accept, mirror, proxy).Token
				return GetManifests(imageName, manifest_v1.Manifests[i].Digest, _accept, platform_os, architecture, _auth_token, mirror, proxy)
			}
		}
	}
	err := fmt.Errorf("not found : ", imageName+":"+tag+"@"+platform_os+"/"+architecture)
	fmt.Fprintf(os.Stderr, "错误: %v\n", err)
	os.Exit(1)
	return Manifest{}
}
