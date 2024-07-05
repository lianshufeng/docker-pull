package file

import (
	"io"
	"os"
)

func Copy(SourceFile string, DestFile string) {
	// 打开源文件
	sourceFile, err := os.Open(SourceFile)
	if err != nil {
		panic(err)
	}
	defer sourceFile.Close()

	// 创建或打开目标文件用于写入
	targetFile, err := os.Create(DestFile)
	if err != nil {
		panic(err)
	}
	defer targetFile.Close()

	// 使用io.Copy()进行文件内容的复制
	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		panic(err)
	}
}
