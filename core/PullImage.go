package core

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	arg_tools "github.com/lianshufeng/docker-pull/arg"
	"github.com/lianshufeng/docker-pull/docker"
	"github.com/lianshufeng/docker-pull/file"
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
	Layer       docker_tools.Layer
	FakeLayerId string
	Output      string
	Image       arg_tools.Image
	Args        arg_tools.Args
}

type Image struct {
	Image     arg_tools.Image
	AuthToken docker_tools.AuthToken
	Manifest  docker_tools.Manifest
}

type ImageProject struct {
	CacheDirectory     string
	ProjectName        string
	ProjectDirectory   string
	BlobsDirectory     string
	ConfigJsonFile     string
	DockerTarManifests docker_tools.DockerTarManifest
	Image              arg_tools.Image
	Args               arg_tools.Args
	Manifest           docker_tools.Manifest
	AuthToken          docker_tools.AuthToken
}

func MakeLayerId(ublob string) string {
	return ublob[7:]
}

func DownLoadLayer(task DownLoadLayerTask) {
	docker_tools.DownLoadLayer(task.Image.ImageName, task.Layer.Digest, task.Image.Mirror, task.Args.Proxy, task.Args.BuffByte, task.Output)
}

// CheckImageProject /*
func CheckImageProject(downloadPool *ants.Pool, imageProject ImageProject) bool {
	count := 0
	for _, layer := range imageProject.Manifest.Layers {
		//层的存储文件
		FakeLayerId := MakeLayerId(filepath.Base(layer.Digest))
		layerStoreName := FakeLayerId + ".gzip.tar"
		layerStoreFile := filepath.Clean(imageProject.CacheDirectory + "/" + layerStoreName)

		destFile := filepath.Clean(imageProject.BlobsDirectory + "/" + FakeLayerId + ".tar")

		//已解压完成
		if file.IsExist(destFile) {
			count++
			continue
		}
		//如果缓存文件存在，则解压，如果解压失败则重新添加下载任务
		if docker_tools.CompleteFile(layerStoreFile) {
			//解压转换
			UnSuccess, _ := file.UnGzip(layerStoreFile, destFile)
			//解压失败，重新下载
			if UnSuccess == false {
				fmt.Println("Re DownLoadLayer : ", layerStoreFile)
				if file.IsExist(layerStoreFile) {
					fmt.Println("Delete : ", layerStoreFile)
					os.Remove(layerStoreFile)
				}
				addDownloadTasks(downloadPool, DownLoadLayerTask{
					Args:        imageProject.Args,
					Image:       imageProject.Image,
					FakeLayerId: FakeLayerId,
					Layer:       layer,
					Output:      layerStoreFile,
				})
				continue
			}
			count++
		}
	}

	//如果磁盘数量不等于层的数量则继续等待磁盘上的文件下载完成
	if count != len(imageProject.Manifest.Layers) {
		return false
	}

	//写出 manifest.json
	dockerTarManifestJson, _ := json.Marshal(imageProject.DockerTarManifests)
	projectDirectory := imageProject.ProjectDirectory
	os.WriteFile(projectDirectory+"/manifest.json", dockerTarManifestJson, os.ModePerm)

	//构建 repositories
	repositories := map[string]map[string]string{
		"imageName": {
			"tag": imageProject.Manifest.Layers[len(imageProject.Manifest.Layers)-1].Digest,
		},
	}
	repositoriesJson, _ := json.Marshal(repositories)
	os.WriteFile(projectDirectory+"/repositories", repositoriesJson, os.ModePerm)

	// 压缩目录为tar格式
	pwdPath, _ := os.Getwd()
	outputFilePath := filepath.Clean(pwdPath + "/" + imageProject.ProjectName + ".tar")
	fmt.Println("compress file:", filepath.Base(outputFilePath))
	outputFile, err := os.Create(outputFilePath)
	defer outputFile.Close()
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return true
	}
	tarWriter := tar.NewWriter(outputFile)
	defer tarWriter.Close()
	file.TarDir(projectDirectory, tarWriter)

	//删除项目的临时目录
	os.RemoveAll(projectDirectory)

	//载入
	if imageProject.Args.IsLoad {
		fmt.Println("load image:", filepath.Base(outputFilePath))
		docker_tools.ImageLoad(outputFilePath)
	}

	// 清除镜像
	if imageProject.Args.CleanImage {
		fmt.Println("image delete :", outputFilePath)
		outputFile.Close()
		os.Remove(outputFilePath)
	} else {
		fmt.Println("image save :", outputFilePath)
	}

	return true
}

// 当前下载的任务队列
var currentDownloadTasks sync.Map

// 所有的下载任务
var totalDownloadTasks sync.Map

// CheckDownLoadTask /**
func CheckDownLoadTask(pool *ants.Pool) {
	totalDownloadTasks.Range(func(key, value interface{}) bool {
		fakeLayerId := key.(string)
		downLoadLayerTask := value.(DownLoadLayerTask)
		//下载完成
		_, ok := currentDownloadTasks.Load(fakeLayerId)
		if !ok {
			//文件未下载完成则重新添加到下载任务队列
			if !docker_tools.CompleteFile(downLoadLayerTask.Output) {
				fmt.Println("add download task : ", downLoadLayerTask.Output)
				addDownloadTasks(pool, downLoadLayerTask)
			}
		}
		return true
	})
}

func addDownloadTasks(pool *ants.Pool, task DownLoadLayerTask) {
	fakeLayerId := task.FakeLayerId
	//记录所有的下载记录
	totalDownloadTasks.Store(fakeLayerId, task)
	// 判断是否已经存在下载任务
	_, ok := currentDownloadTasks.Load(fakeLayerId)
	if ok {
		fmt.Println("Download Task Exist :", fakeLayerId)
		return
	}
	currentDownloadTasks.Store(fakeLayerId, true)
	_ = pool.Submit(func() {
		DownLoadLayer(task)
		currentDownloadTasks.Delete(fakeLayerId)
	})
}

func PullImage(images []arg_tools.Image, args arg_tools.Args) {

	//定一个数组用于装在需要更新的镜像
	var updateImages []Image
	for _, image := range images {
		//获取token
		authToken := docker_tools.GetAuthToken(image.ImageName, DefaultAccept, image.Mirror, args.Proxy)

		//获取层的清单
		manifest := docker_tools.GetManifests(image.ImageName, image.Digest, image.Tag, args.Os, args.Architecture, args.Variant, authToken.Token, image.Mirror, args.Proxy)

		//查询本地镜像
		if args.IsLoad {
			localImage := docker_tools.GetImage(manifest.Config.Digest)
			if localImage.ID != "" {
				err := docker_tools.ImageTag(localImage.ID, image.ImageName, image.Tag)
				if err == nil {
					fmt.Println(fmt.Sprintf("%s -> %s:%s", localImage.ID, image.ImageName, image.Tag))
					continue
				}
			}
		}

		//追加到数组里
		updateImages = append(updateImages, Image{
			Image:     image,
			AuthToken: authToken,
			Manifest:  manifest,
		})
	}
	if len(updateImages) == 0 {
		return
	}

	//创建缓存目录
	cacheName := "store"
	cacheDirectory := filepath.Clean(args.Cache + "/" + cacheName)
	os.MkdirAll(cacheDirectory, os.ModeDir)

	// 构建需要下载的镜像工程
	updateImageProjects := sync.Map{}
	for _, updateImage := range updateImages {
		image := updateImage.Image
		manifest := updateImage.Manifest
		authToken := updateImage.AuthToken

		//创建工程目录
		projectName := strings.ReplaceAll(image.ImageName, "/", "_") + "@" + image.Tag
		projectDirectory := filepath.Clean(cacheDirectory + "/" + projectName + "_" + strconv.FormatInt(time.Now().Unix(), 10))
		blobsDirectory := filepath.Clean(projectDirectory + "/blobs")
		os.MkdirAll(blobsDirectory, os.ModeDir)

		config_json_name := updateImage.Manifest.Config.Digest[7:] + ".json"
		config_json_file := filepath.Clean(projectDirectory + "/" + config_json_name)

		//写配置文件
		config_ret := docker_tools.GetConfigManifests(image.ImageName, manifest.Config.Digest, authToken.Token, image.Mirror, args.Proxy)
		os.WriteFile(config_json_file, config_ret, os.ModePerm)

		//镜像对应的工程
		imageProject := ImageProject{
			CacheDirectory:   cacheDirectory,
			ProjectName:      projectName,
			ProjectDirectory: projectDirectory,
			BlobsDirectory:   blobsDirectory,
			ConfigJsonFile:   config_json_file,
			DockerTarManifests: docker_tools.DockerTarManifest{
				{
					Config:   config_json_name,
					RepoTags: []string{image.ImageName + ":" + image.Tag},
					Layers:   []string{},
				},
			},
			Image:     image,
			Args:      args,
			Manifest:  manifest,
			AuthToken: authToken,
		}

		// 计算layers
		for i := range manifest.Layers {
			fakeLayerId := MakeLayerId(manifest.Layers[i].Digest)
			imageProject.DockerTarManifests[0].Layers = append(imageProject.DockerTarManifests[0].Layers, "blobs/"+fakeLayerId+".tar")
		}

		//添加到线程安全对象
		updateImageProjects.Store(blobsDirectory, imageProject)
	}

	//需要下载的层 (过滤重复的)
	updateLayers := sync.Map{}
	for _, image := range updateImages {
		for _, layer := range image.Manifest.Layers {
			updateLayers.Store(layer.Digest, image)
		}
	}
	//开始下载
	downloadPool, _ := ants.NewPool(args.ThreadCount)
	defer downloadPool.Release()
	currentDownloadTasks = sync.Map{}
	totalDownloadTasks = sync.Map{}
	updateLayers.Range(func(key, value interface{}) bool {
		layerId := key.(string)
		fakeLayerId := MakeLayerId(layerId)
		layerFile := filepath.Clean(cacheDirectory + "/" + fakeLayerId + ".gzip.tar")
		downLoadTask := DownLoadLayerTask{
			Layer:       docker_tools.Layer{Digest: layerId},
			FakeLayerId: MakeLayerId(layerId),
			Output:      layerFile,
			Image:       value.(Image).Image,
			Args:        args,
		}
		addDownloadTasks(downloadPool, downLoadTask)
		return true
	})

	//启动下载任务检查定时器
	checkDownLoadTaskTicker := time.NewTicker(10 * time.Second)
	defer checkDownLoadTaskTicker.Stop()
	go func() {
		for range checkDownLoadTaskTicker.C {
			CheckDownLoadTask(downloadPool)
		}
	}()

	//启动线程进行检查任务是否下载完成
	checkTaskCompletePool, _ := ants.NewPool(len(updateImages))
	defer checkTaskCompletePool.Release()
	var checkTaskWG sync.WaitGroup
	updateImageProjects.Range(func(key, value interface{}) bool {
		imageProject := value.(ImageProject)
		checkTaskWG.Add(1)
		_ = checkTaskCompletePool.Submit(func() {
			defer checkTaskWG.Done()
			for !CheckImageProject(downloadPool, imageProject) {
				time.Sleep(1 * time.Second)
			}
		})
		return true
	})
	//等待线程结束
	checkTaskWG.Wait()

	//清除缓存
	if args.CleanCache {
		fmt.Println("cache delete :", cacheDirectory)
		os.RemoveAll(cacheDirectory)
	}

	fmt.Println("Done ...")

}
