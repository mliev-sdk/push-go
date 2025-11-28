# mliev-push-go

消息推送服务的 Go SDK，支持短信、邮件、企业微信、钉钉等多种消息类型。

## 特性

- ✅ 完整的 API 支持（发送单条、批量发送、查询状态）
- ✅ HMAC-SHA256 签名认证
- ✅ Context 支持（超时、取消）
- ✅ 完善的错误处理
- ✅ 并发安全
- ✅ 单元测试覆盖

## 安装

```bash
go get github.com/mliev-sdk/push-go
```

## 快速开始

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/mliev-sdk/push-go"
)

func main() {
    // 创建客户端
    client := mlievpush.NewClient(
        "https://your-domain.com",  // 基础URL
        "your_app_id",              // 应用ID
        "your_app_secret",          // 应用密钥
    )

    // 发送短信
    ctx := context.Background()
    data, err := client.SendMessage(ctx, &mlievpush.SendMessageRequest{
        ChannelID:     1,
        SignatureName: "【您的签名】",
        Receiver:      "13800138000",
        TemplateParams: map[string]interface{}{
            "code": "123456",
        },
    })

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("发送成功，任务ID: %s\n", data.TaskID)
}
```

## API 文档

### 创建客户端

```go
// 基础创建
client := mlievpush.NewClient(baseURL, appID, appSecret)

// 使用选项配置
client := mlievpush.NewClient(
    baseURL,
    appID,
    appSecret,
    mlievpush.WithTimeout(15*time.Second),      // 设置超时
    mlievpush.WithHTTPClient(customHTTPClient), // 自定义HTTP客户端
)
```

### 发送单条消息

发送消息到单个接收者。

```go
req := &mlievpush.SendMessageRequest{
    ChannelID:     1,                          // 通道ID（必填）
    SignatureName: "【您的签名】",              // 签名名称（必填）
    Receiver:      "13800138000",              // 接收者（必填）
    TemplateParams: map[string]interface{}{    // 模板参数（可选）
        "code":        "123456",
        "expire_time": "5",
    },
    ScheduledAt: "2025-11-26T10:00:00Z",      // 定时发送（可选）
}

data, err := client.SendMessage(ctx, req)
if err != nil {
    // 处理错误
}

fmt.Printf("任务ID: %s\n", data.TaskID)
fmt.Printf("状态: %s\n", data.Status)
```

### 批量发送消息

批量发送消息到多个接收者（共用相同的模板参数）。

```go
req := &mlievpush.SendBatchRequest{
    ChannelID:     1,
    SignatureName: "【您的签名】",
    Receivers: []string{
        "13800138000",
        "13800138001",
        "13800138002",
    },
    TemplateParams: map[string]interface{}{
        "content":  "系统维护通知",
        "duration": "2小时",
    },
}

data, err := client.SendBatch(ctx, req)
if err != nil {
    // 处理错误
}

fmt.Printf("批次ID: %s\n", data.BatchID)
fmt.Printf("成功: %d, 失败: %d\n", data.SuccessCount, data.FailedCount)
```

### 查询任务状态

根据任务 ID 查询发送状态。

```go
taskID := "550e8400-e29b-41d4-a716-446655440000"

data, err := client.QueryTask(ctx, taskID)
if err != nil {
    // 处理错误
}

fmt.Printf("状态: %s\n", data.Status)
fmt.Printf("消息类型: %s\n", data.MessageType)
fmt.Printf("接收者: %s\n", data.Receiver)
fmt.Printf("内容: %s\n", data.Content)
```

## 错误处理

SDK 提供了完善的错误处理机制。

### 判断 API 错误

```go
data, err := client.SendMessage(ctx, req)
if err != nil {
    if mlievpush.IsAPIError(err) {
        // API 返回的业务错误
        apiErr := err.(*mlievpush.APIError)
        fmt.Printf("错误码: %d\n", apiErr.Code)
        fmt.Printf("错误信息: %s\n", apiErr.Message)
    } else {
        // 网络错误或其他错误
        fmt.Printf("请求失败: %v\n", err)
    }
}
```

### 错误码处理

```go
if mlievpush.IsAPIError(err) {
    apiErr := err.(*mlievpush.APIError)

    switch apiErr.Code {
    case mlievpush.ErrCodeInvalidSignature:
        // 签名验证失败
        fmt.Println("请检查 app_secret 是否正确")

    case mlievpush.ErrCodeChannelNotFound:
        // 通道不存在
        fmt.Println("请检查 channel_id 是否正确")

    case mlievpush.ErrCodeRateLimitExceeded:
        // 超出速率限制
        fmt.Println("请降低请求频率")

    default:
        // 其他错误
        fmt.Printf("%s\n", mlievpush.GetErrorMessage(apiErr.Code))
    }
}
```

### 常见错误码

| 错误码 | 常量 | 说明 |
|--------|------|------|
| 20003 | `ErrCodeInvalidSignature` | 签名验证失败 |
| 20004 | `ErrCodeInvalidTimestamp` | 时间戳无效 |
| 30001 | `ErrCodeRateLimitExceeded` | 超出速率限制 |
| 30003 | `ErrCodeChannelNotFound` | 通道不存在 |
| 30007 | `ErrCodeTaskNotFound` | 任务不存在 |

完整错误码列表请参考 [API 文档](doc/API_INTEGRATION.md#错误码参考)。

## Context 支持

所有 API 方法都支持 Context，可以用于超时控制和请求取消。

### 超时控制

```go
// 创建一个 5 秒超时的 context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

req := &mlievpush.SendMessageRequest{
    ChannelID:     1,
    SignatureName: "【您的签名】",
    Receiver:      "13800138000",
}

data, err := client.SendMessage(ctx, req)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        fmt.Println("请求超时")
    }
}
```

### 请求取消

```go
ctx, cancel := context.WithCancel(context.Background())

// 在另一个 goroutine 中可以取消请求
go func() {
    time.Sleep(1 * time.Second)
    cancel()
}()

req := &mlievpush.SendMessageRequest{
    ChannelID:     1,
    SignatureName: "【您的签名】",
    Receiver:      "13800138000",
}

data, err := client.SendMessage(ctx, req)
if err != nil {
    if ctx.Err() == context.Canceled {
        fmt.Println("请求已取消")
    }
}
```

## 常量定义

### 任务状态

```go
mlievpush.TaskStatusPending     // "pending" - 待处理
mlievpush.TaskStatusProcessing  // "processing" - 处理中
mlievpush.TaskStatusSuccess     // "success" - 成功
mlievpush.TaskStatusFailed      // "failed" - 失败
```

### 消息类型

```go
mlievpush.MessageTypeSMS         // "sms" - 短信
mlievpush.MessageTypeEmail       // "email" - 邮件
mlievpush.MessageTypeWechatWork  // "wechat_work" - 企业微信
mlievpush.MessageTypeDingtalk    // "dingtalk" - 钉钉
mlievpush.MessageTypeWebhook     // "webhook" - Webhook
mlievpush.MessageTypePush        // "push" - 推送通知
```

### 回调状态

```go
mlievpush.CallbackStatusDelivered  // "delivered" - 已送达
mlievpush.CallbackStatusFailed     // "failed" - 发送失败
mlievpush.CallbackStatusRejected   // "rejected" - 被拒绝
```

## 完整示例

查看 [examples/main.go](examples/main.go) 获取更多使用示例：

- 发送单条消息
- 批量发送消息
- 查询任务状态
- 错误处理
- Context 超时控制

运行示例：

```bash
cd examples
go run main.go
```

## 测试

运行单元测试：

```bash
go test -v
```

运行测试并查看覆盖率：

```bash
go test -v -cover
```

## 最佳实践

1. **重用客户端实例**：`Client` 是并发安全的，可以在多个 goroutine 中共享使用
2. **使用 Context**：为每个请求设置合理的超时时间，避免长时间阻塞
3. **错误处理**：区分 API 错误和网络错误，进行针对性处理
4. **日志记录**：保存返回的 `task_id` 便于问题排查
5. **批量限制**：单次批量发送建议不超过 500 条

## 依赖

- Go 1.21+
- github.com/google/uuid v1.5.0

## 许可证

MIT License

## 支持

如有问题或建议，请提交 Issue。