package main

import (
	"archive/tar"
	arg_tools "docker-pull/arg"
	docker_tools "docker-pull/docker"
	"docker-pull/file"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/panjf2000/ants"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultAccept = "application/vnd.docker.distribution.manifest.v2+json"
)

type DownLoadLayerTask struct {
	Layer        docker_tools.Layer
	Fake_layerid string
	Output       string
	ImageName    string
	Tag          string
	Args         arg_tools.Args
}

func InitEnv(ConfigENV map[string]string) {
	for name, _ := range ConfigENV {
		os.Setenv(name, ConfigENV[name])
	}
}

//func hashSHA256(parentID string, ublob string) string {
//	input := parentID + "\n" + ublob + "\n"
//	hasher := sha256.New()
//	hasher.Write([]byte(input))
//	return hex.EncodeToString(hasher.Sum(nil))
//}

func MakeLayerId(ublob string) string {
	return ublob[7:]
}

func DownLoadLayer(task DownLoadLayerTask) {
	docker_tools.DownLoadLayer(task.ImageName, task.Layer.Digest, task.Args.Mirror, task.Args.Proxy, task.Args.BuffByte, task.Output)
}

/**
 * 下载并解压
 */
func DownAndUnGzip(task DownLoadLayerTask, cacheDirectory string, blobsDirectory string) {

	//开始下载层
	DownLoadLayer(task)

	//层的目录
	layerFile := cacheDirectory + "/" + task.Fake_layerid + ".gzip.tar"
	fmt.Println("UnGzip : ", task.Fake_layerid)
	//解压转换
	UnSuccess, _ := file.UnGzip(layerFile, blobsDirectory+"/"+task.Fake_layerid+".tar")
	//解压失败，重新下载
	if UnSuccess == false {
		fmt.Println("Re DownLoadLayer : ", task.Fake_layerid)
		if file.IsExist(task.Output) {
			fmt.Println("Delete : ", task.Output)
			os.Remove(task.Output)
		}
		DownAndUnGzip(task, cacheDirectory, blobsDirectory)
	}

}

func PullImage(imageName string, digest string, tag string, args arg_tools.Args) {

	//获取token
	authToken := docker_tools.GetAuthToken(imageName, DefaultAccept, args.Mirror, args.Proxy)

	//获取层的清单
	manifest := docker_tools.GetManifests(imageName, digest, tag, args.Os, args.Architecture, args.Variant, authToken.Token, args.Mirror, args.Proxy)

	//创建缓存目录
	cacheName := strings.ReplaceAll(imageName, "/", "_") + "@" + tag
	cacheDirectory := args.Cache + "/" + cacheName
	config_json_name := manifest.Config.Digest[7:] + ".json"
	config_json_file := cacheDirectory + "/" + config_json_name
	os.MkdirAll(cacheDirectory, os.ModeDir)

	//写配置文件
	config_ret := docker_tools.GetConfigManifests(imageName, manifest.Config.Digest, authToken.Token, args.Mirror, args.Proxy)
	os.WriteFile(config_json_file, config_ret, os.ModePerm)

	//创建工程目录
	projectDirectory := cacheDirectory + "/_tmp_" + strconv.FormatInt(time.Now().Unix(), 10)
	blobsDirectory := projectDirectory + "/blobs"
	os.MkdirAll(blobsDirectory, os.ModeDir)

	// 创建DockerTarManifest
	dockerTarManifest := docker_tools.DockerTarManifest{
		{
			Config:   config_json_name,
			RepoTags: []string{imageName + ":" + tag},
			Layers:   []string{},
		},
	}

	//线程池
	pool, _ := ants.NewPool(args.ThreadCount)
	defer pool.Release()
	var wg sync.WaitGroup
	//var parentID string
	for i := range manifest.Layers {
		//取出层
		layer := manifest.Layers[i]
		fake_layerid := MakeLayerId(layer.Digest)
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
		//添加任务
		wg.Add(1)
		_ = pool.Submit(func() {
			defer wg.Done()
			//记录最后一层的id
			//LastLayerId = fake_layerid
			fmt.Println("download layer:", layer.Digest)
			DownAndUnGzip(task, cacheDirectory, blobsDirectory)
		})
	}
	// 等待所有协程执行完成
	wg.Wait()

	//添加层
	for i := range manifest.Layers {
		layer := manifest.Layers[i]
		dockerTarManifest[0].Layers = append(dockerTarManifest[0].Layers, "blobs/"+MakeLayerId(layer.Digest)+".tar")
	}

	// 循环解压到tar
	//var LastLayerId string
	//for i := range manifest.Layers {
	//	//取出层
	//	layer := manifest.Layers[i]
	//	fake_layerid := MakeLayerId(parentID, layer.Digest)
	//	//记录最后一层的id
	//	LastLayerId = fake_layerid
	//	//层的目录
	//	layerFile := cacheDirectory + "/" + fake_layerid + ".gzip.tar"
	//	fmt.Println("UnGzip : ", fake_layerid)
	//	//解压转换
	//	file.UnGzip(layerFile, blobsDirectory+"/"+fake_layerid+".tar")
	//
	//	//添加层
	//	dockerTarManifest[0].Layers = append(dockerTarManifest[0].Layers, "blobs/"+fake_layerid+".tar")
	//}

	//拷贝config
	file.Copy(config_json_file, projectDirectory+"/"+config_json_name)

	//写出 manifest.json
	dockerTarManifestJson, _ := json.Marshal(dockerTarManifest)
	os.WriteFile(projectDirectory+"/manifest.json", dockerTarManifestJson, os.ModePerm)

	//构建 repositories
	repositories := map[string]map[string]string{
		imageName: {
			tag: manifest.Layers[len(manifest.Layers)-1].Digest,
		},
	}
	repositoriesJson, _ := json.Marshal(repositories)
	os.WriteFile(projectDirectory+"/repositories", repositoriesJson, os.ModePerm)

	// 压缩目录为tar格式
	pwdPath, _ := os.Getwd()
	outputFilePath := filepath.Clean(pwdPath + "/" + cacheName + ".tar")
	fmt.Println("compress file:", filepath.Base(outputFilePath))
	outputFile, err := os.Create(outputFilePath)
	defer outputFile.Close()
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	tarWriter := tar.NewWriter(outputFile)
	defer tarWriter.Close()
	file.TarDir(projectDirectory, tarWriter)

	//删除临时目录
	os.RemoveAll(projectDirectory)

	//载入
	if args.IsLoad {
		fmt.Println("load image:", filepath.Base(outputFilePath))
		docker_tools.ImageLoad(outputFilePath)
	}

	fmt.Println("image save :", outputFilePath)

}

func main() {
	//读取命令行参数
	args := arg_tools.LoadArgs()
	args_json, _ := json.Marshal(args)
	fmt.Println("task : ", string(args_json))

	//环境变量
	InitEnv(map[string]string{
		"DOCKER_API_VERSION": args.DockerApiVersion,
	})

	if args.ImageName == "" {
		flag.PrintDefaults()
		return
	}

	//开始下载镜像
	PullImage(args.ImageName, args.Digest, args.Tag, args)

}
