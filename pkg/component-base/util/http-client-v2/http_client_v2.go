package http_client_v2

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

var globalClient *resty.Client

func init() {
	// 1. 初始化全局 Client
	// 使用完整的连接池配置，与 httpclient.InitTransport() 保持一致
	globalClient = resty.New().
		SetTimeout(10 * time.Second). // 默认超时 10s
		SetRetryCount(0).             // 默认不重试
		//SetHeader("Content-Type", "application/json").
		// 关键：显式配置连接池，优化支付场景下的并发性能
		SetTransport(&http.Transport{
			// 拨号上下文：控制建立 TCP 连接的超时时间
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,  // 连接超时：5秒连不上直接报错
				KeepAlive: 30 * time.Second, // 长连接保活探测间隔
			}).DialContext,
			// TLS 握手超时 (HTTPS)
			TLSHandshakeTimeout: 5 * time.Second,
			// 连接池核心配置
			// 🎯 关键配置：100 渠道均衡方案
			MaxIdleConns:        6000,             // 所有 Host 的最大空闲连接总数，总连接池 6000
			MaxIdleConnsPerHost: 80,               // 每个 Host (域名) 保持的最大空闲连接数， 每渠道 80
			IdleConnTimeout:     90 * time.Second, // 空闲连接的存活时间：如果连接空闲超过 90秒，自动关闭

			MaxConnsPerHost:       0,
			DisableKeepAlives:     false,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		})
}

// Option 定义配置选项函数
type Option func(*resty.Client)

// WithTimeout 选项：自定义超时时间
func WithTimeout(d time.Duration) Option {
	return func(c *resty.Client) {
		c.SetTimeout(d)
	}
}

// WithRetry 选项：自定义重试次数
func WithRetry(count int) Option {
	return func(c *resty.Client) {
		c.SetRetryCount(count)
	}
}

// ---------------------------------------------------------
// 内部辅助函数：获取客户端
// ---------------------------------------------------------
func getClient(opts []Option) *resty.Client {
	client := globalClient
	if len(opts) > 0 {
		client = globalClient.Clone()
		for _, opt := range opts {
			opt(client)
		}
	}
	return client
}

// ---------------------------------------------------------
// 内部辅助函数：统一的响应处理与日志
// ---------------------------------------------------------
func handleResponse(resp *resty.Response, err error) (*resty.Response, error) {
	// 1. 网络层面的错误（DNS失败、超时等）
	if err != nil {
		return nil, err
	}

	// 2. 业务状态码检查
	if resp.StatusCode() > 399 {
		log.Printf("❌ [ReqUtil Error] Method: %s, URL: %s, Status: %d, Raw Body: %s",
			resp.Request.Method, resp.Request.URL, resp.StatusCode(), resp.String())
		return resp, fmt.Errorf("server error: status %d, body: %s", resp.StatusCode(), resp.String())
	} else {
		// 调试日志 (生产环境可考虑降低日志级别或截断长度)
		log.Printf("✅ [ReqUtil Success] Method: %s, URL: %s, Body: %s",
			resp.Request.Method, resp.Request.URL, resp.String())
	}

	return resp, nil
}

// ---------------------------------------------------------
// 核心方法实现
// ---------------------------------------------------------

// PostJSON 发送 POST 请求 (JSON Body)
func PostJSON(ctx context.Context, url string, body interface{}, result interface{}, headers map[string]string, opts ...Option) (*resty.Response, error) {
	client := globalClient
	// 应用可选配置（如自定义超时、重试）
	if len(opts) > 0 {
		client = client.Clone()
		for _, opt := range opts {
			opt(client)
		}
	}
	//headers
	req := client.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(result)

	req.SetHeader("Content-Type", "application/json")

	if len(headers) > 0 {
		req.SetHeaders(headers)
	}

	resp, err := req.Post(url)
	return handleResponse(resp, err)
}

// PostJSONRaw 发送 POST 请求 (JSON Body)，但**不**根据 HTTP 状态码包装为错误。
// 适用于需要业务侧自行根据 StatusCode / Body 做复杂判断的场景：
//   - err 仅表示网络层错误（超时、DNS失败、连接失败等）
//   - 无论 2xx / 4xx / 5xx，都会返回 resp（除非网络错误导致 resp 为 nil）
func PostJSONRaw(ctx context.Context, url string, body interface{}, result interface{}, headers map[string]string, opts ...Option) (*resty.Response, error) {
	client := globalClient
	// 应用可选配置（如自定义超时、重试）
	if len(opts) > 0 {
		client = client.Clone()
		for _, opt := range opts {
			opt(client)
		}
	}

	req := client.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(result).
		SetHeader("Content-Type", "application/json")

	if len(headers) > 0 {
		req.SetHeaders(headers)
	}

	// 与 PostJSON 的区别：不调用 handleResponse，不检查 StatusCode
	return req.Post(url)
}

// PutJSON 发送 PUT 请求 (通常用于更新)
// 与 PostJSONRaw 语义一致：不调用 handleResponse，不检查 StatusCode
//   - err 仅表示网络层错误（超时、DNS失败、连接失败等）
//   - 无论 2xx / 4xx / 5xx，都会返回 resp（除非网络错误导致 resp 为 nil）
func PutJSON(ctx context.Context, url string, body interface{}, result interface{}, headers map[string]string, opts ...Option) (*resty.Response, error) {
	client := globalClient
	if len(opts) > 0 {
		client = client.Clone()
		for _, opt := range opts {
			opt(client)
		}
	}

	req := client.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(result).
		SetHeader("Content-Type", "application/json")

	if len(headers) > 0 {
		req.SetHeaders(headers)
	}

	return req.Put(url)
}

// Get 发送 GET 请求
// 不调用 handleResponse，不检查 StatusCode
func Get(ctx context.Context, url string, params map[string]string, result interface{}, opts ...Option) (*resty.Response, error) {
	client := getClient(opts)
	resp, err := client.R().
		SetContext(ctx).
		SetQueryParams(params).
		SetResult(result).
		Get(url)
	return resp, err
}

// Delete 发送 DELETE 请求
// 不调用 handleResponse，不检查 StatusCode
func Delete(ctx context.Context, url string, params map[string]string, result interface{}, opts ...Option) (*resty.Response, error) {
	client := getClient(opts)
	req := client.R().
		SetContext(ctx).
		SetResult(result)

	// DELETE 请求有时也带 Query Params
	if len(params) > 0 {
		req.SetQueryParams(params)
	}

	resp, err := req.Delete(url)
	return resp, err
}

// PostForm 发送表单请求 (application/x-www-form-urlencoded)
// 注意：formData 通常是 map[string]string
func PostForm(ctx context.Context, url string, formData map[string]string, result interface{}, headers map[string]string, opts ...Option) (*resty.Response, error) {
	client := globalClient

	req := client.R().
		SetContext(ctx).
		SetFormData(formData).
		SetResult(result).
		SetHeader("Content-Type", "application/x-www-form-urlencoded")

	if len(headers) > 0 {
		req.SetHeaders(headers)
	}

	// 与 PostJSON 的区别：不调用 handleResponse，不检查 StatusCode
	return req.Post(url)

}

// PostMultipart 发送文件上传请求 (multipart/form-data)
// files 参数结构：map["表单字段名"]"文件路径"
// 不调用 handleResponse，不检查 StatusCode
func PostMultipart(ctx context.Context, url string, formData map[string]string, files map[string]string, result interface{}, opts ...Option) (*resty.Response, error) {
	client := getClient(opts)
	req := client.R().
		SetContext(ctx).
		SetFormData(formData).
		SetResult(result)

	// 添加文件
	if len(files) > 0 {
		req.SetFiles(files)
	}

	resp, err := req.Post(url)
	return resp, err
}

// ---------------------------------------------------------
// 获取底层 HTTP 客户端（供业务代码直接使用 net/http）
// ---------------------------------------------------------

// GetHttpClient 获取全局 resty 客户端的底层 *http.Client
// 供需要直接使用 net/http 的业务代码使用，复用连接池
// 默认超时时间为 10 秒
func GetHttpClient() *http.Client {
	return globalClient.GetClient()
}

// GetHttpClientWithTimeout 获取带自定义超时的 HTTP 客户端
// 通过 resty 客户端克隆并设置超时，然后获取底层 http.Client
// 这样既能复用连接池，又能自定义超时时间
func GetHttpClientWithTimeout(timeout time.Duration) *http.Client {
	// 使用 resty 的 Clone 和 SetTimeout，然后获取底层 http.Client
	// 这样 Transport（连接池）会被复用，只有超时时间改变
	clonedClient := globalClient.Clone().SetTimeout(timeout)
	return clonedClient.GetClient()
}

func GetRestyHttpClientWithTimeout(timeout time.Duration) *resty.Client {
	// 使用 resty 的 Clone 和 SetTimeout，然后获取底层 http.Client
	// 这样 Transport（连接池）会被复用，只有超时时间改变
	clonedClient := globalClient.Clone().SetTimeout(timeout)
	return clonedClient
}
func GetRestyHttpClient() *resty.Client {
	clonedClient := globalClient.Clone()
	return clonedClient
}
