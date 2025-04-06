package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func getEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func worker(ctx context.Context, id int, wg *sync.WaitGroup, url string, interval time.Duration, userAgent string) {
	defer wg.Done()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	noInterval := interval == 0

	fmt.Printf("Worker %d 启动 [模式: %s]\n",
		id,
		map[bool]string{true: "无间隔", false: fmt.Sprintf("间隔 %v", interval)}[noInterval],
	)

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d 停止\n", id)
			return
		default:
			start := time.Now()

			// 创建带有UA的请求
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				fmt.Printf("[Worker %d][%s] 创建请求失败: %v\n",
					id, time.Now().Format("2006-01-02 15:04:05"), err)
				continue
			}
			req.Header.Set("User-Agent", userAgent)

			resp, err := client.Do(req)
			if handled := handleResponse(id, resp, err, start); handled {
				return
			}

			if !noInterval {
				select {
				case <-ctx.Done():
					return
				case <-time.After(interval):
				}
			}
		}
	}
}

func handleResponse(id int, resp *http.Response, err error, start time.Time) (stop bool) {
	if err != nil {
		fmt.Printf("[Worker %d][%s] 请求失败: %v\n",
			id, time.Now().Format("2006-01-02 15:04:05"), err)
		return false
	}

	defer func() {
		if resp != nil && resp.Body != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	fmt.Printf("[Worker %d][%s] 状态码: %d 耗时: %v\n",
		id,
		time.Now().Format("2006-01-02 15:04:05"),
		resp.StatusCode,
		time.Since(start).Round(time.Millisecond),
	)
	return false
}

func main() {
	var (
		url = flag.String("u",
			getEnvString("u", "https://s3.pysio.online/pcl2-ce/PCL2_CE_x64.exe"),
			"请求的目标URL地址")

		interval = flag.Duration("i",
			getEnvDuration("i", 0),
			"请求间隔时间")

		workers = flag.Int("w",
			getEnvInt("w", 4),
			"并发Worker数量")

		userAgent = flag.String("ua",
			getEnvString("ua", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36"),
			"User-Agent")
	)
	flag.Parse()

	// 参数校验
	if *workers <= 0 {
		fmt.Println("[错误] worker数量必须 > 0")
		return
	}
	if *interval < 0 {
		fmt.Println("[错误] 间隔时间不能为负数")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	for i := 1; i <= *workers; i++ {
		wg.Add(1)
		go worker(ctx, i, &wg, *url, *interval, *userAgent)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("程序已启动，按 Ctrl+C 停止...")
	<-sigChan

	fmt.Println("\n接收到停止信号，停止中...")
	cancel()
	wg.Wait()

	fmt.Println("所有Worker已安全停止")
}
