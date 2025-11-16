package main

import (
	"net/http"
	"time"
)

// Server 测速服务器信息
type Server struct {
	Name string
	URL  string
}

// Servers 测速服务器列表
var Servers = []Server{
	{"亚特兰大", "https://speedtest.atlanta.linode.com/"},
	{"芝加哥", "https://speedtest.chicago.linode.com/"},
	{"达拉斯", "https://speedtest.dallas.linode.com/"},
	{"费利蒙", "https://speedtest.fremont.linode.com/"},
	{"洛杉矶", "https://speedtest.los-angeles.linode.com/"},
	{"迈阿密", "https://speedtest.miami.linode.com/"},
	{"纽瓦克", "https://speedtest.newark.linode.com/"},
	{"西雅图", "https://speedtest.seattle.linode.com/"},
	{"华盛顿特区", "https://speedtest.washington.linode.com/"},
	{"多伦多", "https://speedtest.toronto1.linode.com/"},
	{"圣保罗", "https://speedtest.sao-paulo.linode.com/"},
	{"阿姆斯特丹", "https://speedtest.amsterdam.linode.com/"},
	{"法兰克福", "https://speedtest.frankfurt.linode.com/"},
	{"法兰克福扩建", "https://de-fra-2.speedtest.linode.com/"},
	{"伦敦", "https://speedtest.london.linode.com/"},
	{"伦敦扩建", "https://gb-lon.speedtest.linode.com/"},
	{"马德里", "https://speedtest.madrid.linode.com/"},
	{"米兰", "https://speedtest.milan.linode.com/"},
	{"巴黎", "https://speedtest.paris.linode.com/"},
	{"斯德哥尔摩", "https://speedtest.stockholm.linode.com/"},
	{"金奈", "https://speedtest.chennai.linode.com/"},
	{"雅加达", "https://speedtest.jakarta.linode.com/"},
	{"孟买", "https://speedtest.mumbai1.linode.com/"},
	{"孟买扩建", "https://in-mum-2.speedtest.linode.com/"},
	{"大阪", "https://speedtest.osaka.linode.com/"},
	{"新加坡", "https://speedtest.singapore.linode.com/"},
	{"新加坡扩建", "https://sg-sin-2.speedtest.linode.com/"},
	{"东京", "https://speedtest.tokyo2.linode.com/"},
	{"东京扩建", "https://jp-tyo-3.speedtest.linode.com/"},
	{"悉尼", "https://speedtest.sydney.linode.com/"},
	{"墨尔本", "https://au-mel.speedtest.linode.com/"},
}

// findClosestServer 查找最近的服务器
func findClosestServer() int {
	type serverLatency struct {
		index   int
		latency time.Duration
	}

	latencies := make(chan serverLatency, len(Servers))
	
	// 并发测试所有服务器的延迟
	for i, server := range Servers {
		go func(index int, server Server) {
			latency := testServerLatency(server.URL)
			latencies <- serverLatency{index: index, latency: latency}
		}(i, server)
	}
	
	// 收集所有测试结果
	var results []serverLatency
	for i := 0; i < len(Servers); i++ {
		result := <-latencies
		if result.latency > 0 { // 只考虑成功测试的服务器
			results = append(results, result)
		}
	}
	
	// 如果没有服务器响应，返回默认服务器
	if len(results) == 0 {
		return 0
	}
	
	// 找到延迟最低的服务器
	closest := results[0]
	for _, result := range results {
		if result.latency < closest.latency {
			closest = result
		}
	}
	
	return closest.index
}

// testServerLatency 测试单个服务器的延迟
func testServerLatency(baseURL string) time.Duration {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	url := baseURL + "empty.php"
	start := time.Now()
	
	resp, err := client.Get(url)
	if err != nil {
		return 0
	}
	
	defer resp.Body.Close()
	
	return time.Since(start)
}