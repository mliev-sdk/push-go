package mlievpush

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Client 消息推送客户端
type Client struct {
	baseURL    string       // 基础URL
	appID      string       // 应用ID
	appSecret  string       // 应用密钥
	httpClient *http.Client // HTTP客户端
}

// ClientOption 客户端配置选项
type ClientOption func(*Client)

// WithHTTPClient 设置自定义HTTP客户端
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTimeout 设置请求超时时间
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient 创建消息推送客户端
func NewClient(baseURL, appID, appSecret string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL:   baseURL,
		appID:     appID,
		appSecret: appSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// 应用配置选项
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// doRequest 执行HTTP请求
func (c *Client) doRequest(ctx context.Context, method, path string, reqData interface{}) (*Response, error) {
	// 生成时间戳和随机数
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	nonce := uuid.New().String()

	// 构建请求体和参数map（用于签名）
	var bodyBytes []byte
	var params map[string]interface{}

	if reqData != nil {
		// 序列化请求数据
		var err error
		bodyBytes, err = json.Marshal(reqData)
		if err != nil {
			return nil, fmt.Errorf("marshal request data: %w", err)
		}

		// 将请求数据转换为map（用于签名）
		if err := json.Unmarshal(bodyBytes, &params); err != nil {
			return nil, fmt.Errorf("unmarshal request data to map: %w", err)
		}
	}

	// 生成签名
	signature := generateSignature(method, path, params, timestamp, nonce, c.appSecret)

	// 构建HTTP请求
	url := c.baseURL + path
	var body io.Reader
	if len(bodyBytes) > 0 {
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// 设置请求头
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-App-Id", c.appID)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Nonce", nonce)
	req.Header.Set("X-Signature", signature)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	// 解析响应
	var result Response
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// 检查业务错误
	if result.Code != 0 {
		return &result, NewAPIError(result.Code, result.Message)
	}

	return &result, nil
}

// SendMessage 发送单条消息
func (c *Client) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageData, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/api/v1/messages", req)
	if err != nil {
		return nil, err
	}

	var data SendMessageData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal response data: %w", err)
	}

	return &data, nil
}

// SendBatch 批量发送消息
func (c *Client) SendBatch(ctx context.Context, req *SendBatchRequest) (*SendBatchData, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/api/v1/messages/batch", req)
	if err != nil {
		return nil, err
	}

	var data SendBatchData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal response data: %w", err)
	}

	return &data, nil
}

// QueryTask 查询任务状态
func (c *Client) QueryTask(ctx context.Context, taskID string) (*QueryTaskData, error) {
	path := "/api/v1/messages/" + taskID
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var data QueryTaskData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal response data: %w", err)
	}

	return &data, nil
}
