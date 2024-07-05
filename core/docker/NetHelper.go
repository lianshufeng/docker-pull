package docker_tools

import (
	"fmt"
	"io"
	"net/http"
	url2 "net/url"
	"strings"
)

func MakeUrl(hostName string, uri string, mirror string) string {
	var _hostName string
	if mirror == "" {
		_hostName = hostName
	} else {
		_hostName = mirror
	}
	return fmt.Sprintf("https://%s/%s", _hostName, uri)
}

func Net_Get(hostName string, uri string, headers map[string]string, mirror string, proxy string) []byte {
	url := MakeUrl(hostName, uri, mirror)
	fmt.Println("access : ", url)
	req, err := http.NewRequest("GET", url, strings.NewReader(""))

	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 创建 http 客户端
	client := MakeHttpClient(proxy)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}
	defer resp.Body.Close()
	// 读取响应体内容
	body, err := io.ReadAll(resp.Body)
	return body
}

func MakeHttpClient(proxy string) *http.Client {
	//使用代理
	transport := &http.Transport{}
	if proxy != "" {
		proxyUrl, _ := url2.Parse(proxy)
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}
	return &http.Client{
		Transport: transport,
	}
}
