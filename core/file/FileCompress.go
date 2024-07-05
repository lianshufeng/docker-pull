package file

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func UnGzip(SourceFile string, DestFile string) {
	// 打开一个 Gzip 压缩文件
	gzipFile, err := os.Open(SourceFile)
	if err != nil {
		panic(err)
	}
	defer gzipFile.Close()

	gzipReader, _ := gzip.NewReader(io.Reader(gzipFile))
	defer gzipReader.Close()

	//写出文件
	outputFile, err := os.Create(DestFile)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, gzipReader)
	if err != nil {
		panic(err)
	}

}

func TarDir(dirPath string, writer *tar.Writer) error {
	// 遍历目录下的所有文件和子目录
	return filepath.Walk(dirPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 确保我们不处理根目录自身（因为它已经被作为参数传入）
		if filePath == dirPath {
			return nil
		}

		// 创建tar.Header
		header, err := tar.FileInfoHeader(info, filePath)
		if err != nil {
			return err
		}

		// 调整Header的名字，确保它是相对于基础目录的
		header.Name = strings.ReplaceAll(filePath[len(dirPath)+1:], "\\", "/")

		// 写入Header
		if err := writer.WriteHeader(header); err != nil {
			return err
		}

		// 如果是目录，则不需要写入内容
		if info.IsDir() {
			return nil
		}

		// 打开文件并复制内容到tar包中
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}
