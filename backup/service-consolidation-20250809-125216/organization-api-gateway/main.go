package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
)

// 项目默认租户配置
const (
	DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	DefaultTenantName     = "高谷集团"
)

var DefaultTenantID = uuid.MustParse(DefaultTenantIDString)

// ===== 服务端点配置 =====

type ServiceEndpoints struct {
	QueryService  string // 查询端Neo4j服务
	CommandService string // 命令端PostgreSQL服务
}

var endpoints = ServiceEndpoints{
	QueryService:  "http://localhost:8080",
	CommandService: "http://localhost:9090",
}

// ===== 标准API模型 =====

type StandardOrganization struct {
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	UnitType    string                 `json:"unit_type"`
	Status      string                 `json:"status"`
	Level       int                    `json:"level"`
	Path        string                 `json:"path"`
	SortOrder   int                    `json:"sort_order"`
	Description string                 `json:"description"`
	Profile     map[string]interface{} `json:"profile,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type StandardOrganizationsResponse struct {
	Organizations []StandardOrganization `json:"organizations"`
	TotalCount    int                    `json:"total_count"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
	HasNext       bool                   `json:"has_next"`
}

type StandardStatsResponse struct {
	TotalCount int                    `json:"total_count"`
	ByType     map[string]int         `json:"by_type"`
	ByStatus   map[string]int         `json:"by_status"`
	ByLevel    map[string]int         `json:"by_level"`
}

// ===== CoreHR API模型 =====

type CoreHROrganization struct {
	ID           string                 `json:"id"`
	Code         string                 `json:"code"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Status       string                 `json:"status"`
	Level        int                    `json:"level"`
	ParentCode   *string                `json:"parent_code,omitempty"`
	SortOrder    int                    `json:"sort_order"`
	Description  string                 `json:"description"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedTime  time.Time              `json:"created_time"`
	ModifiedTime time.Time              `json:"modified_time"`
}

type CoreHROrganizationsResponse struct {
	Data       []CoreHROrganization `json:"data"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	HasMore    bool                 `json:"has_more"`
}

type CoreHRStatsResponse struct {
	Summary struct {
		Total    int            `json:"total"`
		ByType   map[string]int `json:"by_type"`
		ByStatus map[string]int `json:"by_status"`
		ByLevel  map[string]int `json:"by_level"`
	} `json:"summary"`
}

// ===== 数据格式转换器 =====

type DataMapper struct {
	logger *log.Logger
}

func NewDataMapper(logger *log.Logger) *DataMapper {
	return &DataMapper{logger: logger}
}

// 标准格式 -> CoreHR格式
func (m *DataMapper) ToCorehrOrganization(std StandardOrganization) CoreHROrganization {
	return CoreHROrganization{
		ID:           std.Code, // CoreHR使用Code作为ID
		Code:         std.Code,
		Name:         std.Name,
		Type:         m.mapUnitTypeToCorehr(std.UnitType),
		Status:       strings.ToLower(std.Status),
		Level:        std.Level,
		ParentCode:   m.extractParentCode(std.Path),
		SortOrder:    std.SortOrder,
		Description:  std.Description,
		Metadata:     std.Profile,
		CreatedTime:  std.CreatedAt,
		ModifiedTime: std.UpdatedAt,
	}
}

func (m *DataMapper) ToCorehrResponse(std StandardOrganizationsResponse) CoreHROrganizationsResponse {
	data := make([]CoreHROrganization, len(std.Organizations))
	for i, org := range std.Organizations {
		data[i] = m.ToCorehrOrganization(org)
	}

	return CoreHROrganizationsResponse{
		Data:     data,
		Total:    std.TotalCount,
		Page:     std.Page,
		PageSize: std.PageSize,
		HasMore:  std.HasNext,
	}
}

func (m *DataMapper) ToCorehrStats(std StandardStatsResponse) CoreHRStatsResponse {
	return CoreHRStatsResponse{
		Summary: struct {
			Total    int            `json:"total"`
			ByType   map[string]int `json:"by_type"`
			ByStatus map[string]int `json:"by_status"`
			ByLevel  map[string]int `json:"by_level"`
		}{
			Total:    std.TotalCount,
			ByType:   m.mapTypesToCorehr(std.ByType),
			ByStatus: m.mapStatusToCorehr(std.ByStatus),
			ByLevel:  std.ByLevel,
		},
	}
}

// CoreHR格式 -> 标准格式 (用于命令)
func (m *DataMapper) FromCorehrCreateRequest(req map[string]interface{}) map[string]interface{} {
	standardReq := make(map[string]interface{})

	if name, ok := req["name"]; ok {
		standardReq["name"] = name
	}
	if orgType, ok := req["type"]; ok {
		standardReq["unit_type"] = m.mapCorehrTypeToStandard(fmt.Sprintf("%v", orgType))
	}
	if parentCode, ok := req["parent_code"]; ok {
		standardReq["parent_code"] = parentCode
	}
	if desc, ok := req["description"]; ok {
		standardReq["description"] = desc
	}
	if sortOrder, ok := req["sort_order"]; ok {
		standardReq["sort_order"] = sortOrder
	}

	return standardReq
}

// 辅助方法
func (m *DataMapper) mapUnitTypeToCorehr(unitType string) string {
	switch unitType {
	case "COMPANY":
		return "company"
	case "DEPARTMENT":
		return "department"
	case "TEAM":
		return "team"
	default:
		return "department"
	}
}

func (m *DataMapper) mapCorehrTypeToStandard(corehrType string) string {
	switch strings.ToLower(corehrType) {
	case "company":
		return "COMPANY"
	case "department":
		return "DEPARTMENT"
	case "team":
		return "TEAM"
	default:
		return "DEPARTMENT"
	}
}

func (m *DataMapper) mapTypesToCorehr(types map[string]int) map[string]int {
	result := make(map[string]int)
	for k, v := range types {
		result[strings.ToLower(k)] = v
	}
	return result
}

func (m *DataMapper) mapStatusToCorehr(status map[string]int) map[string]int {
	result := make(map[string]int)
	for k, v := range status {
		result[strings.ToLower(k)] = v
	}
	return result
}

func (m *DataMapper) extractParentCode(path string) *string {
	// 从路径中提取父代码，例如 "/1000000/1000001" -> "1000000"
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 1 {
		return &parts[len(parts)-2]
	}
	return nil
}

// ===== HTTP客户端 =====

type HTTPClient struct {
	client *http.Client
	logger *log.Logger
}

func NewHTTPClient(logger *log.Logger) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{Timeout: 10 * time.Second},
		logger: logger,
	}
}

func (c *HTTPClient) ForwardRequest(method, url string, body []byte, headers map[string]string) (*http.Response, error) {
	var reqBody io.Reader
	if len(body) > 0 {
		reqBody = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 复制头部
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	c.logger.Printf("转发请求: %s %s", method, url)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	return resp, nil
}

// ===== API网关 =====

type OrganizationAPIGateway struct {
	httpClient *HTTPClient
	mapper     *DataMapper
	logger     *log.Logger
}

func NewOrganizationAPIGateway(logger *log.Logger) *OrganizationAPIGateway {
	return &OrganizationAPIGateway{
		httpClient: NewHTTPClient(logger),
		mapper:     NewDataMapper(logger),
		logger:     logger,
	}
}

// ===== 标准API路径处理 =====

func (gw *OrganizationAPIGateway) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	// 直接转发到查询服务
	url := endpoints.QueryService + r.URL.Path
	if r.URL.RawQuery != "" {
		url += "?" + r.URL.RawQuery
	}

	headers := gw.extractHeaders(r)
	resp, err := gw.httpClient.ForwardRequest("GET", url, nil, headers)
	if err != nil {
		gw.logger.Printf("查询组织失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	gw.copyResponse(w, resp)
}

func (gw *OrganizationAPIGateway) GetOrganizationStats(w http.ResponseWriter, r *http.Request) {
	// 直接转发到查询服务
	url := endpoints.QueryService + r.URL.Path
	if r.URL.RawQuery != "" {
		url += "?" + r.URL.RawQuery
	}

	headers := gw.extractHeaders(r)
	resp, err := gw.httpClient.ForwardRequest("GET", url, nil, headers)
	if err != nil {
		gw.logger.Printf("查询组织统计失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	gw.copyResponse(w, resp)
}

func (gw *OrganizationAPIGateway) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	// 直接转发到命令服务
	body, err := io.ReadAll(r.Body)
	if err != nil {
		gw.logger.Printf("读取请求体失败: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	url := endpoints.CommandService + r.URL.Path
	headers := gw.extractHeaders(r)
	headers["Content-Type"] = "application/json"

	resp, err := gw.httpClient.ForwardRequest("POST", url, body, headers)
	if err != nil {
		gw.logger.Printf("创建组织失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	gw.copyResponse(w, resp)
}

func (gw *OrganizationAPIGateway) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	// 直接转发到命令服务
	body, err := io.ReadAll(r.Body)
	if err != nil {
		gw.logger.Printf("读取请求体失败: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	url := endpoints.CommandService + r.URL.Path
	headers := gw.extractHeaders(r)
	headers["Content-Type"] = "application/json"

	resp, err := gw.httpClient.ForwardRequest("PUT", url, body, headers)
	if err != nil {
		gw.logger.Printf("更新组织失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	gw.copyResponse(w, resp)
}

func (gw *OrganizationAPIGateway) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	// 直接转发到命令服务
	url := endpoints.CommandService + r.URL.Path
	headers := gw.extractHeaders(r)

	resp, err := gw.httpClient.ForwardRequest("DELETE", url, nil, headers)
	if err != nil {
		gw.logger.Printf("删除组织失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	gw.copyResponse(w, resp)
}

// ===== CoreHR API路径处理 =====

func (gw *OrganizationAPIGateway) GetCorehrOrganizations(w http.ResponseWriter, r *http.Request) {
	// 1. 转发到查询服务获取标准格式数据
	url := endpoints.QueryService + "/api/v1/organization-units"
	if r.URL.RawQuery != "" {
		url += "?" + r.URL.RawQuery
	}

	headers := gw.extractHeaders(r)
	resp, err := gw.httpClient.ForwardRequest("GET", url, nil, headers)
	if err != nil {
		gw.logger.Printf("查询组织失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 2. 读取标准格式响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		gw.logger.Printf("读取响应失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
		return
	}

	// 3. 解析标准格式
	var stdResp StandardOrganizationsResponse
	if err := json.Unmarshal(respBody, &stdResp); err != nil {
		gw.logger.Printf("解析标准响应失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 4. 转换为CoreHR格式
	corehrResp := gw.mapper.ToCorehrResponse(stdResp)

	// 5. 返回CoreHR格式响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(corehrResp); err != nil {
		gw.logger.Printf("编码CoreHR响应失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	gw.logger.Printf("✅ CoreHR组织查询成功，返回 %d 个组织", len(corehrResp.Data))
}

func (gw *OrganizationAPIGateway) GetCorehrOrganizationStats(w http.ResponseWriter, r *http.Request) {
	// 1. 转发到查询服务获取标准格式统计
	url := endpoints.QueryService + "/api/v1/organization-units/stats"
	if r.URL.RawQuery != "" {
		url += "?" + r.URL.RawQuery
	}

	headers := gw.extractHeaders(r)
	resp, err := gw.httpClient.ForwardRequest("GET", url, nil, headers)
	if err != nil {
		gw.logger.Printf("查询组织统计失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 2. 读取标准格式响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		gw.logger.Printf("读取响应失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
		return
	}

	// 3. 解析标准格式
	var stdResp StandardStatsResponse
	if err := json.Unmarshal(respBody, &stdResp); err != nil {
		gw.logger.Printf("解析标准统计响应失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 4. 转换为CoreHR格式
	corehrResp := gw.mapper.ToCorehrStats(stdResp)

	// 5. 返回CoreHR格式响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(corehrResp); err != nil {
		gw.logger.Printf("编码CoreHR统计响应失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	gw.logger.Printf("✅ CoreHR组织统计查询成功")
}

func (gw *OrganizationAPIGateway) CreateCorehrOrganization(w http.ResponseWriter, r *http.Request) {
	// 1. 读取CoreHR格式请求
	body, err := io.ReadAll(r.Body)
	if err != nil {
		gw.logger.Printf("读取请求体失败: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// 2. 解析CoreHR格式
	var corehrReq map[string]interface{}
	if err := json.Unmarshal(body, &corehrReq); err != nil {
		gw.logger.Printf("解析CoreHR请求失败: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// 3. 转换为标准格式
	stdReq := gw.mapper.FromCorehrCreateRequest(corehrReq)

	// 4. 编码标准请求
	stdBody, err := json.Marshal(stdReq)
	if err != nil {
		gw.logger.Printf("编码标准请求失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 5. 转发到命令服务
	url := endpoints.CommandService + "/api/v1/organization-units"
	headers := gw.extractHeaders(r)
	headers["Content-Type"] = "application/json"

	resp, err := gw.httpClient.ForwardRequest("POST", url, stdBody, headers)
	if err != nil {
		gw.logger.Printf("创建组织失败: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 6. 直接返回标准响应（CoreHR创建响应与标准响应兼容）
	gw.copyResponse(w, resp)
	gw.logger.Printf("✅ CoreHR组织创建成功")
}

// ===== 辅助方法 =====

func (gw *OrganizationAPIGateway) extractHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string)
	
	// 复制重要的头部
	importantHeaders := []string{
		"X-Tenant-ID", "Authorization", "Content-Type", 
		"Accept", "User-Agent", "X-Request-ID",
	}
	
	for _, header := range importantHeaders {
		if value := r.Header.Get(header); value != "" {
			headers[header] = value
		}
	}
	
	// 确保有默认租户ID
	if headers["X-Tenant-ID"] == "" {
		headers["X-Tenant-ID"] = DefaultTenantIDString
	}
	
	return headers
}

func (gw *OrganizationAPIGateway) copyResponse(w http.ResponseWriter, resp *http.Response) {
	// 复制头部
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	
	// 设置状态码
	w.WriteHeader(resp.StatusCode)
	
	// 复制响应体
	io.Copy(w, resp.Body)
}

// ===== 主程序 =====

func main() {
	logger := log.New(os.Stdout, "[API-GATEWAY] ", log.LstdFlags)

	// 创建API网关
	gateway := NewOrganizationAPIGateway(logger)

	// 创建HTTP路由器
	r := chi.NewRouter()

	// 中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// 标准组织API路径 (现有格式)
	r.Route("/api/v1/organization-units", func(r chi.Router) {
		// 查询端点
		r.Get("/", gateway.GetOrganizations)
		r.Get("/stats", gateway.GetOrganizationStats)
		
		// 命令端点
		r.Post("/", gateway.CreateOrganization)
		r.Put("/{code}", gateway.UpdateOrganization)
		r.Delete("/{code}", gateway.DeleteOrganization)
	})

	// CoreHR组织API路径 (新格式)
	r.Route("/api/v1/corehr/organizations", func(r chi.Router) {
		// 查询端点
		r.Get("/", gateway.GetCorehrOrganizations)
		r.Get("/stats", gateway.GetCorehrOrganizationStats)
		
		// 命令端点
		r.Post("/", gateway.CreateCorehrOrganization)
		r.Put("/{code}", gateway.UpdateOrganization) // 复用标准处理器
		r.Delete("/{code}", gateway.DeleteOrganization) // 复用标准处理器
	})

	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"service": "organization-api-gateway",
		})
	})

	// 根路径信息
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "Organization API Gateway",
			"version": "1.0.0",
			"paths": []string{
				"/api/v1/organization-units",
				"/api/v1/corehr/organizations",
			},
			"features": []string{
				"CQRS Architecture",
				"Dual-Path API Support",
				"Format Adapter Pattern",
				"Real-time Data Sync",
			},
		})
	})

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    ":8000", // 网关使用8000端口
		Handler: r,
	}

	// 优雅关闭
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		logger.Println("正在关闭API网关...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Printf("API网关关闭失败: %v", err)
		}
	}()

	logger.Printf("🚀 组织API网关启动成功 - 端口 :8000")
	logger.Printf("📍 标准API路径: http://localhost:8000/api/v1/organization-units")
	logger.Printf("📍 CoreHR API路径: http://localhost:8000/api/v1/corehr/organizations")
	logger.Printf("📍 查询服务后端: %s", endpoints.QueryService)
	logger.Printf("📍 命令服务后端: %s", endpoints.CommandService)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("API网关启动失败: %v", err)
	}

	logger.Println("API网关已关闭")
}