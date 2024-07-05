package docker_tools

import (
	"fmt"
)

func GetConfigManifests(imageName string, blobs string, auth_token string, mirror string, proxy string) []byte {
	header := map[string]string{
		"Authorization": "Bearer " + auth_token,
		"Accept":        "application/vnd.docker.distribution.manifest.v2+json",
	}
	body := Net_Get("registry.hub.docker.com", fmt.Sprintf("v2/%s/blobs/%s", imageName, blobs), header, mirror, proxy)
	return body

}
