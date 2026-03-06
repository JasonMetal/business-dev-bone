package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

func GetJSON(url string, headers map[string]string) ([]byte, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/json"
	return Get(url, headers)
}

func PostJSON(url string, headers map[string]string, body any) ([]byte, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/json"

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return Post(url, headers, b)
}

// TODO 需要考虑超时时间/及重试
func Get(url string, headers map[string]string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	body, err := defaultHTTPClient.DoRequest(ctx, "GET", url, headers, nil)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func Post(url string, headers map[string]string, body []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	body, err := defaultHTTPClient.DoRequest(ctx, "POST", url, headers, io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return nil, err
	}
	return body, nil
}

var defaultHTTPClient = NewHttpClient(10, 2, 30*time.Second, 30*time.Second, 30*time.Second, nil)

type HttpClient struct {
	client *http.Client
}

func NewHttpClient(maxIdleConns, maxIdleConnsPerHost int, idleConnTimeout, timeout, keepAlive time.Duration, proxyAddr *string) *HttpClient {
	transport := &http.Transport{
		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,
		IdleConnTimeout:     idleConnTimeout,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: keepAlive,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	if proxyAddr != nil && *proxyAddr != "" {
		proxyURL, _ := url.Parse(*proxyAddr)
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	client := &http.Client{Transport: transport}
	return &HttpClient{client: client}
}

func (h HttpClient) DoRequest(ctx context.Context, method, url string, header map[string]string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	for k, v := range header {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpcted status code %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}
