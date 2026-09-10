package mlievpush

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSendRequestsIncludeEmailAttachmentsInBodyAndSignature(t *testing.T) {
	attachment, err := NewEmailAttachment("报告.txt", []byte("hello"))
	if err != nil {
		t.Fatalf("NewEmailAttachment() error = %v", err)
	}
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		body, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			t.Errorf("read request body: %v", readErr)
		}
		var params map[string]interface{}
		if err := json.Unmarshal(body, &params); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		attachments, ok := params["attachments"].([]interface{})
		if !ok || len(attachments) != 1 {
			t.Errorf("attachments = %#v, want one item", params["attachments"])
		} else {
			item, _ := attachments[0].(map[string]interface{})
			if item["filename"] != "报告.txt" || item["content_base64"] != "aGVsbG8=" {
				t.Errorf("attachment body = %#v", item)
			}
		}
		wantSignature := generateSignature(
			r.Method,
			r.URL.Path,
			params,
			r.Header.Get("X-Timestamp"),
			r.Header.Get("X-Nonce"),
			"test_secret",
		)
		if got := r.Header.Get("X-Signature"); got != wantSignature {
			t.Errorf("X-Signature = %q, want %q", got, wantSignature)
		}

		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/messages/batch" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{"batch_id": "batch-1"},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"task_id": "task-1"},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test_app", "test_secret")
	if _, err := client.SendMessage(context.Background(), &SendMessageRequest{
		ChannelID:   1,
		Receiver:    "recipient@example.com",
		Attachments: []EmailAttachment{attachment},
	}); err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if _, err := client.SendBatch(context.Background(), &SendBatchRequest{
		ChannelID:   1,
		Receivers:   []string{"first@example.com", "second@example.com"},
		Attachments: []EmailAttachment{attachment},
	}); err != nil {
		t.Fatalf("SendBatch() error = %v", err)
	}
	if requestCount != 2 {
		t.Fatalf("request count = %d, want 2", requestCount)
	}
}

func TestSendMessageOmitsEmptyAttachments(t *testing.T) {
	request, err := json.Marshal(SendMessageRequest{ChannelID: 1, Receiver: "recipient@example.com"})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if string(request) == "" || bytes.Contains(request, []byte("attachments")) {
		t.Fatalf("request unexpectedly contains attachments: %s", request)
	}
}

// TestSortParams 测试参数排序功能
func TestSortParams(t *testing.T) {
	tests := []struct {
		name     string
		params   map[string]interface{}
		expected string
	}{
		{
			name:     "空参数",
			params:   nil,
			expected: "",
		},
		{
			name:     "空map",
			params:   map[string]interface{}{},
			expected: "",
		},
		{
			name: "单个参数",
			params: map[string]interface{}{
				"key": "value",
			},
			expected: `{"key":"value"}`,
		},
		{
			name: "多个参数排序",
			params: map[string]interface{}{
				"channel_id": 1,
				"receiver":   "13800138000",
				"template_params": map[string]interface{}{
					"code": "123456",
				},
			},
			expected: `{"channel_id":1,"receiver":"13800138000","template_params":{"code":"123456"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sortParams(tt.params)
			if result != tt.expected {
				t.Errorf("sortParams() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGenerateSignature 测试签名生成
func TestGenerateSignature(t *testing.T) {
	method := "POST"
	path := "/api/v1/messages"
	params := map[string]interface{}{
		"channel_id": 1,
		"receiver":   "13800138000",
		"template_params": map[string]interface{}{
			"code": "123456",
		},
	}
	timestamp := "1700000000"
	nonce := "abc123"
	appSecret := "secret123456"

	signature := generateSignature(method, path, params, timestamp, nonce, appSecret)

	// 签名应该是64个字符的十六进制字符串（SHA256的输出）
	if len(signature) != 64 {
		t.Errorf("signature length = %d, want 64", len(signature))
	}

	// 验证签名的一致性
	signature2 := generateSignature(method, path, params, timestamp, nonce, appSecret)
	if signature != signature2 {
		t.Error("signature should be deterministic")
	}

	// 不同参数应该生成不同的签名
	params2 := map[string]interface{}{
		"channel_id": 2,
		"receiver":   "13800138001",
	}
	signature3 := generateSignature(method, path, params2, timestamp, nonce, appSecret)
	if signature == signature3 {
		t.Error("different params should generate different signatures")
	}
}

// TestSendMessage 测试发送单条消息
func TestSendMessage(t *testing.T) {
	// 创建mock服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// 验证请求路径
		if r.URL.Path != "/api/v1/messages" {
			t.Errorf("expected /api/v1/messages, got %s", r.URL.Path)
		}

		// 验证请求头
		if r.Header.Get("X-App-Id") == "" {
			t.Error("missing X-App-Id header")
		}
		if r.Header.Get("X-Timestamp") == "" {
			t.Error("missing X-Timestamp header")
		}
		if r.Header.Get("X-Nonce") == "" {
			t.Error("missing X-Nonce header")
		}
		if r.Header.Get("X-Signature") == "" {
			t.Error("missing X-Signature header")
		}

		// 返回成功响应
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"data": map[string]interface{}{
				"task_id":    "550e8400-e29b-41d4-a716-446655440000",
				"status":     "pending",
				"created_at": "2025-11-25T10:00:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(server.URL, "test_app_id", "test_secret")

	// 发送消息
	ctx := context.Background()
	req := &SendMessageRequest{
		ChannelID:     1,
		SignatureName: "【测试签名】",
		Receiver:      "13800138000",
		TemplateParams: map[string]string{
			"code": "123456",
		},
	}

	data, err := client.SendMessage(ctx, req)
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	if data.TaskID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("TaskID = %v, want %v", data.TaskID, "550e8400-e29b-41d4-a716-446655440000")
	}
	if data.Status != "pending" {
		t.Errorf("Status = %v, want %v", data.Status, "pending")
	}
}

// TestSendBatch 测试批量发送消息
func TestSendBatch(t *testing.T) {
	// 创建mock服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法和路径
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/messages/batch" {
			t.Errorf("expected /api/v1/messages/batch, got %s", r.URL.Path)
		}

		// 返回成功响应
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"data": map[string]interface{}{
				"batch_id":      "660e8400-e29b-41d4-a716-446655440001",
				"total_count":   3,
				"success_count": 3,
				"failed_count":  0,
				"created_at":    "2025-11-25T10:00:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(server.URL, "test_app_id", "test_secret")

	// 批量发送消息
	ctx := context.Background()
	req := &SendBatchRequest{
		ChannelID:     1,
		SignatureName: "【测试签名】",
		Receivers:     []string{"13800138000", "13800138001", "13800138002"},
		TemplateParams: map[string]string{
			"content":  "系统维护通知",
			"duration": "2小时",
		},
	}

	data, err := client.SendBatch(ctx, req)
	if err != nil {
		t.Fatalf("SendBatch() error = %v", err)
	}

	if data.TotalCount != 3 {
		t.Errorf("TotalCount = %v, want %v", data.TotalCount, 3)
	}
	if data.SuccessCount != 3 {
		t.Errorf("SuccessCount = %v, want %v", data.SuccessCount, 3)
	}
}

// TestQueryTask 测试查询任务状态
func TestQueryTask(t *testing.T) {
	taskID := "550e8400-e29b-41d4-a716-446655440000"

	// 创建mock服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求方法和路径
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		expectedPath := "/api/v1/messages/" + taskID
		if r.URL.Path != expectedPath {
			t.Errorf("expected %s, got %s", expectedPath, r.URL.Path)
		}

		// 返回成功响应
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"data": map[string]interface{}{
				"id":              1,
				"task_id":         taskID,
				"app_id":          "test_app_001",
				"channel_id":      1,
				"message_type":    "sms",
				"receiver":        "13800138000",
				"content":         "您的验证码是123456，5分钟内有效。",
				"status":          "success",
				"callback_status": "delivered",
				"retry_count":     0,
				"max_retry":       3,
				"created_at":      "2025-11-25T10:00:00Z",
				"updated_at":      "2025-11-25T10:00:02Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(server.URL, "test_app_id", "test_secret")

	// 查询任务
	ctx := context.Background()
	data, err := client.QueryTask(ctx, taskID)
	if err != nil {
		t.Fatalf("QueryTask() error = %v", err)
	}

	if data.TaskID != taskID {
		t.Errorf("TaskID = %v, want %v", data.TaskID, taskID)
	}
	if data.Status != "success" {
		t.Errorf("Status = %v, want %v", data.Status, "success")
	}
	if data.MessageType != "sms" {
		t.Errorf("MessageType = %v, want %v", data.MessageType, "sms")
	}
}

// TestAPIError 测试API错误处理
func TestAPIError(t *testing.T) {
	// 创建mock服务器返回错误
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    20003,
			"message": "签名验证失败",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(server.URL, "test_app_id", "test_secret")

	// 发送消息（应该返回错误）
	ctx := context.Background()
	req := &SendMessageRequest{
		ChannelID:     1,
		SignatureName: "【测试签名】",
		Receiver:      "13800138000",
	}

	_, err := client.SendMessage(ctx, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// 验证是否为API错误
	if !IsAPIError(err) {
		t.Error("expected APIError")
	}

	apiErr := err.(*APIError)
	if apiErr.Code != 20003 {
		t.Errorf("error code = %v, want %v", apiErr.Code, 20003)
	}
}

// TestContextTimeout 测试Context超时
func TestContextTimeout(t *testing.T) {
	// 创建一个慢响应的mock服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(server.URL, "test_app_id", "test_secret")

	// 创建一个带超时的context（100ms）
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := &SendMessageRequest{
		ChannelID:     1,
		SignatureName: "【测试签名】",
		Receiver:      "13800138000",
	}

	// 应该超时
	_, err := client.SendMessage(ctx, req)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// TestContextCancellation 测试Context取消
func TestContextCancellation(t *testing.T) {
	// 创建一个慢响应的mock服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		resp := map[string]interface{}{
			"code":    0,
			"message": "success",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// 创建客户端
	client := NewClient(server.URL, "test_app_id", "test_secret")

	// 创建可取消的context
	ctx, cancel := context.WithCancel(context.Background())

	// 启动goroutine在100ms后取消
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	req := &SendMessageRequest{
		ChannelID:     1,
		SignatureName: "【测试签名】",
		Receiver:      "13800138000",
	}

	// 应该被取消
	_, err := client.SendMessage(ctx, req)
	if err == nil {
		t.Fatal("expected cancellation error, got nil")
	}
}
