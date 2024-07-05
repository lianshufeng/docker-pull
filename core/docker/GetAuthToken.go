package docker_tools

import (
	"encoding/json"
	"fmt"
)

type AuthToken struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
	ExpiresAt   int    `json:"expires_in"`
	IssuedAt    string `json:"issued_at"`
}

func GetAuthToken(imageName string, accept string, mirror string, proxy string) AuthToken {
	header := map[string]string{
		"Accept": accept,
	}
	body, _ := Net_Get("auth.docker.io", fmt.Sprintf("token?service=registry.docker.io&scope=repository:%s:pull", imageName), header, mirror, proxy)
	var result AuthToken
	json.Unmarshal(body, &result)
	return result
}
