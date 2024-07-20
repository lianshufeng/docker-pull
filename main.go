package main

import (
	"flag"
	arg_tools "github.com/lianshufeng/docker-pull/arg"
	"github.com/lianshufeng/docker-pull/core"
	"os"
)

func InitEnv(ConfigENV map[string]string) {
	for name, _ := range ConfigENV {
		os.Setenv(name, ConfigENV[name])
	}
}

func main() {
	//读取命令行参数
	arg, images := arg_tools.LoadArgs()

	if len(images) == 0 {
		flag.PrintDefaults()
		return
	}

	//环境变量
	InitEnv(map[string]string{
		"DOCKER_API_VERSION": arg.DockerApiVersion,
	})

	//开始下载镜像
	core.PullImage(images, arg)

}
