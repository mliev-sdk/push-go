package mlievpush

import "encoding/json"

// SendMessageRequest 发送单条消息请求
type SendMessageRequest struct {
	ChannelID      int               `json:"channel_id"`                // 通道ID（必填）
	SignatureName  string            `json:"signature_name"`            // 签名名称（必填）
	Receiver       string            `json:"receiver"`                  // 接收者（必填）
	TemplateParams map[string]string `json:"template_params,omitempty"` // 模板参数（可选）
	ScheduledAt    string            `json:"scheduled_at,omitempty"`    // 定时发送时间（ISO 8601格式，可选）
}

// SendBatchRequest 批量发送消息请求
type SendBatchRequest struct {
	ChannelID      int               `json:"channel_id"`                // 通道ID（必填）
	SignatureName  string            `json:"signature_name"`            // 签名名称（必填）
	Receivers      []string          `json:"receivers"`                 // 接收者列表（必填）
	TemplateParams map[string]string `json:"template_params,omitempty"` // 模板参数（可选）
	ScheduledAt    string            `json:"scheduled_at,omitempty"`    // 定时发送时间（ISO 8601格式，可选）
}

// Response 通用API响应结构
type Response struct {
	Code    int             `json:"code"`    // 状态码，0表示成功
	Message string          `json:"message"` // 状态描述
	Data    json.RawMessage `json:"data"`    // 响应数据（原始JSON）
}

// SendMessageData 发送单条消息响应数据
type SendMessageData struct {
	TaskID    string `json:"task_id"`    // 任务ID（UUID格式）
	Status    string `json:"status"`     // 任务状态
	CreatedAt string `json:"created_at"` // 创建时间
}

// SendBatchData 批量发送消息响应数据
type SendBatchData struct {
	BatchID      string `json:"batch_id"`      // 批次ID
	TotalCount   int    `json:"total_count"`   // 总数量
	SuccessCount int    `json:"success_count"` // 成功入队数量
	FailedCount  int    `json:"failed_count"`  // 失败数量
	CreatedAt    string `json:"created_at"`    // 创建时间
}

// QueryTaskData 查询任务状态响应数据
type QueryTaskData struct {
	ID             int    `json:"id"`              // 任务内部ID
	TaskID         string `json:"task_id"`         // 任务ID
	AppID          string `json:"app_id"`          // 应用ID
	ChannelID      int    `json:"channel_id"`      // 通道ID
	MessageType    string `json:"message_type"`    // 消息类型
	Receiver       string `json:"receiver"`        // 接收者
	Content        string `json:"content"`         // 消息内容
	Status         string `json:"status"`          // 任务状态
	CallbackStatus string `json:"callback_status"` // 回调状态
	RetryCount     int    `json:"retry_count"`     // 已重试次数
	MaxRetry       int    `json:"max_retry"`       // 最大重试次数
	CreatedAt      string `json:"created_at"`      // 创建时间
	UpdatedAt      string `json:"updated_at"`      // 更新时间
}

// TaskStatus 任务状态枚举
const (
	TaskStatusPending    = "pending"    // 待处理
	TaskStatusProcessing = "processing" // 处理中
	TaskStatusSuccess    = "success"    // 成功
	TaskStatusFailed     = "failed"     // 失败
)

// CallbackStatus 回调状态枚举
const (
	CallbackStatusDelivered = "delivered" // 已送达
	CallbackStatusFailed    = "failed"    // 发送失败
	CallbackStatusRejected  = "rejected"  // 被拒绝
)

// MessageType 消息类型枚举
const (
	MessageTypeSMS        = "sms"         // 短信
	MessageTypeEmail      = "email"       // 邮件
	MessageTypeWechatWork = "wechat_work" // 企业微信
	MessageTypeDingtalk   = "dingtalk"    // 钉钉
	MessageTypeWebhook    = "webhook"     // Webhook
	MessageTypePush       = "push"        // 推送通知
)
