package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// IPResponse IP信息响应结构
type IPResponse struct {
	ProcessedString string `json:"processedString"`
	RawIspInfo      string `json:"rawIspInfo"`
}

// TelemetryData 遥测数据
type TelemetryData struct {
	ISPInfo string `json:"ispinfo"`
	DL      string `json:"dl"`
	UL      string `json:"ul"`
	Ping    string `json:"ping"`
	Jitter  string `json:"jitter"`
	Log     string `json:"log"`
	Extra   string `json:"extra"`
}

// downloadWorker 下载工作协程
func downloadWorker(ctx context.Context, wg *sync.WaitGroup, baseURL string, chunkSize int, results chan<- int64) {
	defer wg.Done()

	// 为HTTP客户端设置超时
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	url := fmt.Sprintf("%s%sr=%f&ckSize=%d", baseURL, urlSep(baseURL), rand.Float64(), chunkSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			start := time.Now()
			resp, err := client.Get(url)
			if err != nil {
				// 忽略错误继续下一次请求
				continue
			}

			// 读取响应体
			n, _ := io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			duration := time.Since(start).Seconds()
			
			// 如果在合理时间内完成，则发送结果
			if duration > 0.1 { // 至少100ms才计算
				results <- n
			}
		}
	}
}

// uploadWorker 上传工作协程
func uploadWorker(ctx context.Context, wg *sync.WaitGroup, baseURL string, data []byte, results chan<- int64) {
	defer wg.Done()

	// 为HTTP客户端设置超时
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 构造随机数据
			randomData := make([]byte, len(data))
			rand.Read(randomData)
			
			url := fmt.Sprintf("%s%sr=%f", baseURL, urlSep(baseURL), rand.Float64())
			start := time.Now()
			
			resp, err := client.Post(url, "application/octet-stream", bytes.NewReader(randomData))
			if err != nil {
				// 忽略错误继续下一次请求
				continue
			}
			
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			
			duration := time.Since(start).Seconds()
			
			// 如果在合理时间内完成，则发送结果
			if duration > 0.1 { // 至少100ms才计算
				results <- int64(len(randomData))
			}
		}
	}
}

// pingWorker Ping工作协程
func pingWorker(ctx context.Context, wg *sync.WaitGroup, baseURL string, results chan<- time.Duration) {
	defer wg.Done()

	// 为HTTP客户端设置超时
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for i := 0; i < config.CountPing; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			url := fmt.Sprintf("%s%sr=%f", baseURL, urlSep(baseURL), rand.Float64())
			start := time.Now()
			
			resp, err := client.Get(url)
			if err != nil {
				// 忽略错误继续下一次请求
				continue
			}
			
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			
			duration := time.Since(start)
			results <- duration
			
			// 等待一点时间再进行下次ping
			time.Sleep(200 * time.Millisecond)
		}
	}
}

// getIP 获取客户端IP信息
func getIP(baseURL string) (*IPResponse, error) {
	// 为HTTP客户端设置超时
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	ispParam := ""
	if config.GetIPISPInfo {
		ispParam = "isp=true&"
	}
	
	url := fmt.Sprintf("%s%s%sr=%f", baseURL, urlSep(baseURL), ispParam, rand.Float64())
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ipResp IPResponse
	if err := json.Unmarshal(body, &ipResp); err != nil {
		// 如果不是JSON格式，当作纯文本处理
		ipResp.ProcessedString = string(body)
	}

	return &ipResp, nil
}

// testDownload 实际执行下载测速
func testDownload(baseURL string) float64 {
	fmt.Print("Testing download speed... ")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TimeDLMax)*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	// 确保channel缓冲区足够大以避免阻塞
	streams := config.XhrDLMultistream
	if streams <= 0 {
		streams = 1 // 至少需要一个stream
	}
	results := make(chan int64, streams*2)
	
	// 启动多个并发下载协程
	concurrentStreams := config.XhrDLMultistream
	if concurrentStreams <= 0 {
		concurrentStreams = 1 // 单线程模式
	}
	
	for i := 0; i < concurrentStreams; i++ {
		wg.Add(1)
		go downloadWorker(ctx, &wg, baseURL, config.GarbagePHPChunkSize, results)
		// 只在多线程模式下添加延迟
		if concurrentStreams > 1 {
			time.Sleep(time.Duration(config.XhrMultistreamDelay) * time.Millisecond)
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var totalBytes int64
	startTime := time.Now()

	// 收集结果
	for bytes := range results {
		atomic.AddInt64(&totalBytes, bytes)
	}

	duration := time.Since(startTime).Seconds()
	
	// 计算速度 Mbps
	speed := float64(totalBytes) * 8 / duration / 1024 / 1024
	
	// 应用补偿因子
	speed = speed / config.OverheadCompensationFactor
	
	if config.UseMebibits {
		speed = speed * 1000 * 1000 / 1024 / 1024
	}
	
	fmt.Printf("%.2f Mbps\n", speed)
	return speed
}

// testUpload 实际执行上传测速
func testUpload(baseURL string) float64 {
	fmt.Print("Testing upload speed... ")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.TimeULMax)*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	// 确保channel缓冲区足够大以避免阻塞
	streams := config.XhrULMultistream
	if streams <= 0 {
		streams = 1 // 至少需要一个stream
	}
	results := make(chan int64, streams*2)
	
	// 创建测试数据
	testData := make([]byte, 1024*1024) // 1MB
	rand.Read(testData)
	
	// 启动多个并发上传协程
	concurrentStreams := config.XhrULMultistream
	if concurrentStreams <= 0 {
		concurrentStreams = 1 // 单线程模式
	}
	
	for i := 0; i < concurrentStreams; i++ {
		wg.Add(1)
		go uploadWorker(ctx, &wg, baseURL, testData, results)
		// 只在多线程模式下添加延迟
		if concurrentStreams > 1 {
			time.Sleep(time.Duration(config.XhrMultistreamDelay) * time.Millisecond)
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var totalBytes int64
	startTime := time.Now()

	// 收集结果
	for bytes := range results {
		atomic.AddInt64(&totalBytes, bytes)
	}

	duration := time.Since(startTime).Seconds()
	
	// 计算速度 Mbps
	speed := float64(totalBytes) * 8 / duration / 1024 / 1024
	
	// 应用补偿因子
	speed = speed / config.OverheadCompensationFactor
	
	if config.UseMebibits {
		speed = speed * 1000 * 1000 / 1024 / 1024
	}
	
	fmt.Printf("%.2f Mbps\n", speed)
	return speed
}

// testLatency 实际执行延迟测试
func testLatency(baseURL string) (float64, float64) {
	fmt.Print("Testing ping... ")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	results := make(chan time.Duration, config.CountPing)
	
	wg.Add(1)
	go pingWorker(ctx, &wg, baseURL, results)

	go func() {
		wg.Wait()
		close(results)
	}()

	var pings []float64
	for duration := range results {
		pings = append(pings, float64(duration)/float64(time.Millisecond))
	}

	if len(pings) == 0 {
		fmt.Println("Ping test failed")
		return 0, 0
	}

	// 计算平均ping和抖动
	var sum float64
	for _, p := range pings {
		sum += p
	}
	avgPing := sum / float64(len(pings))

	// 计算抖动 (连续延迟差值的平均)
	var jitterSum float64
	for i := 1; i < len(pings); i++ {
		diff := pings[i] - pings[i-1]
		if diff < 0 {
			diff = -diff
		}
		jitterSum += diff
	}
	jitter := jitterSum / float64(len(pings)-1)

	fmt.Printf("%.2f ms (jitter: %.2f ms)\n", avgPing, jitter)
	return avgPing, jitter
}

// sendTelemetry 发送遥测数据
func sendTelemetry(baseURL, extra string, result TestResult) error {
	if config.TelemetryLevel < 1 {
		return nil
	}

	// 为HTTP客户端设置超时
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	url := fmt.Sprintf("%s%sr=%f", baseURL, urlSep(baseURL), rand.Float64())

	telemetry := TelemetryData{
		ISPInfo: result.ClientIP,
		DL:      fmt.Sprintf("%.2f", result.DLSpeed),
		UL:      fmt.Sprintf("%.2f", result.ULSpeed),
		Ping:    fmt.Sprintf("%.2f", result.Ping),
		Jitter:  fmt.Sprintf("%.2f", result.Jitter),
		Extra:   extra,
	}

	data, err := json.Marshal(telemetry)
	if err != nil {
		return err
	}

	resp, err := client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// urlSep 返回URL分隔符
func urlSep(url string) string {
	for i := len(url) - 1; i >= 0; i-- {
		if url[i] == '?' {
			return "&"
		}
		if url[i] == '/' || url[i] == '\\' {
			break
		}
	}
	return "?"
}