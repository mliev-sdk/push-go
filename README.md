# mliev-push-go

Go SDK for message push service, supporting multiple message types including SMS, email, WeChat Work, DingTalk, and more.

## Features

- ✅ Complete API support (single send, batch send, status query, channel catalog)
- ✅ HMAC-SHA256 signature authentication
- ✅ Context support (timeout, cancellation)
- ✅ Comprehensive error handling
- ✅ Thread-safe
- ✅ Unit test coverage

## Installation

```bash
go get github.com/mliev-sdk/push-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/mliev-sdk/push-go"
)

func main() {
    // Create client
    client := mlievpush.NewClient(
        "https://your-domain.com",  // Base URL
        "your_app_id",              // App ID
        "your_app_secret",          // App secret
    )

    // Send SMS
    ctx := context.Background()
    data, err := client.SendMessage(ctx, &mlievpush.SendMessageRequest{
        ChannelID:     1,
        SignatureName: "【Your Signature】",
        Receiver:      "13800138000",
        TemplateParams: map[string]string{
            "code": "123456",
        },
    })

    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Sent successfully, Task ID: %s\n", data.TaskID)
}
```

## API Documentation

### Create Client

```go
// Basic creation
client := mlievpush.NewClient(baseURL, appID, appSecret)

// Using options
client := mlievpush.NewClient(
    baseURL,
    appID,
    appSecret,
    mlievpush.WithTimeout(15*time.Second),      // Set timeout
    mlievpush.WithHTTPClient(customHTTPClient), // Custom HTTP client
)
```

### Send Single Message

Send a message to a single recipient.

```go
req := &mlievpush.SendMessageRequest{
    ChannelID:     1,                          // Channel ID (required)
    SignatureName: "【Your Signature】",        // Alias from SignatureNames; required when SignatureRequired is true
    Receiver:      "13800138000",              // Recipient (required)
    TemplateParams: map[string]string{    // Template parameters (optional)
        "code":        "123456",
        "expire_time": "5",
    },
    ScheduledAt: "2025-11-26T10:00:00Z",      // Scheduled sending (optional)
}

data, err := client.SendMessage(ctx, req)
if err != nil {
    // Handle error
}

fmt.Printf("Task ID: %s\n", data.TaskID)
fmt.Printf("Status: %s\n", data.Status)
```

### Send Batch Messages

Send messages to multiple recipients (using the same template parameters).

```go
req := &mlievpush.SendBatchRequest{
    ChannelID:     1,
    SignatureName: "【Your Signature】",
    Receivers: []string{
        "13800138000",
        "13800138001",
        "13800138002",
    },
    TemplateParams: map[string]string{
        "content":  "System maintenance notification",
        "duration": "2 hours",
    },
}

data, err := client.SendBatch(ctx, req)
if err != nil {
    // Handle error
}

fmt.Printf("Batch ID: %s\n", data.BatchID)
fmt.Printf("Success: %d, Failed: %d\n", data.SuccessCount, data.FailedCount)
```

### Query Task Status

Query sending status by task ID.

```go
taskID := "550e8400-e29b-41d4-a716-446655440000"

data, err := client.QueryTask(ctx, taskID)
if err != nil {
    // Handle error
}

fmt.Printf("Status: %s\n", data.Status)
fmt.Printf("Message Type: %s\n", data.MessageType)
fmt.Printf("Recipient: %s\n", data.Receiver)
fmt.Printf("Content: %s\n", data.Content)
```

### Channel Catalog

Use the catalog to populate channel, template, variable, and signature controls in your business application. The server must provide `GET /api/v1/channels` and `GET /api/v1/channels/{id}`.

```go
// nil (or zero-valued fields) uses the server defaults: all types, page 1, size 20.
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

// Use the ID selected by the user from the channel list.
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

- `ListChannelsData` contains `Items`, `Total`, `Page`, and `Size`. Empty pages retain a non-nil empty `Items` slice. `PageSize` is limited to 100 by the server; nonzero request values are forwarded for server validation.
- `ChannelData` includes `ID`, `Name`, `Type`, `MessageTemplateID`, `TemplateName`, and `Readiness`. `ChannelDetailData` adds `Template`, `SignatureRequired`, and `SignatureNames`.
- `ChannelTemplate` includes `ID`, `TemplateName`, `ContentType`, `Content`, `Variables`, and `Description`. It is the system template, not the provider's final message.
- `ChannelReadinessReady` and `ChannelReadinessDegraded` are selectable; `ChannelReadinessBlocked` is a successful configuration response that your UI should disable. Display its `BlockerCodes` as needed.
- `Template == nil` means the template is missing/deleted. `Template.Variables == nil` means invalid configuration; a non-nil empty slice means no variables. Preserve that distinction when building a form.
- Populate every returned variable name in `TemplateParams` with a string value. Select `SignatureName` from `SignatureNames`; SMS signatures and email titles share this alias mechanism. Selection is required when `SignatureRequired` is true, and can be left empty otherwise.
- Queries use the existing four authentication headers and rate limit, without consuming sending quota. Query parameters are URL-encoded but excluded from the HMAC input; GET has an empty body. The SDK does not filter blocked channels, cache results, or fetch additional pages automatically.
- The server validates configuration again when sending. Catalog errors use `APIError` with codes `400`, `404`, or `500`; authentication and rate-limit errors can still arrive with HTTP 200 and a nonzero business code.

The runnable [catalog example](examples/catalog/main.go) covers list → detail → form values → signature selection → sending. Configure credentials in your backend environment:

```bash
export PUSH_BASE_URL='https://your-domain.com'
export PUSH_APP_ID='your_app_id'
export PUSH_APP_SECRET='your_app_secret'
# Read configuration only.
go run ./examples/catalog -channel 42
# Replace the ID, alias, variable names, and receiver with your actual selections.
go run ./examples/catalog -channel 42 -signature '验证码' -params '{"code":"123456","expire":"5"}' -receiver '13800138000' -send
```

## Error Handling

The SDK provides comprehensive error handling mechanisms.

### Check API Errors

```go
data, err := client.SendMessage(ctx, req)
if err != nil {
    if mlievpush.IsAPIError(err) {
        // API business error
        apiErr := err.(*mlievpush.APIError)
        fmt.Printf("Error Code: %d\n", apiErr.Code)
        fmt.Printf("Error Message: %s\n", apiErr.Message)
    } else {
        // Network error or other errors
        fmt.Printf("Request failed: %v\n", err)
    }
}
```

### Error Code Handling

```go
if mlievpush.IsAPIError(err) {
    apiErr := err.(*mlievpush.APIError)

    switch apiErr.Code {
    case mlievpush.ErrCodeInvalidSignature:
        // Signature verification failed
        fmt.Println("Please check if app_secret is correct")

    case mlievpush.ErrCodeChannelNotFound:
        // Channel not found
        fmt.Println("Please check if channel_id is correct")

    case mlievpush.ErrCodeRateLimitExceeded:
        // Rate limit exceeded
        fmt.Println("Please reduce request frequency")

    default:
        // Other errors
        fmt.Printf("%s\n", mlievpush.GetErrorMessage(apiErr.Code))
    }
}
```

### Common Error Codes

| Error Code | Constant | Description |
|------------|----------|-------------|
| 20003 | `ErrCodeInvalidSignature` | Signature verification failed |
| 20004 | `ErrCodeInvalidTimestamp` | Invalid timestamp |
| 30001 | `ErrCodeRateLimitExceeded` | Rate limit exceeded |
| 30003 | `ErrCodeChannelNotFound` | Channel not found |
| 30007 | `ErrCodeTaskNotFound` | Task not found |

For a complete list of error codes, please refer to [API Documentation](doc/API_INTEGRATION.md#错误码参考).

## Context Support

All API methods support Context for timeout control and request cancellation.

### Timeout Control

```go
// Create a context with 5-second timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

req := &mlievpush.SendMessageRequest{
    ChannelID:     1,
    SignatureName: "【Your Signature】",
    Receiver:      "13800138000",
}

data, err := client.SendMessage(ctx, req)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        fmt.Println("Request timeout")
    }
}
```

### Request Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())

// Can cancel request in another goroutine
go func() {
    time.Sleep(1 * time.Second)
    cancel()
}()

req := &mlievpush.SendMessageRequest{
    ChannelID:     1,
    SignatureName: "【Your Signature】",
    Receiver:      "13800138000",
}

data, err := client.SendMessage(ctx, req)
if err != nil {
    if ctx.Err() == context.Canceled {
        fmt.Println("Request canceled")
    }
}
```

## Constants

### Task Status

```go
mlievpush.TaskStatusPending     // "pending" - Pending
mlievpush.TaskStatusProcessing  // "processing" - Processing
mlievpush.TaskStatusSuccess     // "success" - Success
mlievpush.TaskStatusFailed      // "failed" - Failed
```

### Message Types

```go
mlievpush.MessageTypeSMS         // "sms" - SMS
mlievpush.MessageTypeEmail       // "email" - Email
mlievpush.MessageTypeWechatWork  // "wechat_work" - WeChat Work
mlievpush.MessageTypeDingtalk    // "dingtalk" - DingTalk
mlievpush.MessageTypeWebhook     // "webhook" - Webhook
mlievpush.MessageTypePush        // "push" - Push notification
```

### Callback Status

```go
mlievpush.CallbackStatusDelivered  // "delivered" - Delivered
mlievpush.CallbackStatusFailed     // "failed" - Failed
mlievpush.CallbackStatusRejected   // "rejected" - Rejected
```

## Complete Examples

Check [examples/main.go](examples/main.go) for more usage examples:

- Send single message
- Send batch messages
- Query task status
- Error handling
- Context timeout control

Run the examples:

```bash
cd examples
go run main.go
```

## Testing

Run unit tests:

```bash
go test -v
```

Run tests with coverage:

```bash
go test -v -cover
```

## Best Practices

1. **Reuse client instance**: `Client` is thread-safe and can be shared across multiple goroutines
2. **Use Context**: Set reasonable timeout for each request to avoid long blocking
3. **Error handling**: Differentiate between API errors and network errors for targeted handling
4. **Logging**: Save returned `task_id` for easier troubleshooting
5. **Batch limits**: Recommend not exceeding 500 items per batch send

## Dependencies

- Go 1.21+
- github.com/google/uuid v1.5.0

## License

MIT License

## Support

For questions or suggestions, please submit an Issue.