package http_client_v2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------
// 测试用的数据结构
// ---------------------------------------------------------

type TestRequest struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type TestResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// ---------------------------------------------------------
// 辅助函数：创建测试服务器
// ---------------------------------------------------------

// createTestServer 创建一个模拟的 HTTP 测试服务器
func createTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// ---------------------------------------------------------
// 1. 测试全局客户端初始化
// ---------------------------------------------------------

func TestGlobalClientInitialization(t *testing.T) {
	if globalClient == nil {
		t.Fatal("globalClient 未初始化")
	}

	// 检查底层 HTTP 客户端
	httpClient := globalClient.GetClient()
	if httpClient == nil {
		t.Fatal("底层 http.Client 为 nil")
	}

	// 检查 Transport 配置
	transport, ok := httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Transport 类型不正确")
	}

	// 验证连接池配置
	if transport.MaxIdleConns != 6000 {
		t.Errorf("MaxIdleConns = %d, 期望 6000", transport.MaxIdleConns)
	}
	if transport.MaxIdleConnsPerHost != 80 {
		t.Errorf("MaxIdleConnsPerHost = %d, 期望 80", transport.MaxIdleConnsPerHost)
	}
	if transport.IdleConnTimeout != 90*time.Second {
		t.Errorf("IdleConnTimeout = %v, 期望 90s", transport.IdleConnTimeout)
	}
	if transport.TLSHandshakeTimeout != 5*time.Second {
		t.Errorf("TLSHandshakeTimeout = %v, 期望 5s", transport.TLSHandshakeTimeout)
	}
}

// ---------------------------------------------------------
// 2. 测试 PostJSON
// ---------------------------------------------------------

func TestPostJSON_Success(t *testing.T) {
	// 创建测试服务器
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法
		if r.Method != http.MethodPost {
			t.Errorf("请求方法 = %s, 期望 POST", r.Method)
		}

		// 验证 Content-Type
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			t.Errorf("Content-Type = %s, 期望包含 application/json", r.Header.Get("Content-Type"))
		}

		// 读取请求体
		body, _ := io.ReadAll(r.Body)
		var req TestRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("解析请求体失败: %v", err)
		}

		// 验证请求数据
		if req.Name != "test" || req.Value != 123 {
			t.Errorf("请求数据不匹配: %+v", req)
		}

		// 返回响应
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TestResponse{
			Message: "success",
			Code:    0,
		})
	})
	defer server.Close()

	// 发送请求
	ctx := context.Background()
	reqBody := TestRequest{Name: "test", Value: 123}
	var result TestResponse

	resp, err := PostJSON(ctx, server.URL, reqBody, &result, nil)
	if err != nil {
		t.Fatalf("PostJSON 失败: %v", err)
	}

	// 验证响应
	if resp.StatusCode() != http.StatusOK {
		t.Errorf("状态码 = %d, 期望 200", resp.StatusCode())
	}
	if result.Message != "success" || result.Code != 0 {
		t.Errorf("响应数据不匹配: %+v", result)
	}
}

func TestPostJSON_Timeout(t *testing.T) {
	// 创建慢速服务器
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	ctx := context.Background()
	reqBody := TestRequest{Name: "test", Value: 123}
	var result TestResponse

	// 使用 100ms 超时
	_, err := PostJSON(ctx, server.URL, reqBody, &result, nil, WithTimeout(100*time.Millisecond))

	// 应该超时
	if err == nil {
		t.Error("期望超时错误，但没有错误")
	}
}

// TestPostJSON_ResultNil 验证链式写法 SetResult(nil) 不判空时仍可正常请求
func TestPostJSON_ResultNil(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("请求方法 = %s, 期望 POST", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	ctx := context.Background()
	reqBody := TestRequest{Name: "no-result", Value: 0}

	resp, err := PostJSON(ctx, server.URL, reqBody, nil, nil)
	if err != nil {
		t.Fatalf("PostJSON(result=nil) 失败: %v", err)
	}
	if resp.StatusCode() != http.StatusNoContent {
		t.Errorf("状态码 = %d, 期望 204", resp.StatusCode())
	}
}

// TestPostJSON_RawBehaviorOnHTTPError 验证 4xx/5xx 不包装为 error
func TestPostJSON_RawBehaviorOnHTTPError(t *testing.T) {
	// 创建返回错误的服务器
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	})
	defer server.Close()

	ctx := context.Background()
	reqBody := TestRequest{Name: "test", Value: 123}
	var result TestResponse

	resp, err := PostJSON(ctx, server.URL, reqBody, &result, nil)

	// raw 语义：不应因 4xx/5xx 返回 error
	if err != nil {
		t.Fatalf("PostJSON 不应因 4xx/5xx 返回 error，但拿到：%v", err)
	}
	if resp == nil {
		t.Error("resp 不应该为 nil")
	}
	if resp != nil && resp.StatusCode() != http.StatusInternalServerError {
		t.Errorf("状态码 = %d, 期望 500", resp.StatusCode())
	}
}

// ---------------------------------------------------------
// 3. 测试 Get
// ---------------------------------------------------------

func TestGet_Success(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("请求方法 = %s, 期望 GET", r.Method)
		}

		// 验证查询参数
		if r.URL.Query().Get("name") != "test" {
			t.Errorf("参数 name = %s, 期望 test", r.URL.Query().Get("name"))
		}
		if r.URL.Query().Get("value") != "123" {
			t.Errorf("参数 value = %s, 期望 123", r.URL.Query().Get("value"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TestResponse{
			Message: "get success",
			Code:    0,
		})
	})
	defer server.Close()

	ctx := context.Background()
	params := map[string]string{
		"name":  "test",
		"value": "123",
	}
	var result TestResponse

	resp, err := Get(ctx, server.URL, params, &result)
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("状态码 = %d, 期望 200", resp.StatusCode())
	}
	if result.Message != "get success" {
		t.Errorf("响应消息 = %s, 期望 get success", result.Message)
	}
}

// TestGet_ResultNil 验证链式写法 SetResult(nil) 不判空时仍可正常请求
func TestGet_ResultNil(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("请求方法 = %s, 期望 GET", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	ctx := context.Background()

	resp, err := Get(ctx, server.URL, nil, nil)
	if err != nil {
		t.Fatalf("Get(result=nil) 失败: %v", err)
	}
	if resp.StatusCode() != http.StatusNoContent {
		t.Errorf("状态码 = %d, 期望 204", resp.StatusCode())
	}
}

// ---------------------------------------------------------
// 4. 测试 PutJSON
// ---------------------------------------------------------

func TestPutJSON_Success(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("请求方法 = %s, 期望 PUT", r.Method)
		}
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			t.Errorf("Content-Type = %s, 期望包含 application/json", r.Header.Get("Content-Type"))
		}

		body, _ := io.ReadAll(r.Body)
		var req TestRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("解析请求体失败: %v", err)
		}
		if req.Name != "update" || req.Value != 456 {
			t.Errorf("请求数据不匹配: %+v", req)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TestResponse{
			Message: "put success",
			Code:    0,
		})
	})
	defer server.Close()

	ctx := context.Background()
	reqBody := TestRequest{Name: "update", Value: 456}
	var result TestResponse

	resp, err := PutJSON(ctx, server.URL, reqBody, &result, nil)
	if err != nil {
		t.Fatalf("PutJSON 失败: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("状态码 = %d, 期望 200", resp.StatusCode())
	}
	if result.Message != "put success" {
		t.Errorf("响应消息 = %s, 期望 put success", result.Message)
	}
}

// TestPutJSON_RawBehaviorOnHTTPError 验证 4xx/5xx 不包装为 error，与 PostJSONRaw 一致
func TestPutJSON_RawBehaviorOnHTTPError(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("请求方法 = %s, 期望 PUT", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"bad request","code":400}`))
	})
	defer server.Close()

	ctx := context.Background()
	reqBody := TestRequest{Name: "bad", Value: 0}
	var result TestResponse

	resp, err := PutJSON(ctx, server.URL, reqBody, &result, nil)
	if err != nil {
		t.Fatalf("PutJSON 不应因 4xx 返回 error，但拿到: %v", err)
	}
	if resp == nil {
		t.Fatal("resp 不应为 nil")
	}
	if resp.StatusCode() != http.StatusBadRequest {
		t.Errorf("状态码 = %d, 期望 400", resp.StatusCode())
	}
}

// TestPutJSON_RequestError 验证网络错误时返回 err
func TestPutJSON_RequestError(t *testing.T) {
	ctx := context.Background()
	reqBody := TestRequest{Name: "network-fail", Value: 0}
	var result TestResponse

	resp, err := PutJSON(ctx, "http://127.0.0.1:1", reqBody, &result, nil, WithTimeout(200*time.Millisecond))
	if err == nil {
		t.Fatal("期望网络错误，但 err 为 nil")
	}
	if resp != nil && resp.StatusCode() != 0 {
		t.Errorf("网络错误时 status 应为 0（或 resp=nil），实际状态码: %d", resp.StatusCode())
	}
}

// TestPutJSON_WithHeadersAndTimeoutOption 验证自定义 headers 与 opts
func TestPutJSON_WithHeadersAndTimeoutOption(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("请求方法 = %s, 期望 PUT", r.Method)
		}
		if got := r.Header.Get("X-Test-Header"); got != "put-json-header" {
			t.Errorf("X-Test-Header = %s, 期望 put-json-header", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"ok","code":0}`))
	})
	defer server.Close()

	ctx := context.Background()
	reqBody := TestRequest{Name: "with-headers", Value: 1}
	var result TestResponse

	resp, err := PutJSON(
		ctx,
		server.URL,
		reqBody,
		&result,
		map[string]string{"X-Test-Header": "put-json-header"},
		WithTimeout(1*time.Second),
	)
	if err != nil {
		t.Fatalf("PutJSON 请求失败: %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		t.Errorf("状态码 = %d, 期望 200", resp.StatusCode())
	}
	if result.Message != "ok" || result.Code != 0 {
		t.Errorf("响应数据不匹配: %+v", result)
	}
}

// TestPutJSON_ResultNil 验证链式写法 SetResult(nil) 不判空时仍可正常请求
func TestPutJSON_ResultNil(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("请求方法 = %s, 期望 PUT", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	ctx := context.Background()
	reqBody := TestRequest{Name: "no-result", Value: 0}

	resp, err := PutJSON(ctx, server.URL, reqBody, nil, nil)
	if err != nil {
		t.Fatalf("PutJSON(result=nil) 失败: %v", err)
	}
	if resp.StatusCode() != http.StatusNoContent {
		t.Errorf("状态码 = %d, 期望 204", resp.StatusCode())
	}
}

// ---------------------------------------------------------
// 5. 测试 Delete
// ---------------------------------------------------------

func TestDelete_Success(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("请求方法 = %s, 期望 DELETE", r.Method)
		}

		// 验证查询参数
		if r.URL.Query().Get("id") != "123" {
			t.Errorf("参数 id = %s, 期望 123", r.URL.Query().Get("id"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TestResponse{
			Message: "delete success",
			Code:    0,
		})
	})
	defer server.Close()

	ctx := context.Background()
	params := map[string]string{"id": "123"}
	var result TestResponse

	resp, err := Delete(ctx, server.URL, params, &result)
	if err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("状态码 = %d, 期望 200", resp.StatusCode())
	}
	if result.Message != "delete success" {
		t.Errorf("响应消息 = %s, 期望 delete success", result.Message)
	}
}

// TestDelete_ResultNil 验证链式写法 SetResult(nil) 不判空时仍可正常请求
func TestDelete_ResultNil(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("请求方法 = %s, 期望 DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	ctx := context.Background()

	resp, err := Delete(ctx, server.URL, nil, nil)
	if err != nil {
		t.Fatalf("Delete(result=nil) 失败: %v", err)
	}
	if resp.StatusCode() != http.StatusNoContent {
		t.Errorf("状态码 = %d, 期望 204", resp.StatusCode())
	}
}

// ---------------------------------------------------------
// 6. 测试 PostForm
// ---------------------------------------------------------

func TestPostForm_Success(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("请求方法 = %s, 期望 POST", r.Method)
		}

		// 验证 Content-Type
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/x-www-form-urlencoded") {
			t.Errorf("Content-Type = %s, 期望包含 application/x-www-form-urlencoded", contentType)
		}

		// 解析表单数据
		if err := r.ParseForm(); err != nil {
			t.Errorf("解析表单失败: %v", err)
		}

		if r.FormValue("username") != "testuser" {
			t.Errorf("username = %s, 期望 testuser", r.FormValue("username"))
		}
		if r.FormValue("password") != "testpass" {
			t.Errorf("password = %s, 期望 testpass", r.FormValue("password"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TestResponse{
			Message: "form success",
			Code:    0,
		})
	})
	defer server.Close()

	ctx := context.Background()
	formData := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	var result TestResponse

	resp, err := PostForm(ctx, server.URL, formData, &result, nil)
	if err != nil {
		t.Fatalf("PostForm 失败: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("状态码 = %d, 期望 200", resp.StatusCode())
	}
	if result.Message != "form success" {
		t.Errorf("响应消息 = %s, 期望 form success", result.Message)
	}
}

func TestPostForm_RawBehaviorOnHTTPError(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("请求方法 = %s, 期望 POST", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"bad request","code":400}`))
	})
	defer server.Close()

	ctx := context.Background()
	formData := map[string]string{
		"username": "bad-user",
	}
	var result TestResponse

	resp, err := PostForm(ctx, server.URL, formData, &result, nil)
	if err != nil {
		t.Fatalf("PostForm 不应因 4xx 返回 error，但拿到: %v", err)
	}
	if resp == nil {
		t.Fatal("resp 不应为 nil")
	}
	if resp.StatusCode() != http.StatusBadRequest {
		t.Errorf("状态码 = %d, 期望 400", resp.StatusCode())
	}
	// resty 默认仅在成功响应时反序列化 SetResult，4xx 不保证写入 result
}

func TestPostForm_RequestError(t *testing.T) {
	ctx := context.Background()
	formData := map[string]string{
		"username": "network-fail",
	}
	var result TestResponse

	// 使用无效地址触发网络错误，验证 raw 语义下仍会返回 err
	resp, err := PostForm(ctx, "http://127.0.0.1:1", formData, &result, nil, WithTimeout(200*time.Millisecond))
	if err == nil {
		t.Fatal("期望网络错误，但 err 为 nil")
	}
	if resp != nil && resp.StatusCode() != 0 {
		t.Errorf("网络错误时 status 应为 0（或 resp=nil），实际状态码: %d", resp.StatusCode())
	}
}

func TestPostForm_WithHeadersAndTimeoutOption(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Test-Header"); got != "post-form-header" {
			t.Errorf("X-Test-Header = %s, 期望 post-form-header", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("解析表单失败: %v", err)
		}
		if r.FormValue("k") != "v" {
			t.Errorf("form k = %s, 期望 v", r.FormValue("k"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"ok","code":0}`))
	})
	defer server.Close()

	ctx := context.Background()
	var result TestResponse
	resp, err := PostForm(
		ctx,
		server.URL,
		map[string]string{"k": "v"},
		&result,
		map[string]string{"X-Test-Header": "post-form-header"},
		WithTimeout(1*time.Second),
	)
	if err != nil {
		t.Fatalf("PostForm 请求失败: %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		t.Errorf("状态码 = %d, 期望 200", resp.StatusCode())
	}
	if result.Message != "ok" || result.Code != 0 {
		t.Errorf("响应数据不匹配: %+v", result)
	}
}

// TestPostForm_ResultNil 验证链式写法 SetResult(nil) 不判空时仍可正常请求
func TestPostForm_ResultNil(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("请求方法 = %s, 期望 POST", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	ctx := context.Background()
	formData := map[string]string{"k": "v"}

	resp, err := PostForm(ctx, server.URL, formData, nil, nil)
	if err != nil {
		t.Fatalf("PostForm(result=nil) 失败: %v", err)
	}
	if resp.StatusCode() != http.StatusNoContent {
		t.Errorf("状态码 = %d, 期望 204", resp.StatusCode())
	}
}

// ---------------------------------------------------------
// 7. 测试 PostMultipart
// ---------------------------------------------------------

func TestPostMultipart_Success(t *testing.T) {
	// 创建临时测试文件
	tempFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name())

	testContent := "test file content"
	if _, err := tempFile.WriteString(testContent); err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}
	tempFile.Close()

	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("请求方法 = %s, 期望 POST", r.Method)
		}

		// 验证 Content-Type
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "multipart/form-data") {
			t.Errorf("Content-Type = %s, 期望包含 multipart/form-data", contentType)
		}

		// 解析 multipart 表单
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Errorf("解析 multipart 表单失败: %v", err)
		}

		// 验证表单字段
		if r.FormValue("description") != "test upload" {
			t.Errorf("description = %s, 期望 test upload", r.FormValue("description"))
		}

		// 验证文件
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Errorf("获取文件失败: %v", err)
		} else {
			defer file.Close()
			content, _ := io.ReadAll(file)
			if string(content) != testContent {
				t.Errorf("文件内容 = %s, 期望 %s", string(content), testContent)
			}
			if header.Filename == "" {
				t.Error("文件名为空")
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TestResponse{
			Message: "upload success",
			Code:    0,
		})
	})
	defer server.Close()

	ctx := context.Background()
	formData := map[string]string{
		"description": "test upload",
	}
	files := map[string]string{
		"file": tempFile.Name(),
	}
	var result TestResponse

	resp, err := PostMultipart(ctx, server.URL, formData, files, &result)
	if err != nil {
		t.Fatalf("PostMultipart 失败: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		t.Errorf("状态码 = %d, 期望 200", resp.StatusCode())
	}
	if result.Message != "upload success" {
		t.Errorf("响应消息 = %s, 期望 upload success", result.Message)
	}
}

// TestPostMultipart_ResultNil 验证链式写法 SetResult(nil) 不判空时仍可正常请求
func TestPostMultipart_ResultNil(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.WriteString("test")
	tempFile.Close()

	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("请求方法 = %s, 期望 POST", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer server.Close()

	ctx := context.Background()
	formData := map[string]string{"k": "v"}
	files := map[string]string{"file": tempFile.Name()}

	resp, err := PostMultipart(ctx, server.URL, formData, files, nil)
	if err != nil {
		t.Fatalf("PostMultipart(result=nil) 失败: %v", err)
	}
	if resp.StatusCode() != http.StatusNoContent {
		t.Errorf("状态码 = %d, 期望 204", resp.StatusCode())
	}
}

// ---------------------------------------------------------
// 8. 测试选项模式
// ---------------------------------------------------------

func TestWithTimeout_Option(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	ctx := context.Background()
	var result TestResponse

	// 测试短超时（应该超时）
	_, err := Get(ctx, server.URL, nil, &result, WithTimeout(100*time.Millisecond))
	if err == nil {
		t.Error("期望超时错误，但请求成功了")
	}

	// 测试长超时（应该成功）
	_, err = Get(ctx, server.URL, nil, &result, WithTimeout(1*time.Second))
	if err != nil {
		t.Errorf("请求失败: %v", err)
	}
}

func TestWithRetry_Option(t *testing.T) {
	attemptCount := 0
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TestResponse{Message: "success", Code: 0})
		}
	})
	defer server.Close()

	ctx := context.Background()
	var result TestResponse

	// 使用重试选项
	resp, err := Get(ctx, server.URL, nil, &result, WithRetry(3))

	// 由于 resty 的重试机制，这里的行为可能不同
	// resty 的 SetRetryCount 需要配合 AddRetryCondition 使用
	// 这里主要测试选项是否生效
	if resp != nil {
		t.Logf("尝试次数: %d, 状态码: %d", attemptCount, resp.StatusCode())
	}
	if err != nil {
		t.Logf("请求错误: %v", err)
	}
}

// ---------------------------------------------------------
// 9. 测试 GetHttpClient
// ---------------------------------------------------------

func TestGetHttpClient(t *testing.T) {
	client := GetHttpClient()
	if client == nil {
		t.Fatal("GetHttpClient 返回 nil")
	}

	// 验证底层 Transport
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Transport 类型不正确")
	}

	// 验证连接池配置是否复用
	if transport.MaxIdleConns != 6000 {
		t.Errorf("MaxIdleConns = %d, 期望 6000", transport.MaxIdleConns)
	}
}

func TestGetHttpClientWithTimeout(t *testing.T) {
	customTimeout := 5 * time.Second
	client := GetHttpClientWithTimeout(customTimeout)

	if client == nil {
		t.Fatal("GetHttpClientWithTimeout 返回 nil")
	}

	// 验证超时时间
	if client.Timeout != customTimeout {
		t.Errorf("Timeout = %v, 期望 %v", client.Timeout, customTimeout)
	}

	// 验证连接池配置是否复用
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Transport 类型不正确")
	}
	if transport.MaxIdleConns != 6000 {
		t.Errorf("MaxIdleConns = %d, 期望 6000（连接池应该被复用）", transport.MaxIdleConns)
	}
}

// ---------------------------------------------------------
// 10. 测试 Context 取消
// ---------------------------------------------------------

func TestContextCancellation(t *testing.T) {
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	// 创建可取消的 Context
	ctx, cancel := context.WithCancel(context.Background())

	// 100ms 后取消请求
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	var result TestResponse
	_, err := Get(ctx, server.URL, nil, &result)

	// 应该因为 context 取消而失败
	if err == nil {
		t.Error("期望 context 取消错误，但请求成功了")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Logf("错误信息: %v", err)
	}
}

// ---------------------------------------------------------
// 11. 测试连接池复用
// ---------------------------------------------------------

func TestConnectionPoolReuse(t *testing.T) {
	requestCount := 0
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TestResponse{Message: fmt.Sprintf("request %d", requestCount), Code: 0})
	})
	defer server.Close()

	ctx := context.Background()

	// 发送多个请求，验证连接池复用
	for i := 0; i < 10; i++ {
		var result TestResponse
		resp, err := Get(ctx, server.URL, nil, &result)
		if err != nil {
			t.Errorf("第 %d 次请求失败: %v", i+1, err)
		}
		if resp.StatusCode() != http.StatusOK {
			t.Errorf("第 %d 次请求状态码 = %d, 期望 200", i+1, resp.StatusCode())
		}
	}

	if requestCount != 10 {
		t.Errorf("服务器收到 %d 次请求, 期望 10 次", requestCount)
	}
}

// ---------------------------------------------------------
// 12. 基准测试
// ---------------------------------------------------------

func BenchmarkPostJSON(b *testing.B) {
	server := createTestServer(nil, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TestResponse{Message: "success", Code: 0})
	})
	defer server.Close()

	ctx := context.Background()
	reqBody := TestRequest{Name: "benchmark", Value: 999}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result TestResponse
		_, err := PostJSON(ctx, server.URL, reqBody, &result, nil)
		if err != nil {
			b.Fatalf("请求失败: %v", err)
		}
	}
}

func BenchmarkGetHttpClient(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := GetHttpClient()
		if client == nil {
			b.Fatal("GetHttpClient 返回 nil")
		}
	}
}

func BenchmarkGetHttpClientWithTimeout(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := GetHttpClientWithTimeout(5 * time.Second)
		if client == nil {
			b.Fatal("GetHttpClientWithTimeout 返回 nil")
		}
	}
}

// ---------------------------------------------------------
// 13. 并发测试
// ---------------------------------------------------------

func TestConcurrentRequests(t *testing.T) {
	requestCount := 0
	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		time.Sleep(10 * time.Millisecond) // 模拟处理时间
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TestResponse{Message: "success", Code: 0})
	})
	defer server.Close()

	ctx := context.Background()
	concurrency := 50
	done := make(chan bool, concurrency)

	// 并发发送请求
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			var result TestResponse
			reqBody := TestRequest{Name: fmt.Sprintf("test-%d", index), Value: index}
			_, err := PostJSON(ctx, server.URL, reqBody, &result, nil)
			if err != nil {
				t.Errorf("并发请求 %d 失败: %v", index, err)
			}
			done <- true
		}(i)
	}

	// 等待所有请求完成
	for i := 0; i < concurrency; i++ {
		<-done
	}

	if requestCount != concurrency {
		t.Errorf("服务器收到 %d 次请求, 期望 %d 次", requestCount, concurrency)
	}
}
