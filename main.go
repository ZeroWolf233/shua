package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

func worker(id int, wg *sync.WaitGroup, url string, interval time.Duration, stopChan <-chan struct{}) {
	defer wg.Done()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 零间隔处理模式
	if interval == 0 {
		fmt.Printf("工作进程 %d 启动（无间隔）\n", id)
		for {
			select {
			case <-stopChan:
				fmt.Printf("工作进程 %d 停止\n", id)
				return
			default:
				start := time.Now()
				resp, err := client.Get(url)
				if handleResponse(id, resp, err, start) {
					return
				}
			}
		}
	}

	// 正常间隔模式
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	fmt.Printf("工作进程 %d 启动（间隔: %v）\n", id, interval)

	for {
		select {
		case <-stopChan:
			fmt.Printf("工作进程 %d 停止\n", id)
			return
		case <-ticker.C:
			start := time.Now()
			resp, err := client.Get(url)
			if handleResponse(id, resp, err, start) {
				return
			}
		}
	}
}

// 公共响应处理函数
func handleResponse(id int, resp *http.Response, err error, start time.Time) bool {
	if err != nil {
		fmt.Printf("[工作进程 %d][%s] 请求失败: %v\n",
			id, time.Now().Format("2006-01-02 15:04:05"), err)
		return false
	}

	defer func() {
		if resp != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	fmt.Printf("[工作进程 %d][%s] 状态码: %d 耗时: %v\n",
		id,
		time.Now().Format("2006-01-02 15:04:05"),
		resp.StatusCode,
		time.Since(start).Round(time.Millisecond),
	)
	return false
}

func main() {
	var (
		url      = flag.String("u", "https://s3.pysio.online/pcl2-ce/PCL2_CE_x64.exe", "请求的目标URL地址")
		interval = flag.Duration("i", 0*time.Second, "单个工作进程的请求间隔")
		workers  = flag.Int("w", 4, "并发工作进程数量")
	)
	flag.Parse()

	// 参数校验
	if *workers <= 0 {
		fmt.Println("[错误] 工作进程数量必须 > 0")
		return
	}
	if *interval < 0 {
		fmt.Println("[错误] 间隔时间不能为负数")
		return
	}

	stopChan := make(chan struct{})
	var wg sync.WaitGroup

	// 启动worker
	for i := 1; i <= *workers; i++ {
		wg.Add(1)
		go worker(i, &wg, *url, *interval, stopChan)
	}

	// 处理退出信号（示例需替换为实际信号处理）
	go func() {
		<-time.After(30 * time.Second) // 示例：30秒后自动停止
		close(stopChan)
	}()

	wg.Wait()
	fmt.Println("所有工作进程已停止")
}
