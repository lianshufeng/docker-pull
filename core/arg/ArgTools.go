package arg_tools

import (
	"flag"
	"os"
	"strings"
)

type Args struct {
	ImageName    string
	Tag          string
	Mirror       string
	Proxy        string
	Architecture string
	Os           string
	Cache        string
}

func LoadArgs() Args {
	//定义一个字符串变量
	var mirror string
	flag.StringVar(&mirror, "m", "", "mirror url , docker.jpy.wang")

	var proxy string
	flag.StringVar(&proxy, "proxy", "", "proxy server")

	var imageName string
	flag.StringVar(&imageName, "i", "lianshufeng/springboot", "image name,not empty")

	var architecture string
	flag.StringVar(&architecture, "architecture", "amd64", "platform.architecture")

	var platform_os string
	flag.StringVar(&platform_os, "os", "linux", "platform.os")

	var cache string
	flag.StringVar(&cache, "cache", "_cache", "cache directory")

	flag.Parse()

	// cache 判断首字母是否为 /
	if cache[0] != '/' && !strings.ContainsRune(cache, ':') {
		dir, _ := os.Getwd()
		cache = dir + "/" + cache
	}

	//提取tag
	var tag string
	at := strings.Index(imageName, ":")
	if at > -1 {
		tag = imageName[at+1:]
		imageName = imageName[0:at]
	} else {
		tag = "latest"
		imageName = imageName
	}

	imageNames := strings.Split(imageName, "/")
	if len(imageNames) == 1 {
		imageName = "library/" + imageName
	} else if len(imageNames) == 2 {
		//nothing
	} else if len(imageNames) > 2 {
		//取出除了最后2级以外的所有目录
		imageName = strings.Join(imageNames[len(imageNames)-2:len(imageNames)], "/")
		mirror = strings.Join(imageNames[0:len(imageNames)-2], "/")
	}

	//增加官方库library
	//if strings.Index(imageName, "/") == -1 {
	//	imageName = "library/" + imageName
	//}

	return Args{
		ImageName:    imageName,
		Tag:          tag,
		Mirror:       mirror,
		Proxy:        proxy,
		Architecture: architecture,
		Os:           platform_os,
		Cache:        cache,
	}
}
