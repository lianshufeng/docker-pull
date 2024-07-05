package file

import (
	"os"
)

// IsExist 文件是否存在/**
func IsExist(DestFile string) bool {
	//判断文件是否存在
	file1, err := os.Stat(DestFile)
	if err == nil {
		if file1.IsDir() {
			return true
		}
		if file1.Size() > 0 {
			return true
		}
	}
	return false
}

// Remove 删除文件
func Remove(DestFile string) error {
	//判断文件是否存在
	file1, err := os.Stat(DestFile)
	if err == nil {
		if file1.IsDir() {
			return os.RemoveAll(DestFile)
		}
		return os.Remove(DestFile)
	}
	return nil
}
