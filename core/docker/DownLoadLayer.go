package docker_tools

import (
	"fmt"
	"path"
	"time"
)

func DownLoadLayer(imageName string, ublob string, mirror string, proxy string, outputFile string) {
	//判断是否下载完成
	if CompleteFile(outputFile) {
		fmt.Println("Download Complete : " + path.Base(outputFile))
		return
	}
	authToken := GetAuthToken(imageName, "application/vnd.docker.distribution.manifest.v2+json", mirror, proxy)

	url := MakeUrl("registry.hub.docker.com", fmt.Sprintf("v2/%s/blobs/%s", imageName, ublob), mirror)
	headers := map[string]string{
		"Authorization": "Bearer " + authToken.Token,
	}
	err := DownLoad(url, headers, outputFile, proxy)

	//如果有错误，延迟后重新下载
	if err != nil {
		fmt.Println("error,  sleep 5 s")
		time.Sleep(time.Second * 5)
		DownLoadLayer(imageName, ublob, mirror, proxy, outputFile)
	}

}
