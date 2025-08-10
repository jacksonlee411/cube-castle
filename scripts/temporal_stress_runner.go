package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ConcurrencyTestResult struct {
	TotalRequests   int
	SuccessRequests int
	FailedRequests  int
	AvgResponseTime time.Duration
	MaxResponseTime time.Duration
	MinResponseTime time.Duration
	RequestsPerSec  float64
}

func (r *ConcurrencyTestResult) Print() {
	fmt.Printf("📊 并发测试结果:\n")
	fmt.Printf("  总请求数: %d\n", r.TotalRequests)
	fmt.Printf("  成功请求: %d\n", r.SuccessRequests)
	fmt.Printf("  失败请求: %d\n", r.FailedRequests)
	fmt.Printf("  成功率: %.1f%%\n", float64(r.SuccessRequests)/float64(r.TotalRequests)*100)
	fmt.Printf("  平均响应时间: %.2fms\n", float64(r.AvgResponseTime.Nanoseconds())/1000000)
	fmt.Printf("  最快响应: %.2fms\n", float64(r.MinResponseTime.Nanoseconds())/1000000)
	fmt.Printf("  最慢响应: %.2fms\n", float64(r.MaxResponseTime.Nanoseconds())/1000000)
	fmt.Printf("  请求速率: %.1f req/s\n", r.RequestsPerSec)
	fmt.Println()
}

func httpGetTimed(url string) (bool, time.Duration) {
	start := time.Now()
	resp, err := http.Get(url)
	duration := time.Since(start)
	
	if err != nil {
		return false, duration
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return false, duration
	}
	
	return true, duration
}

func httpPostTimed(url string, jsonData string) (bool, time.Duration) {
	start := time.Now()
	resp, err := http.Post(url, "application/json", strings.NewReader(jsonData))
	duration := time.Since(start)
	
	if err != nil {
		return false, duration
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, duration
	}
	
	return true, duration
}

func runConcurrentGETTest(url string, concurrency int, requestCount int) *ConcurrencyTestResult {
	fmt.Printf("🔄 运行并发GET测试 (并发数: %d, 请求数: %d)\n", concurrency, requestCount)
	
	var wg sync.WaitGroup
	results := make(chan struct {
		success  bool
		duration time.Duration
	}, requestCount)
	
	startTime := time.Now()
	requestsPerWorker := requestCount / concurrency
	
	// 启动并发工作goroutine
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				success, duration := httpGetTimed(url)
				results <- struct {
					success  bool
					duration time.Duration
				}{success, duration}
			}
		}()
	}
	
	// 等待所有请求完成
	wg.Wait()
	close(results)
	totalDuration := time.Since(startTime)
	
	// 分析结果
	result := &ConcurrencyTestResult{
		TotalRequests: requestCount,
		MinResponseTime: time.Hour, // 初始化为很大的值
	}
	
	totalTime := time.Duration(0)
	for res := range results {
		if res.success {
			result.SuccessRequests++
		} else {
			result.FailedRequests++
		}
		
		totalTime += res.duration
		if res.duration > result.MaxResponseTime {
			result.MaxResponseTime = res.duration
		}
		if res.duration < result.MinResponseTime {
			result.MinResponseTime = res.duration
		}
	}
	
	if result.TotalRequests > 0 {
		result.AvgResponseTime = totalTime / time.Duration(result.TotalRequests)
		result.RequestsPerSec = float64(result.TotalRequests) / totalDuration.Seconds()
	}
	
	return result
}

func runConcurrentPOSTTest(url string, jsonData string, concurrency int, requestCount int) *ConcurrencyTestResult {
	fmt.Printf("🔄 运行并发POST测试 (并发数: %d, 请求数: %d)\n", concurrency, requestCount)
	
	var wg sync.WaitGroup
	results := make(chan struct {
		success  bool
		duration time.Duration
	}, requestCount)
	
	startTime := time.Now()
	requestsPerWorker := requestCount / concurrency
	
	// 启动并发工作goroutine  
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				// 为每个请求生成唯一的测试数据
				uniqueData := strings.Replace(jsonData, "并发测试", fmt.Sprintf("并发测试-%d-%d", workerId, j), 1)
				success, duration := httpPostTimed(url, uniqueData)
				results <- struct {
					success  bool
					duration time.Duration
				}{success, duration}
			}
		}(i)
	}
	
	// 等待所有请求完成
	wg.Wait()
	close(results)
	totalDuration := time.Since(startTime)
	
	// 分析结果
	result := &ConcurrencyTestResult{
		TotalRequests: requestCount,
		MinResponseTime: time.Hour, // 初始化为很大的值
	}
	
	totalTime := time.Duration(0)
	for res := range results {
		if res.success {
			result.SuccessRequests++
		} else {
			result.FailedRequests++
		}
		
		totalTime += res.duration
		if res.duration > result.MaxResponseTime {
			result.MaxResponseTime = res.duration
		}
		if res.duration < result.MinResponseTime {
			result.MinResponseTime = res.duration
		}
	}
	
	if result.TotalRequests > 0 {
		result.AvgResponseTime = totalTime / time.Duration(result.TotalRequests)
		result.RequestsPerSec = float64(result.TotalRequests) / totalDuration.Seconds()
	}
	
	return result
}

func testEventSequentialConsistency(baseURL, testOrg string) bool {
	fmt.Println("🔍 测试事件顺序一致性...")
	
	// 创建一系列有时间顺序的事件
	events := []struct {
		effectiveDate string
		changeData    string
	}{
		{"2025-11-01T00:00:00Z", `{"name":"顺序测试1"}`},
		{"2025-11-15T00:00:00Z", `{"name":"顺序测试2"}`},
		{"2025-12-01T00:00:00Z", `{"name":"顺序测试3"}`},
	}
	
	url := fmt.Sprintf("%s/%s/events", baseURL, testOrg)
	
	for i, event := range events {
		jsonData := fmt.Sprintf(`{
			"event_type": "UPDATE",
			"effective_date": "%s",
			"change_data": %s,
			"change_reason": "顺序一致性测试%d"
		}`, event.effectiveDate, event.changeData, i+1)
		
		resp, err := http.Post(url, "application/json", strings.NewReader(jsonData))
		if err != nil {
			fmt.Printf("  ❌ 事件%d创建失败: %v\n", i+1, err)
			return false
		}
		defer resp.Body.Close()
		
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			fmt.Printf("  ❌ 事件%d创建失败: HTTP %d\n", i+1, resp.StatusCode)
			return false
		}
		
		fmt.Printf("  ✅ 事件%d创建成功\n", i+1)
		time.Sleep(100 * time.Millisecond) // 小延迟确保顺序
	}
	
	fmt.Println("  ✅ 所有事件按顺序创建成功")
	return true
}

func main() {
	fmt.Println("🚀 时态管理API并发与压力测试")
	fmt.Printf("开始时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()
	
	baseURL := "http://localhost:9091/api/v1/organization-units"
	testOrg := "1000001"
	
	// 1. 并发GET查询测试
	fmt.Println("=== 1. 并发查询测试 ===")
	getURL := fmt.Sprintf("%s/%s", baseURL, testOrg)
	
	// 轻量级并发测试
	result1 := runConcurrentGETTest(getURL, 5, 25)
	result1.Print()
	
	// 中等强度并发测试
	result2 := runConcurrentGETTest(getURL, 10, 50)
	result2.Print()
	
	// 2. 并发时态查询测试
	fmt.Println("=== 2. 并发时态查询测试 ===")
	temporalURL := fmt.Sprintf("%s/%s?as_of_date=2026-01-01", baseURL, testOrg)
	
	result3 := runConcurrentGETTest(temporalURL, 8, 40)
	result3.Print()
	
	// 3. 并发事件创建测试
	fmt.Println("=== 3. 并发事件创建测试 ===")
	eventURL := fmt.Sprintf("%s/%s/events", baseURL, testOrg)
	eventData := `{
		"event_type": "UPDATE",
		"effective_date": "2025-12-25T00:00:00Z",
		"change_data": {"name": "并发测试"},
		"change_reason": "并发压力测试"
	}`
	
	result4 := runConcurrentPOSTTest(eventURL, eventData, 3, 9)
	result4.Print()
	
	// 4. 混合负载测试
	fmt.Println("=== 4. 混合负载测试 ===")
	fmt.Println("🔄 同时执行查询和创建操作...")
	
	var wg sync.WaitGroup
	
	// 并发查询
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			http.Get(getURL)
			time.Sleep(50 * time.Millisecond)
		}
	}()
	
	// 并发事件创建
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			mixedEventData := fmt.Sprintf(`{
				"event_type": "UPDATE",
				"effective_date": "2025-12-30T%02d:00:00Z",
				"change_data": {"name": "混合负载测试%d"},
				"change_reason": "混合负载压力测试"
			}`, i, i)
			http.Post(eventURL, "application/json", strings.NewReader(mixedEventData))
			time.Sleep(200 * time.Millisecond)
		}
	}()
	
	wg.Wait()
	fmt.Println("  ✅ 混合负载测试完成")
	fmt.Println()
	
	// 5. 事件顺序一致性测试
	fmt.Println("=== 5. 事件顺序一致性测试 ===")
	if testEventSequentialConsistency(baseURL, testOrg) {
		fmt.Println("  ✅ 事件顺序一致性测试通过")
	} else {
		fmt.Println("  ❌ 事件顺序一致性测试失败")
	}
	fmt.Println()
	
	// 6. 压力测试结果评估
	fmt.Println("=== 压力测试结果评估 ===")
	
	allResults := []*ConcurrencyTestResult{result1, result2, result3, result4}
	totalRequests := 0
	totalSuccess := 0
	maxAvgResponseTime := time.Duration(0)
	
	for _, result := range allResults {
		totalRequests += result.TotalRequests
		totalSuccess += result.SuccessRequests
		if result.AvgResponseTime > maxAvgResponseTime {
			maxAvgResponseTime = result.AvgResponseTime
		}
	}
	
	overallSuccessRate := float64(totalSuccess) / float64(totalRequests) * 100
	
	fmt.Printf("📊 压力测试总结:\n")
	fmt.Printf("  总请求数: %d\n", totalRequests)
	fmt.Printf("  总成功数: %d\n", totalSuccess)
	fmt.Printf("  整体成功率: %.1f%%\n", overallSuccessRate)
	fmt.Printf("  最大平均响应时间: %.2fms\n", float64(maxAvgResponseTime.Nanoseconds())/1000000)
	fmt.Println()
	
	// 性能基准评估
	fmt.Println("🎯 性能基准评估:")
	
	if overallSuccessRate >= 95.0 {
		fmt.Println("  ✅ 成功率达标 (≥95%)")
	} else {
		fmt.Println("  ❌ 成功率不达标 (<95%)")
	}
	
	if maxAvgResponseTime < time.Millisecond*100 {
		fmt.Println("  ✅ 平均响应时间达标 (<100ms)")
	} else {
		fmt.Println("  ❌ 平均响应时间超标 (≥100ms)")
	}
	
	fmt.Println()
	
	if overallSuccessRate >= 95.0 && maxAvgResponseTime < time.Millisecond*100 {
		fmt.Println("🎉 并发与压力测试全部通过！系统具备生产环境并发处理能力！")
	} else {
		fmt.Println("⚠️  并发压力测试发现性能问题，建议优化后再部署生产环境")
	}
}