package docker_tools

import (
	"fmt"
	"io"
	"net/http"
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

	client := &http.Client{}
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
