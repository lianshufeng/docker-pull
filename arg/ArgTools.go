package arg_tools

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Args struct {
	Proxy            string
	Architecture     string
	Variant          string
	Os               string
	Cache            string
	CleanCache       bool
	CleanImage       bool
	ThreadCount      int
	BuffByte         int64
	IsLoad           bool
	DockerApiVersion string
}

type Image struct {
	InputName string
	ImageName string
	Tag       string
	Digest    string
	Mirror    string
}

func ToArgs(defaultMirror string, InputName string) Image {
	//提取tag
	var tag string
	var digest string
	imageName := InputName
	mirror := defaultMirror
	if strings.Contains(InputName, "@") {
		digest = InputName[strings.Index(InputName, "@")+1:]
		tag = "latest"
		imageName = InputName[:strings.Index(InputName, "@")]
	} else if strings.Contains(InputName, ":") {
		at := strings.Index(InputName, ":")
		tag = InputName[at+1:]
		imageName = InputName[0:at]
	} else {
		tag = "latest"
	}

	imageNames := strings.Split(InputName, "/")
	if len(imageNames) == 1 {
		imageName = "library/" + InputName
	} else if len(imageNames) == 2 {
		//nothing
	} else if len(imageNames) > 2 {
		//取出除了最后2级以外的所有目录,兼容自定义域名
		imageName = strings.Join(imageNames[len(imageNames)-2:len(imageNames)], "/")
		mirror = strings.Join(imageNames[0:len(imageNames)-2], "/")
	}

	return Image{
		InputName: InputName,
		ImageName: imageName,
		Tag:       tag,
		Digest:    digest,
		Mirror:    mirror,
	}
}

func LoadArgs() (Args, []Image) {

	var mirror string
	flag.StringVar(&mirror, "m", "", "mirror url , docker.jpy.wang")

	var proxy string
	flag.StringVar(&proxy, "proxy", "", "proxy server , http://127.0.0.1:1080")

	//架构
	var platform_architecture string
	flag.StringVar(&platform_architecture, "architecture", runtime.GOARCH, "platform.architecture")

	//版本
	var platform_variant string
	flag.StringVar(&platform_variant, "variant", "", "platform.variant")

	//系统功能
	var platform_os string
	flag.StringVar(&platform_os, "os", "linux", "platform.os")

	var cache string
	flag.StringVar(&cache, "cache", "_cache", "cache directory")

	var ThreadCount int
	flag.IntVar(&ThreadCount, "thread", 10, "thead number")

	var BuffByte int64
	flag.Int64Var(&BuffByte, "buffByte", 1024*1024*10-1, "download buffByte")

	var IsLoad bool
	flag.BoolVar(&IsLoad, "load", true, "check and load the local image")

	var cleanCache bool
	flag.BoolVar(&cleanCache, "cleanCache", false, "clean cache")

	var cleanImage bool
	flag.BoolVar(&cleanImage, "cleanImage", false, "clean image")

	var DockerApiVersion string
	flag.StringVar(&DockerApiVersion, "docker_api_version", "1.30", "DOCKER_API_VERSION")

	flag.Parse()

	if len(os.Args) <= 1 || strings.TrimSpace(flag.Args()[0]) == "" {
		fmt.Println("docker-pull [-config] <imageName>")
		fmt.Println("eg :", "docker-pull -proxy http://127.0.0.1:1080 -thread 5 nginx ")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// cache 判断首字母是否为 /
	if cache[0] != '/' && !strings.ContainsRune(cache, ':') {
		dir, _ := os.Getwd()
		cache = dir + "/" + cache
	}
	cache = filepath.Clean(cache)

	defaultArgs := Args{
		Proxy:            proxy,
		Architecture:     platform_architecture,
		Variant:          platform_variant,
		Os:               platform_os,
		Cache:            cache,
		CleanCache:       cleanCache,
		CleanImage:       cleanImage,
		ThreadCount:      ThreadCount,
		BuffByte:         BuffByte,
		IsLoad:           IsLoad,
		DockerApiVersion: DockerApiVersion,
	}

	//定义 Args 类型的数组
	var images []Image
	for i := range flag.Args() {
		inputName := flag.Args()[i]
		//数组去重
		exists := false
		for _, v := range images {
			if v.InputName == inputName {
				exists = true
				break
			}
		}
		if !exists {
			//数组里添加元素
			images = append(images, ToArgs(mirror, inputName))
		}
	}
	return defaultArgs, images

	//imageName := os.Args[len(os.Args)-1:][0]

	////提取tag
	//var tag string
	//var Digest string
	//if strings.Contains(imageName, "@") {
	//	Digest = imageName[strings.Index(imageName, "@")+1:]
	//	tag = "latest"
	//	imageName = imageName[:strings.Index(imageName, "@")]
	//} else if strings.Contains(imageName, ":") {
	//	at := strings.Index(imageName, ":")
	//	tag = imageName[at+1:]
	//	imageName = imageName[0:at]
	//} else {
	//	tag = "latest"
	//}
	//
	////if len(imageNames) == 1 && !IsContainsMirror {
	//imageNames := strings.Split(imageName, "/")
	//if len(imageNames) == 1 {
	//	imageName = "library/" + imageName
	//} else if len(imageNames) == 2 {
	//	//nothing
	//} else if len(imageNames) > 2 {
	//	//取出除了最后2级以外的所有目录,兼容自定义域名
	//	imageName = strings.Join(imageNames[len(imageNames)-2:len(imageNames)], "/")
	//	mirror = strings.Join(imageNames[0:len(imageNames)-2], "/")
	//}
	//
	//return Args{
	//	ImageName:        imageName,
	//	Tag:              tag,
	//	Digest:           Digest,
	//	Mirror:           mirror,
	//	Proxy:            proxy,
	//	Architecture:     platform_architecture,
	//	Variant:          platform_variant,
	//	Os:               platform_os,
	//	Cache:            cache,
	//	CleanCache:       cleanCache,
	//	CleanImage:       cleanImage,
	//	ThreadCount:      ThreadCount,
	//	BuffByte:         BuffByte,
	//	IsLoad:           IsLoad,
	//	DockerApiVersion: DockerApiVersion,
	//}
}
