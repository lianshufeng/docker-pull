package docker_tools

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type AuthToken struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
	ExpiresAt   int    `json:"expires_in"`
	IssuedAt    string `json:"issued_at"`
}

func GetAuthToken(imageName string, accept string, mirror string, proxy string) AuthToken {

	// 需要先构建权限服务器的URL
	v2Header := map[string]string{
		"Accept": accept,
	}
	_, _, v2RespHeader := Net_Get("registry-1.docker.io", "v2/", v2Header, mirror, proxy)
	authenticate := v2RespHeader.Get("WWW-Authenticate")
	// 取出WWW-Authenticate头
	authenticateItems := map[string]string{}
	if authenticate != "" {
		authenticate = strings.Replace(authenticate, "Bearer ", "", -1)
		//用，拆分字符串
		fields := strings.Split(authenticate, ",")
		// 遍历字段
		for _, field := range fields {
			// 用=拆分字段
			parts := strings.Split(field, "=")
			if len(parts) == 2 {
				// 去除双引号
				value := strings.Trim(parts[1], "\"")
				authenticateItems[parts[0]] = value
			}
		}
	}

	u, _ := url.Parse(authenticateItems["realm"])
	authRealm := u.Host

	uri := fmt.Sprintf("%s?scope=repository:%s:pull&service=%s", u.Path, url.QueryEscape(imageName), url.QueryEscape(authenticateItems["service"]))
	// 循环判断第一个是否/，如果是的话则去掉
	for strings.HasPrefix(uri, "/") {
		uri = strings.TrimPrefix(uri, "/")
	}
	uri = strings.Replace(uri, ":", "%3A", -1)

	// 进行url编码 urlencode
	header := map[string]string{
		//"Accept": accept,
	}
	body, _, _ := Net_Get("auth.docker.io", uri, header, authRealm, proxy)
	var result AuthToken
	err := json.Unmarshal(body, &result)
	if err == nil {
		return result
	}
	return AuthToken{}
}
