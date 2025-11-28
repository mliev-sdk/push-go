package main

import (
	"context"
	"fmt"
	"log"
	"time"

	mlievpush "github.com/mliev-sdk/push-go"
)

func main() {
	// 创建客户端
	client := mlievpush.NewClient(
		"https://your-domain.com",             // 基础URL
		"your_app_id",                         // 应用ID
		"your_app_secret",                     // 应用密钥
		mlievpush.WithTimeout(15*time.Second), // 可选：设置超时时间
	)

	// 创建Context
	ctx := context.Background()

	// 示例1：发送单条消息
	fmt.Println("=== 示例1：发送单条短信 ===")
	sendSingleMessage(ctx, client)

	// 示例2：批量发送消息
	fmt.Println("\n=== 示例2：批量发送消息 ===")
	sendBatchMessages(ctx, client)

	// 示例3：查询任务状态
	fmt.Println("\n=== 示例3：查询任务状态 ===")
	queryTaskStatus(ctx, client)

	// 示例4：错误处理
	fmt.Println("\n=== 示例4：错误处理 ===")
	handleErrors(ctx, client)

	// 示例5：带超时的请求
	fmt.Println("\n=== 示例5：带超时的请求 ===")
	requestWithTimeout(client)
}

// sendSingleMessage 发送单条消息示例
func sendSingleMessage(ctx context.Context, client *mlievpush.Client) {
	req := &mlievpush.SendMessageRequest{
		ChannelID:     1,
		SignatureName: "【您的签名】",
		Receiver:      "13800138000",
		TemplateParams: map[string]string{
			"code":        "123456",
			"expire_time": "5",
		},
	}

	data, err := client.SendMessage(ctx, req)
	if err != nil {
		log.Printf("发送消息失败: %v\n", err)
		return
	}

	fmt.Printf("发送成功！\n")
	fmt.Printf("  任务ID: %s\n", data.TaskID)
	fmt.Printf("  状态: %s\n", data.Status)
	fmt.Printf("  创建时间: %s\n", data.CreatedAt)
}

// sendBatchMessages 批量发送消息示例
func sendBatchMessages(ctx context.Context, client *mlievpush.Client) {
	req := &mlievpush.SendBatchRequest{
		ChannelID:     1,
		SignatureName: "【您的签名】",
		Receivers: []string{
			"13800138000",
			"13800138001",
			"13800138002",
		},
		TemplateParams: map[string]string{
			"content":  "系统将于今晚22:00进行维护",
			"duration": "2小时",
		},
	}

	data, err := client.SendBatch(ctx, req)
	if err != nil {
		log.Printf("批量发送失败: %v\n", err)
		return
	}

	fmt.Printf("批量发送成功！\n")
	fmt.Printf("  批次ID: %s\n", data.BatchID)
	fmt.Printf("  总数量: %d\n", data.TotalCount)
	fmt.Printf("  成功数量: %d\n", data.SuccessCount)
	fmt.Printf("  失败数量: %d\n", data.FailedCount)
	fmt.Printf("  创建时间: %s\n", data.CreatedAt)
}

// queryTaskStatus 查询任务状态示例
func queryTaskStatus(ctx context.Context, client *mlievpush.Client) {
	taskID := "550e8400-e29b-41d4-a716-446655440000"

	data, err := client.QueryTask(ctx, taskID)
	if err != nil {
		log.Printf("查询任务失败: %v\n", err)
		return
	}

	fmt.Printf("查询成功！\n")
	fmt.Printf("  任务ID: %s\n", data.TaskID)
	fmt.Printf("  应用ID: %s\n", data.AppID)
	fmt.Printf("  通道ID: %d\n", data.ChannelID)
	fmt.Printf("  消息类型: %s\n", data.MessageType)
	fmt.Printf("  接收者: %s\n", data.Receiver)
	fmt.Printf("  内容: %s\n", data.Content)
	fmt.Printf("  状态: %s\n", data.Status)
	fmt.Printf("  回调状态: %s\n", data.CallbackStatus)
	fmt.Printf("  重试次数: %d/%d\n", data.RetryCount, data.MaxRetry)
	fmt.Printf("  创建时间: %s\n", data.CreatedAt)
	fmt.Printf("  更新时间: %s\n", data.UpdatedAt)
}

// handleErrors 错误处理示例
func handleErrors(ctx context.Context, client *mlievpush.Client) {
	// 故意使用无效的通道ID来触发错误
	req := &mlievpush.SendMessageRequest{
		ChannelID:     99999,
		SignatureName: "【您的签名】",
		Receiver:      "13800138000",
	}

	_, err := client.SendMessage(ctx, req)
	if err != nil {
		// 检查是否为API错误
		if mlievpush.IsAPIError(err) {
			apiErr := err.(*mlievpush.APIError)
			fmt.Printf("API错误: [%d] %s\n", apiErr.Code, apiErr.Message)

			// 根据错误码进行不同的处理
			switch apiErr.Code {
			case mlievpush.ErrCodeChannelNotFound:
				fmt.Println("  处理建议: 检查通道ID是否正确")
			case mlievpush.ErrCodeInvalidSignature:
				fmt.Println("  处理建议: 检查签名算法和密钥")
			case mlievpush.ErrCodeRateLimitExceeded:
				fmt.Println("  处理建议: 降低请求频率")
			default:
				fmt.Printf("  错误码说明: %s\n", mlievpush.GetErrorMessage(apiErr.Code))
			}
		} else {
			// 网络错误或其他错误
			fmt.Printf("请求失败: %v\n", err)
		}
	}
}

// requestWithTimeout 带超时的请求示例
func requestWithTimeout(client *mlievpush.Client) {
	// 创建一个5秒超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &mlievpush.SendMessageRequest{
		ChannelID:     1,
		SignatureName: "【您的签名】",
		Receiver:      "13800138000",
		TemplateParams: map[string]string{
			"code": "123456",
		},
	}

	data, err := client.SendMessage(ctx, req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("请求超时")
		} else {
			fmt.Printf("请求失败: %v\n", err)
		}
		return
	}

	fmt.Printf("发送成功，任务ID: %s\n", data.TaskID)
}
