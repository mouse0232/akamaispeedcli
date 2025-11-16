package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/gizak/termui/v3"
)

// TestConfig 测速配置
type TestConfig struct {
	TestOrder                  string
	TimeULMax                  int
	TimeDLMax                  int
	TimeAuto                   bool
	TimeULGraceTime            float64
	TimeDLGraceTime            float64
	CountPing                  int
	URLDL                      string
	URLUL                      string
	URLPing                    string
	URLGetIP                   string
	GetIPISPInfo               bool
	GetIPISPInfoDistance       string
	XhrDLMultistream           int
	XhrULMultistream           int
	XhrMultistreamDelay        int
	XhrIgnoreErrors            int
	XhrDLUseBlob               bool
	XhrULBlobMegabytes         int
	GarbagePHPChunkSize        int
	EnableQuirks               bool
	PingAllowPerformanceAPI    bool
	OverheadCompensationFactor float64
	UseMebibits                bool
	TelemetryLevel             int
	URLTelemetry               string
	TelemetryExtra             string
}

// TestResult 测速结果
type TestResult struct {
	DLSpeed  float64
	ULSpeed  float64
	Ping     float64
	Jitter   float64
	ClientIP string
	ISPInfo  string
}

// Global variables
var (
	config TestConfig
	result TestResult
)

func init() {
	// 默认配置，使用实际的测速服务器地址
	config = TestConfig{
		TestOrder:                  "IP_D_U",
		TimeULMax:                  15,
		TimeDLMax:                  15,
		TimeAuto:                   true,
		TimeULGraceTime:            3,
		TimeDLGraceTime:            1.5,
		CountPing:                  10,
		URLDL:                      "https://speedtest.atlanta.linode.com/garbage.php",
		URLUL:                      "https://speedtest.atlanta.linode.com/empty.php",
		URLPing:                    "https://speedtest.atlanta.linode.com/empty.php",
		URLGetIP:                   "https://speedtest.atlanta.linode.com/getIP.php",
		GetIPISPInfo:               true,
		GetIPISPInfoDistance:       "km",
		XhrDLMultistream:           6,
		XhrULMultistream:           3,
		XhrMultistreamDelay:        300,
		XhrIgnoreErrors:            1,
		XhrDLUseBlob:               false,
		XhrULBlobMegabytes:         20,
		GarbagePHPChunkSize:        100,
		EnableQuirks:               true,
		PingAllowPerformanceAPI:    true,
		OverheadCompensationFactor: 1.06,
		UseMebibits:                false,
		TelemetryLevel:             0,
		URLTelemetry:               "https://speedtest.atlanta.linode.com/telemetry/telemetry.php",
		TelemetryExtra:             "",
	}

	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var (
		list       = flag.Bool("list", false, "List servers")
		help       = flag.Bool("help", false, "Show help")
		jsonOutput = flag.Bool("json", false, "Output in JSON format")
		simple     = flag.Bool("simple", false, "Output result only")
		concurrent = flag.Int("c", 0, "Concurrent threads (0 for default, 1 for single-threaded)")
	)

	flag.Parse()

	// 应用用户指定的并发数
	if *concurrent > 0 {
		config.XhrDLMultistream = *concurrent
		config.XhrULMultistream = *concurrent
	}

	if *help {
		showHelp()
		return
	}

	if *list {
		listServers()
		return
	}

	fmt.Println("Atlanta CLI Speed Test")
	fmt.Println("======================")

	// 初始化UI（如果需要）
	if !*jsonOutput && !*simple {
		if err := termui.Init(); err != nil {
			log.Printf("Failed to initialize termui: %v", err)
		} else {
			defer termui.Close()
		}
	}

	// 执行测速
	err := runSpeedTest(*jsonOutput, *simple)
	if err != nil {
		log.Fatal(err)
	}
}

func showHelp() {
	fmt.Println("Atlanta CLI Speed Test")
	fmt.Println("----------------------")
	fmt.Println("Usage: atlanta [options]")
	fmt.Println("")
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  atlanta                  Run a standard speed test")
	fmt.Println("  atlanta -list            List all available test servers")
	fmt.Println("  atlanta -json            Output results in JSON format")
	fmt.Println("  atlanta -simple          Output results only")
	fmt.Println("  atlanta -c 10            Set concurrent threads to 10")
	fmt.Println("  atlanta -c 1             Run in single-threaded mode")
}

func listServers() {
	fmt.Println("Available test servers:")
	// 这里应该从实际服务器列表获取
	fmt.Println("1. Server A - Location A")
	fmt.Println("2. Server B - Location B")
	fmt.Println("3. Server C - Location C")
}

func runSpeedTest(jsonOutput, simple bool) error {
	if !simple && !jsonOutput {
		fmt.Println("\nStarting speed test...")

		// 显示当前使用的并发设置
		if config.XhrDLMultistream == config.XhrULMultistream {
			fmt.Printf("Concurrent threads: %d\n", config.XhrDLMultistream)
		} else {
			fmt.Printf("Download threads: %d\n", config.XhrDLMultistream)
			fmt.Printf("Upload threads: %d\n", config.XhrULMultistream)
		}
		fmt.Println()
	}

	// 根据配置顺序执行测试
	tests := strings.Split(config.TestOrder, "")
	iRun := false
	dRun := false
	uRun := false
	pRun := false

	for _, test := range tests {
		switch test {
		case "I":
			if !iRun {
				if !simple && !jsonOutput {
					fmt.Print("Getting IP information... ")
				}
				getIPInfo()
				iRun = true
				if !simple && !jsonOutput {
					fmt.Printf("%s\n", result.ClientIP)
				}
			}
		case "D":
			if !dRun {
				//if !simple && !jsonOutput {
				//	fmt.Print("Testing download speed... ")
				//}
				dlSpeed := testDownload(config.URLDL)
				result.DLSpeed = dlSpeed
				dRun = true
				//if !simple && !jsonOutput {
				//	fmt.Printf("%.2f Mbps\n", dlSpeed)
				//}
			}
		case "U":
			if !uRun {
				//if !simple && !jsonOutput {
				//	fmt.Print("Testing upload speed... ")
				//}
				ulSpeed := testUpload(config.URLUL)
				result.ULSpeed = ulSpeed
				uRun = true
				//if !simple && !jsonOutput {
				//	fmt.Printf("%.2f Mbps\n", ulSpeed)
				//}
			}
		case "P":
			if !pRun {
				//if !simple && !jsonOutput {
				//	fmt.Print("Testing ping... ")
				//}
				ping, jitter := testLatency(config.URLPing)
				result.Ping = ping
				result.Jitter = jitter
				pRun = true
				//if !simple && !jsonOutput {
				//	fmt.Printf("%.2f ms (jitter: %.2f ms)\n", ping, jitter)
				//}
			}
		}
		time.Sleep(300 * time.Millisecond) // 模拟延迟
	}

	if jsonOutput {
		printJSONResult()
	} else if simple {
		printSimpleResult()
	} else {
		printResult()
	}

	return nil
}

func getIPInfo() {
	// 获取IP信息
	ipResp, err := getIP(config.URLGetIP)
	if err != nil {
		result.ClientIP = "Unknown"
		result.ISPInfo = ""
	} else {
		result.ClientIP = ipResp.ProcessedString
		result.ISPInfo = ipResp.RawIspInfo
	}
}

func printResult() {
	fmt.Println("\nResults:")
	fmt.Println("--------")
	fmt.Printf("Download: %.2f Mbps\n", result.DLSpeed)
	fmt.Printf("Upload: %.2f Mbps\n", result.ULSpeed)
	fmt.Printf("Ping: %.2f ms\n", result.Ping)
	fmt.Printf("Jitter: %.2f ms\n", result.Jitter)
	fmt.Printf("IP: %s\n", result.ClientIP)
}

func printSimpleResult() {
	fmt.Printf("Download: %.2f Mbps\n", result.DLSpeed)
	fmt.Printf("Upload: %.2f Mbps\n", result.ULSpeed)
	fmt.Printf("Ping: %.2f ms\n", result.Ping)
	fmt.Printf("Jitter: %.2f ms\n", result.Jitter)
}

func printJSONResult() {
	fmt.Printf(`{
  "download": %.2f,
  "upload": %.2f,
  "ping": %.2f,
  "jitter": %.2f,
  "ip": "%s"
}\n`, result.DLSpeed, result.ULSpeed, result.Ping, result.Jitter, result.ClientIP)
}
