package docker_tools

import (
	"encoding/json"
	"fmt"
	"github.com/lianshufeng/docker-pull/file"
	"github.com/panjf2000/ants"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const (
	// 断点续传下载正常响应码
	NormalRespCode = "206 Partial Content"
	// 断点续传下载，偏移量值超文件最大值长度错误响应码
	ErrRespCode = "416 Requested Range Not Satisfiable"
	// 默认的文件块
	DefaultChunkSize = 1024 * 1024 * 200
	// 默认下载线程数
	DefaultDownloadThreadCount = 3
)

// FileChunk 文件块
type FileChunk struct {
	FileChunkName string `json:"file"`
	RangeStart    int64  `json:"rangeStart"`
	RangeEnd      int64  `json:"rangeEnd"`
	IsDone        bool   `json:"isDone"`
}

// FileInfo 文件信息
type FileDownloadInfo struct {
	FileSize   int64       `json:"size"`
	FileChunks []FileChunk `json:"chunks"`
	FileUrl    string      `json:"url"`
}

func MakeProcessFileName(DestFile string) string {
	return DestFile + ".dl"
}

// CompleteFile 是否完成文件下载
func CompleteFile(DestFile string) bool {
	// 进度文件
	processFileName := MakeProcessFileName(DestFile)
	// 目标文件存在且进度文件不存在,则下载完成
	return file.IsExist(DestFile) && !file.IsExist(processFileName)
}

// DownLoad 下载文件
func DownLoad(url string, headers map[string]string, DestFile string, proxy string, BuffByte int64) error {
	// 进度文件
	processFileName := MakeProcessFileName(DestFile)

	// 目标文件存在且进度文件不存在,则下载完成
	if file.IsExist(DestFile) && !file.IsExist(processFileName) {
		//文件下载完成
		return error(nil)
	}

	//如果文件不存在，进度文件存在则删除进度文件重新下载
	if !file.IsExist(DestFile) && file.IsExist(processFileName) {
		os.Remove(processFileName)
	}
	IsDownLoadDone := error(nil)

	// 线程锁
	var mutex sync.Mutex

	// 构建文件下载进度
	fileDownloadInfo := makeProcessFile(processFileName)
	if fileDownloadInfo.FileSize <= 0 {
		// 初始化下载的配置文件(完成分块)
		initProcessFile(&fileDownloadInfo, url, headers, proxy, DestFile)

		// 删除存在的下载文件
		if file.IsExist(DestFile) {
			os.Remove(DestFile)
		}

		// 保存配置
		saveProcessFile(&mutex, &fileDownloadInfo, processFileName)
	}

	//开始下载, 线程池数量等于 缓存块占用内存*2
	//downloadPool, _ := ants.NewPool((int(BuffByte/DefaultChunkSize) + 1) * 2)
	downloadPool, _ := ants.NewPool(DefaultDownloadThreadCount)
	defer downloadPool.Release()

	// 下载的文件,更新方式打开句柄
	fileHandle, _ := os.OpenFile(DestFile, os.O_RDWR|os.O_CREATE, os.ModePerm)
	defer fileHandle.Close()

	var checkTaskWG sync.WaitGroup
	checkTaskWG.Add(len(fileDownloadInfo.FileChunks))

	for i := range fileDownloadInfo.FileChunks {
		fileChunk := &fileDownloadInfo.FileChunks[i]
		_ = downloadPool.Submit(func() {
			defer checkTaskWG.Done()
			// 文件块存在则不需要下载
			chunkFile := DestFile + fileDownloadInfo.FileChunks[i].FileChunkName
			if file.IsExist(chunkFile) {
				fmt.Println("skip file chunk : ", chunkFile)
				return
			}

			// 进行分块下载
			DownLoadChunkFile(DestFile, fileChunk, url, headers, proxy)
		})
	}
	checkTaskWG.Wait()

	// 合并文件
	for i := range fileDownloadInfo.FileChunks {
		chunkFile := DestFile + fileDownloadInfo.FileChunks[i].FileChunkName
		readFile, _ := os.OpenFile(chunkFile, os.O_RDWR, os.ModePerm)
		defer readFile.Close()
		io.Copy(fileHandle, readFile)
		readFile.Close()
		file.Remove(chunkFile)
	}

	// 保存配置(删除文件)
	os.Remove(processFileName)

	fileHandle.Close()

	return IsDownLoadDone
}

func DownLoadChunkFile(DestFile string, fileChunks *FileChunk, url string, headers map[string]string, proxy string) bool {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("error : ", err)
		return false
	}

	// 设置 Range 实现断点续传
	range01 := fmt.Sprintf("bytes=%d-%d", fileChunks.RangeStart, fileChunks.RangeEnd)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Range", range01)

	fmt.Println("download : ", url, fmt.Sprintf("%d/%d", fileChunks.RangeStart, fileChunks.RangeEnd))

	// 创建 http 客户端
	client := MakeHttpClient(proxy)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("request error:", err)
		return false
	}
	defer resp.Body.Close()

	// 判断下载状态
	if resp.Status != NormalRespCode {
		if resp.Status == ErrRespCode {
			fmt.Println(fmt.Sprintf("file download ErrRespCode ：[%s]", resp.Status))
			return true
		} else {
			//下载错误
			fmt.Println(fmt.Sprintf("file download ：error[%s]", resp.Status))
			return false
		}
	}

	// 写出到磁盘上
	outputFile, _ := os.OpenFile(DestFile+fileChunks.FileChunkName, os.O_RDWR|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	defer outputFile.Close()
	io.Copy(outputFile, resp.Body)

	return true
}
func writeSpaceFile(destFile string, fileDownloadInfo *FileDownloadInfo) {
	file, _ := os.OpenFile(destFile, os.O_CREATE|os.O_RDWR, os.ModePerm)
	defer file.Close()
	// 写入 fileDownloadInfo.FileChunks 这么大的空间到destfile里, 使用缓存，提高写入效率
	for _, chunk := range fileDownloadInfo.FileChunks {
		file.Seek(chunk.RangeStart, io.SeekStart)
		file.Write(make([]byte, chunk.RangeEnd))
	}
}

func initProcessFile(fileDownloadInfo *FileDownloadInfo, url string, headers map[string]string, proxy string, DestFile string) {
	fileSize := getContentRange(url, headers, proxy)
	fileDownloadInfo.FileSize = fileSize
	fileDownloadInfo.FileUrl = url

	if fileDownloadInfo.FileChunks == nil {
		fileDownloadInfo.FileChunks = []FileChunk{}
	}

	count := int(fileSize / DefaultChunkSize)
	for i := 0; i < count; i++ {
		fileDownloadInfo.FileChunks = append(fileDownloadInfo.FileChunks, FileChunk{
			RangeStart:    int64(i) * DefaultChunkSize,
			RangeEnd:      int64(i+1)*DefaultChunkSize - 1,
			IsDone:        false,
			FileChunkName: "_" + strconv.FormatInt(int64(i), 10),
		})
	}

	if (fileSize % DefaultChunkSize) > 0 {
		fileDownloadInfo.FileChunks = append(fileDownloadInfo.FileChunks, FileChunk{
			RangeStart:    int64(count) * DefaultChunkSize,
			RangeEnd:      fileSize,
			IsDone:        false,
			FileChunkName: "_" + strconv.FormatInt(int64(count), 10),
		})
	}
}

// 保存进度文件
func saveProcessFile(mutex *sync.Mutex, fileDownloadInfo *FileDownloadInfo, processFileName string) {
	mutex.Lock()
	defer mutex.Unlock()
	process_file, _ := os.OpenFile(processFileName, os.O_CREATE|os.O_RDWR, os.ModePerm)
	defer process_file.Close()
	bin, _ := json.Marshal(fileDownloadInfo)
	process_file.Write(bin)
}

func updateFileBytes(mutex *sync.Mutex, fileHandle *os.File, fileChunk *FileChunk, bin []byte) {
	mutex.Lock()
	defer mutex.Unlock()

	fileHandle.Seek(fileChunk.RangeStart, io.SeekStart)
	fileHandle.Write(bin)
	fileChunk.IsDone = true
}

// 创建进度文件
func makeProcessFile(process_file_name string) FileDownloadInfo {
	//创建进度文件
	process_file, _ := os.OpenFile(process_file_name, os.O_CREATE|os.O_RDWR, os.ModePerm)
	defer process_file.Close()

	bs := make([]byte, 10240, 10240)
	n1, _ := process_file.Read(bs)
	fileInfo := FileDownloadInfo{}
	if n1 > 0 {
		json.Unmarshal(bs, &fileInfo)
	}

	return fileInfo
}

func getContentRange(url string, headers map[string]string, proxy string) int64 {
	ret := int64(-1)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return ret
	}

	// 设置 Range 实现断点续传
	range01 := fmt.Sprintf("bytes=%d-%d", 0, 100)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Range", range01)

	// 创建 http 客户端
	client := MakeHttpClient(proxy)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("request error:", err)
		return ret
	}
	defer resp.Body.Close()
	ContentRange := resp.Header.Get("Content-Range")
	totalRange := strings.Split(ContentRange, "/")
	if len(totalRange) >= 2 {
		ret, err = strconv.ParseInt(totalRange[1], 10, 64)
	}
	return ret
}

func DownLoad_bakup(url string, headers map[string]string, DestFile string, proxy string, BuffByte int64) error {
	// 进度文件
	process_file_name := MakeProcessFileName(DestFile)

	// 目标文件存在且进度文件不存在,则下载完成
	if file.IsExist(DestFile) && !file.IsExist(process_file_name) {
		//文件下载完成
		return nil
	}

	//如果文件不存在，进度文件存在则删除进度文件重新下载
	if !file.IsExist(DestFile) && file.IsExist(process_file_name) {
		os.Remove(process_file_name)
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
