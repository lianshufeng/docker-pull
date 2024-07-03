package main

import (
	"crypto/sha256"
	arg_tools "docker-pull/arg"
	docker_tools "docker-pull/docker"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

type DownLoadLayerTask struct {
	Layer        docker_tools.Layer
	Fake_layerid string
	Output       string
	ImageName    string
	Tag          string
	Args         arg_tools.Args
}

func InitEnv() {
	os.Setenv("DOCKER_API_VERSION", "1.30")
}

func hashSHA256(parentID string, ublob string) string {
	input := parentID + "\n" + ublob + "\n"
	hasher := sha256.New()
	hasher.Write([]byte(input))
	return hex.EncodeToString(hasher.Sum(nil))
}

func DownLoadLayer(task DownLoadLayerTask) {
	authToken := docker_tools.GetAuthToken(task.ImageName, "application/vnd.docker.distribution.manifest.v2+json", task.Args.Mirror, task.Args.Proxy)
	fmt.Println("auth : " + authToken.Token)
	docker_tools.DownLoadLayer(authToken, task.ImageName, task.Layer.Digest, task.Args.Mirror, task.Output)

	//'https://{}/v2/{}/blobs/{}'.format(registry, repository, ublob)

	//'https://docker.jpy.wang/v2/lianshufeng/springboot/blobs/sha256:a1d0c75327776413fa0db9ed3adcdbadedc95a662eb1d360dad82bb913f8a1d1'

	//client := resty.New()
	//resp, err := client.R().SetResult(&bytes.Buffer{}).Get(url)
	//if err != nil {
	//	return err
	//}

}

func PullImage(imageName string, tag string, args arg_tools.Args) {
	//获取token
	authToken := docker_tools.GetAuthToken(imageName, "application/vnd.docker.distribution.manifest.v2+json", args.Mirror, args.Proxy)
	fmt.Println("AuthToken:", authToken.Token)

	manifest := docker_tools.GetManifests(imageName, tag, args.Os, args.Architecture, authToken.Token, args.Mirror, args.Proxy)
	fmt.Println("manifests:", manifest)

	//创建缓存目录
	cacheDirectory := args.Cache + "/" + strings.ReplaceAll(imageName, "/", "_") + "@" + tag
	config_json_file := cacheDirectory + "/" + manifest.Config.Digest[7:] + ".json"
	os.MkdirAll(cacheDirectory, os.ModeDir)

	//写配置文件
	manifest_json, _ := json.Marshal(manifest)
	os.WriteFile(config_json_file, manifest_json, os.ModePerm)

	var parentID string
	for i := range manifest.Layers {
		//取出层
		layer := manifest.Layers[i]
		fake_layerid := hashSHA256(parentID, layer.Digest)
		//层的目录
		layerFile := cacheDirectory + "/" + fake_layerid + ".gzip.tar"

		//下载任务
		task := DownLoadLayerTask{
			Output:       layerFile,
			ImageName:    imageName,
			Tag:          tag,
			Fake_layerid: fake_layerid,
			Layer:        layer,
			Args:         args,
		}
		DownLoadLayer(task)
		fmt.Println("layer:", layer.Digest)
	}

	//images, _ := docker_tools.ImageList(image.ListOptions{})
	//for _, image := range images {
	//	fmt.Println("image:", image.RepoTags)
	//}
	//fmt.Println("--")
}

func main() {
	//环境变量
	InitEnv()

	//读取命令行参数
	args := arg_tools.LoadArgs()
	ret, _ := json.Marshal(args)
	fmt.Println("task : " + string(ret))
	if args.ImageName == "" {
		flag.PrintDefaults()
		return
	}

	//开始下载镜像
	PullImage(args.ImageName, args.Tag, args)

}
