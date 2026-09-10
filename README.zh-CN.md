# mliev-push-go

消息推送服务的 Go SDK，支持短信、邮件、企业微信、钉钉等多种消息类型。

## 特性

- ✅ 完整的 API 支持（发送单条、批量发送、查询状态、通道目录）
- ✅ HMAC-SHA256 签名认证
- ✅ Context 支持（超时、取消）
- ✅ 从本地文件或内存内容创建邮件附件
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
        TemplateParams: map[string]string{
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
    SignatureName: "【您的签名】",              // SignatureNames 中的别名；SignatureRequired=true 时必填
    Receiver:      "13800138000",              // 接收者（必填）
    TemplateParams: map[string]string{    // 模板参数（可选）
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
    TemplateParams: map[string]string{
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

### 邮件附件

邮件通道的单发和批量发送都支持附件。可以直接从本地文件创建：

```go
attachment, err := mlievpush.NewEmailAttachmentFromFile("./invoice.pdf")
if err != nil {
    log.Fatal(err)
}

data, err := client.SendMessage(ctx, &mlievpush.SendMessageRequest{
    ChannelID:     12,
    SignatureName: "invoice-ready", // 邮件标题别名
    Receiver:      "customer@example.com",
    TemplateParams: map[string]string{
        "order_id": "ORDER-1001",
    },
    Attachments: []mlievpush.EmailAttachment{attachment},
})
```

对于运行时生成的内容，使用 `NewEmailAttachment`：

```go
attachment, err := mlievpush.NewEmailAttachment("report.csv", []byte("name,total\nAlice,42\n"))
if err != nil {
    log.Fatal(err)
}
attachment.ContentType = "text/csv; charset=utf-8" // 可选覆盖
```

批量请求复用同一个附件切片，每个收件人都会收到这些附件：

```go
batchRequest.Attachments = []mlievpush.EmailAttachment{attachment}
data, err := client.SendBatch(ctx, batchRequest)
```

内容已经编码时也可以直接构造 `EmailAttachment`。`ContentBase64` 必须是标准 RFC 4648 Base64，不能包含 `data:` URL 前缀。批量发送的所有收件人共用同一组附件。服务端默认限制为最多 5 个、单个 5 MiB、合计 10 MiB，但部署方可以修改，因此 SDK 不硬编码容量限制。本地文件助手会将整个文件读入内存。

可运行的[邮件附件示例](examples/attachment/main.go)展示了完整调用方式。

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

### 通道目录与可视化接入

通过通道目录构建业务系统的通道、模板、变量和签名选择表单。服务端需提供 `GET /api/v1/channels` 与 `GET /api/v1/channels/{id}`。

```go
// 传 nil 或字段零值时使用服务端默认值：全部类型、第 1 页、每页 20 条。
page, err := client.ListChannels(ctx, &mlievpush.ListChannelsRequest{
    Type: mlievpush.MessageTypeSMS,
    Page: 1,
    PageSize: 20,
})
if err != nil {
    log.Fatal(err)
}
for _, channel := range page.Items {
    fmt.Println(channel.ID, channel.Name, channel.TemplateName, channel.Readiness.State)
}

// 使用用户从列表中选择的通道 ID。
detail, err := client.GetChannel(ctx, 42)
if err != nil {
    log.Fatal(err)
}
fmt.Println(detail.Readiness.State, detail.Readiness.BlockerCodes)
if detail.Template != nil {
    fmt.Println(detail.Template.ContentType, detail.Template.Content, detail.Template.Variables)
}
fmt.Println(detail.SignatureRequired, detail.SignatureNames)
```

- `ListChannelsData` 包含 `Items/Total/Page/Size`；空页保留非 nil 的空 `Items` 切片。每页最多 100 条，非零请求参数交由服务端校验。
- `ChannelData` 包含 `ID/Name/Type/MessageTemplateID/TemplateName/Readiness`；`ChannelDetailData` 另外包含 `Template/SignatureRequired/SignatureNames`。
- `ChannelTemplate` 包含 `ID/TemplateName/ContentType/Content/Variables/Description`，表示系统模板，实际投递内容由供应商模板决定。
- `ChannelReadinessReady` 和 `ChannelReadinessDegraded` 可选择；`ChannelReadinessBlocked` 也是成功的配置查询结果，应在界面置灰，并按需展示 `BlockerCodes`。
- `Template == nil` 表示模板缺失或被删除；`Template.Variables == nil` 表示变量配置无效，非 nil 的空切片表示没有变量，构建表单时应区分。
- `TemplateParams` 需包含返回的每个变量名，值均为字符串。`SignatureName` 从 `SignatureNames` 中选择，短信签名和邮件标题均使用此别名；`SignatureRequired` 为 true 时必填，否则可留空。
- 查询沿用四个认证头和现有限流，不消耗发送配额。查询参数编码到 URL，但不参与 HMAC 签名，GET 正文为空。SDK 原样返回不可用通道，不缓存或自动翻页。
- 发送时服务端会重新检查配置。目录接口错误保留在 `APIError` 中，错误码为 `400/404/500`；鉴权、限流仍可能使用 HTTP 200 携带非零业务码。

可运行的[通道目录示例](examples/catalog/main.go) 展示“列表 → 详情 → 填写变量 → 选择签名 → 发送”。在业务后端设置凭据：

```bash
export PUSH_BASE_URL='https://your-domain.com'
export PUSH_APP_ID='your_app_id'
export PUSH_APP_SECRET='your_app_secret'
# 只查询配置。
go run ./examples/catalog -channel 42
# 将 ID、签名别名、变量和接收者替换为实际选择。
go run ./examples/catalog -channel 42 -signature '验证码' -params '{"code":"123456","expire":"5"}' -receiver '13800138000' -send
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
