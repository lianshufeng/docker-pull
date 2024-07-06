package docker_tools

import (
	"docker-pull/file"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// 断点续传下载正常响应码
	NormalRespCode = "206 Partial Content"
	// 断点续传下载，偏移量值超文件最大值长度错误响应码
	ErrRespCode = "416 Requested Range Not Satisfiable"
)

func MakeProcessFileName(DestFile string) string {
	return DestFile + ".dl"
}

// CompleteFile 是否完成文件下载
func CompleteFile(DestFile string) bool {
	// 进度文件
	process_file_name := MakeProcessFileName(DestFile)
	// 目标文件存在且进度文件不存在,则下载完成
	return file.IsExist(DestFile) && !file.IsExist(process_file_name)
}

func DownLoad(url string, headers map[string]string, DestFile string, proxy string, BuffByte int64) error {
	// 进度文件
	process_file_name := MakeProcessFileName(DestFile)

	// 目标文件存在且进度文件不存在,则下载完成
	if file.IsExist(DestFile) && !file.IsExist(process_file_name) {
		//文件下载完成
		return nil
	}

	IsDownLoadDone := error(nil)

	//创建进度文件
	process_file, err := os.OpenFile(process_file_name, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer process_file.Close()

	//创建下载文件
	output_file, err := os.OpenFile(DestFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer output_file.Close()

	// 读取临时文件中的数据，根据 seek
	process_file.Seek(0, io.SeekStart)
	bs := make([]byte, 100, 100)
	n1, err := process_file.Read(bs)
	countStr := string(bs[:n1])
	countArr := strings.Split(countStr, "/")
	var count, total int64 = 0, 0
	var bfb = "0%"
	var countf, totalf float64 = 0.0001, 10000
	if len(countArr) >= 3 {
		count, _ = strconv.ParseInt(countArr[0], 10, 64)
		total, _ = strconv.ParseInt(countArr[1], 10, 64)
		bfb = countArr[2]
	}

	//fmt.Println(fmt.Sprintf("download：%d, total：%d, process：%s", count, total, bfb))

	for {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		// 设置 Range 实现断点续传
		range01 := fmt.Sprintf("bytes=%d-%d", count, count+BuffByte)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		req.Header.Set("Range", range01)

		// 创建 http 客户端
		client := MakeHttpClient(proxy)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("request error:", err)
			return err
		}
		defer resp.Body.Close()

		ContentRange := resp.Header.Get("Content-Range")
		totalRange := strings.Split(ContentRange, "/")
		if len(totalRange) >= 2 {
			total, _ = strconv.ParseInt(totalRange[1], 10, 64)
		}

		if resp.Status != NormalRespCode {
			if resp.Status == ErrRespCode {
				fmt.Println("download complete :", path.Base(DestFile))
				process_file.Close()
				file.Remove(process_file_name)
				IsDownLoadDone = nil
			} else {
				fmt.Println(fmt.Sprintf("file download ：error[%s]", resp.Status))
				IsDownLoadDone = fmt.Errorf("file download ：error[%s]", resp.Status)
			}
			break
		}

		data, err := io.ReadAll(resp.Body)
		n3, _ := output_file.Write(data)
		count += int64(n3)
		countf = float64(count)
		totalf = float64(total)
		bfb = fmt.Sprintf("%.2f", countf/totalf*100) + "%"
		fmt.Println("file:", filepath.Base(DestFile), "size:", strconv.FormatInt(count, 10)+"/"+strconv.FormatInt(total, 10), "process:", bfb)

		// 记录已下载的字节数到临时文件
		process_file.Seek(0, io.SeekStart)
		process_file.WriteString(fmt.Sprintf("%d/%d/%s", count, total, bfb))

	}
	return IsDownLoadDone
}
