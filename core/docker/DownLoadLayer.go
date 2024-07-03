package docker_tools

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	// 断点续传下载正常响应码
	NormalRespCode = "206 Partial Content"
	// 断点续传下载，偏移量值超文件最大值长度错误响应码
	ErrRespCode = "416 Requested Range Not Satisfiable"
	BuffByte    = 1024*1024 - 1
)

func downloadFile(url string, headers map[string]string, DestFile string) error {
	file2, err := os.OpenFile(DestFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file2.Close()

	file3, err := os.OpenFile(DestFile+".dl", os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer file3.Close()

	// 读取临时文件中的数据，根据 seek
	file3.Seek(0, io.SeekStart)
	bs := make([]byte, 100, 100)
	n1, err := file3.Read(bs)
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

	fmt.Println(fmt.Sprintf("开始下载，已下载：%d, 总共：%d, 进度：%s", count, total, bfb))

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

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("请求失败:", err)
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
				fmt.Println("文件下载完毕。")
				file3.Close()
			} else {
				fmt.Println(fmt.Sprintf("文件传输异常报错：err[%s]", resp.Status))
			}
			break
		}

		data, err := io.ReadAll(resp.Body)
		n3, _ := file2.Write(data)
		count += int64(n3)
		countf = float64(count)
		totalf = float64(total)
		bfb = fmt.Sprintf("%.2f", countf/totalf*100) + "%"
		fmt.Println("本次读取了：", n3, "总共已下载:", count, "文件总大小：", total, "下载进度：", bfb)

		// 记录已下载的字节数到临时文件
		file3.Seek(0, io.SeekStart)
		file3.WriteString(fmt.Sprintf("%d/%d/%s", count, total, bfb))

	}

	return nil
}

// downloadFile 下载远程文件到本地文件
//func _downloadFile(url string, headers map[string]string, outputFile string) error {
//	// 设置 HTTP 头部信息
//	req, err := http.NewRequest("GET", url, strings.NewReader(""))
//
//	// 设置请求头
//	for key, value := range headers {
//		req.Header.Set(key, value)
//	}
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		fmt.Println("Error sending request:", err)
//	}
//	defer resp.Body.Close()
//
//	// Check if the file exists and can be resumed.
//	if fileInfo, err := os.Stat(outputFile); !os.IsNotExist(err) && err == nil {
//		if fileInfo.Size() > 0 {
//			// Try to get the offset from the Range header.
//			rangeHeader := resp.Header.Get("Content-Range")
//			if rangeHeader != "" {
//				offsetStr := strings.Split(rangeHeader, "/")[0]
//				offset, err := strconv.ParseInt(offsetStr, 10, 64)
//				if err == nil {
//					// Resume the download at the given offset.
//					httpRange := fmt.Sprintf("bytes=%d-", offset)
//					resp, err = http.Get(fmt.Sprintf("%s%s", url, httpRange))
//					if err != nil {
//						return err
//					}
//					defer resp.Body.Close()
//				}
//			}
//		}
//	}
//
//	output, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0666)
//	if err != nil {
//		return err
//	}
//	defer output.Close()
//
//	_, err = io.CopyN(output, resp.Body, -1)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

func DownLoadLayer(authToken AuthToken, imageName string, ublob string, mirror string, outputFile string) {
	url := MakeUrl("registry.hub.docker.com", fmt.Sprintf("v2/%s/blobs/%s", imageName, ublob), mirror)
	headers := map[string]string{
		"Authorization": "Bearer " + authToken.Token,
	}
	downloadFile(url, headers, outputFile)
}
