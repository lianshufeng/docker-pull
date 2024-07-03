package docker_tools

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"io"
	"os"
)

func ImagePull(imageName string, pullOptions image.PullOptions) bool {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	pullResponse, err := cli.ImagePull(ctx, imageName, pullOptions)

	// 开始拉取镜像
	if err != nil {
		fmt.Printf("Error pulling image: %v\n", err)
		return false
	}
	var reader = io.TeeReader(pullResponse, os.Stdout) // 将读取到的数据同时输出到标准输出
	var bufReader = bufio.NewReader(reader)
	// 打印拉取进度
	for {
		line, err := bufReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading line:", err)
			break
		}
		fmt.Println("pull:", string(line))
	}
	return true
}

// ContainerList 获取容器列表 /**
func ContainerList(listOptions container.ListOptions) []types.Container {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	containers, err := cli.ContainerList(ctx, listOptions)
	if err != nil {
		panic(err)
	}
	return containers
}

func ContainerRemove(containerID string, options container.RemoveOptions) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	err = cli.ContainerRemove(ctx, containerID, options)
	return err
}

func ContainerCreate(config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	return cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, platform, containerName)
}

func ContainerStart(containerID string, options container.StartOptions) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	err = cli.ContainerStart(ctx, containerID, options)
	return err
}

func ContainerStop(containerID string, options container.StopOptions) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	err = cli.ContainerStop(ctx, containerID, options)
	return err
}

func ContainerPause(containerID string, signal string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	err = cli.ContainerPause(ctx, containerID)
	return err
}

func ContainerKill(containerID string, signal string) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	err = cli.ContainerKill(ctx, containerID, signal)
	return err
}

func ContainerStats(containerID string, stream bool) (container.StatsResponseReader, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	return cli.ContainerStats(ctx, containerID, stream)
}

func ImageList(options image.ListOptions) ([]image.Summary, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	return cli.ImageList(ctx, options)
}
