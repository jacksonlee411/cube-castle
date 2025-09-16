package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// 测试结果结构
type TestResult struct {
	Name        string
	Passed      bool
	Description string
	Details     string
	Duration    time.Duration
}

type TestSuite struct {
	Results     []TestResult
	StartTime   time.Time
	TotalTests  int
	PassedTests int
	FailedTests int
}

func NewTestSuite() *TestSuite {
	return &TestSuite{
		Results:   make([]TestResult, 0),
		StartTime: time.Now(),
	}
}

func (ts *TestSuite) RunTest(name, description string, testFunc func() (bool, string)) {
	fmt.Printf("🧪 [%d] %s\n", ts.TotalTests+1, name)
	if description != "" {
		fmt.Printf("    描述: %s\n", description)
	}

	start := time.Now()
	passed, details := testFunc()
	duration := time.Since(start)

	result := TestResult{
		Name:        name,
		Passed:      passed,
		Description: description,
		Details:     details,
		Duration:    duration,
	}

	ts.Results = append(ts.Results, result)
	ts.TotalTests++

	if passed {
		ts.PassedTests++
		fmt.Printf("    ✅ PASS (%.2fms)\n", float64(duration.Nanoseconds())/1000000)
	} else {
		ts.FailedTests++
		fmt.Printf("    ❌ FAIL (%.2fms)\n", float64(duration.Nanoseconds())/1000000)
		fmt.Printf("    详情: %s\n", details)
	}
	fmt.Println()
}

func (ts *TestSuite) PrintSummary() {
	duration := time.Since(ts.StartTime)
	passRate := float64(ts.PassedTests) / float64(ts.TotalTests) * 100

	fmt.Println("=== 测试结果汇总 ===")
	fmt.Printf("总测试数: %d\n", ts.TotalTests)
	fmt.Printf("通过数: %d\n", ts.PassedTests)
	fmt.Printf("失败数: %d\n", ts.FailedTests)
	fmt.Printf("通过率: %.1f%%\n", passRate)
	fmt.Printf("总耗时: %.2f秒\n", duration.Seconds())
	fmt.Println()

	if ts.FailedTests == 0 {
		fmt.Println("🎉 所有测试通过！时态管理API已达到生产就绪标准！")
	} else {
		fmt.Printf("❌ %d个测试失败，需要修复后才能部署生产环境\n", ts.FailedTests)
	}
}

// HTTP客户端辅助函数
func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func httpPost(url string, jsonData string) ([]byte, error) {
	resp, err := http.Post(url, "application/json", strings.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// 数据库查询辅助函数
func execPSQL(query string) (string, error) {
	cmd := exec.Command("psql", "-h", "localhost", "-U", "user", "-d", "cubecastle", "-t", "-c", query)
	cmd.Env = append(cmd.Env, "PGPASSWORD=password")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func main() {
	fmt.Println("🚀 时态管理API深度测试验证")
	fmt.Printf("开始时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()

	ts := NewTestSuite()

	baseURL := "http://localhost:9091/api/v1/organization-units"
	testOrg := "1000001"

	// 第1组：基础功能测试
	fmt.Println("=== 第1组：基础功能测试 ===")

	ts.RunTest("服务健康检查", "验证时态API服务运行状态", func() (bool, string) {
		data, err := httpGet("http://localhost:9091/health")
		if err != nil {
			return false, fmt.Sprintf("请求失败: %v", err)
		}

		var health map[string]interface{}
		if err := json.Unmarshal(data, &health); err != nil {
			return false, fmt.Sprintf("解析响应失败: %v", err)
		}

		if status, ok := health["status"].(string); ok && status == "healthy" {
			return true, "服务状态正常"
		}

		return false, fmt.Sprintf("服务状态异常: %s", string(data))
	})

	ts.RunTest("基础组织查询", "验证能够查询测试组织", func() (bool, string) {
		data, err := httpGet(fmt.Sprintf("%s/%s", baseURL, testOrg))
		if err != nil {
			return false, fmt.Sprintf("请求失败: %v", err)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(data, &response); err != nil {
			return false, fmt.Sprintf("解析响应失败: %v", err)
		}

		if resultCount, ok := response["result_count"].(float64); ok && resultCount == 1 {
			return true, "成功查询到1个组织记录"
		}

		return false, fmt.Sprintf("查询结果异常: %s", string(data))
	})

	ts.RunTest("时态字段完整性", "验证响应包含必需的时态字段", func() (bool, string) {
		data, err := httpGet(fmt.Sprintf("%s/%s", baseURL, testOrg))
		if err != nil {
			return false, fmt.Sprintf("请求失败: %v", err)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(data, &response); err != nil {
			return false, fmt.Sprintf("解析响应失败: %v", err)
		}

		orgs, ok := response["organizations"].([]interface{})
		if !ok || len(orgs) == 0 {
			return false, "未找到组织记录"
		}

		org := orgs[0].(map[string]interface{})
		requiredFields := []string{"version", "effective_date", "is_current"}

		for _, field := range requiredFields {
			if _, exists := org[field]; !exists {
				return false, fmt.Sprintf("缺失字段: %s", field)
			}
		}

		return true, "所有时态字段完整"
	})

	// 第2组：时态查询功能测试
	fmt.Println("=== 第2组：时态查询功能测试 ===")

	ts.RunTest("未来日期查询", "验证未来日期查询功能", func() (bool, string) {
		url := fmt.Sprintf("%s/%s?as_of_date=2026-01-01", baseURL, testOrg)
		data, err := httpGet(url)
		if err != nil {
			return false, fmt.Sprintf("请求失败: %v", err)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(data, &response); err != nil {
			return false, fmt.Sprintf("解析响应失败: %v", err)
		}

		if resultCount, ok := response["result_count"].(float64); ok && resultCount == 1 {
			return true, "未来日期查询成功"
		}

		return false, fmt.Sprintf("查询结果异常: %s", string(data))
	})

	ts.RunTest("过去日期查询", "验证过去日期查询返回NOT_FOUND", func() (bool, string) {
		url := fmt.Sprintf("%s/%s?as_of_date=2020-01-01", baseURL, testOrg)
		data, err := httpGet(url)
		if err != nil {
			return false, fmt.Sprintf("请求失败: %v", err)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(data, &response); err != nil {
			return false, fmt.Sprintf("解析响应失败: %v", err)
		}

		if resultCount, ok := response["result_count"].(float64); ok && resultCount == 0 {
			return true, "过去日期查询未找到记录（符合预期）"
		}

		return false, fmt.Sprintf("查询结果异常: %s", string(data))
	})

	ts.PrintSummary()
}

