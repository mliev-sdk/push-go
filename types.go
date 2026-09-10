package mlievpush

import "encoding/json"

// EmailAttachment 邮件附件。ContentBase64 不包含 data URL 前缀。
type EmailAttachment struct {
	Filename      string `json:"filename"`
	ContentType   string `json:"content_type,omitempty"`
	ContentBase64 string `json:"content_base64"`
}

// SendMessageRequest 发送单条消息请求
type SendMessageRequest struct {
	ChannelID      int               `json:"channel_id"`                // 通道ID（必填）
	SignatureName  string            `json:"signature_name"`            // signature_names 中的别名；signature_required=true 时必填
	Receiver       string            `json:"receiver"`                  // 接收者（必填）
	TemplateParams map[string]string `json:"template_params,omitempty"` // 模板参数（可选）
	Attachments    []EmailAttachment `json:"attachments,omitempty"`     // 邮件附件（可选）
	ScheduledAt    string            `json:"scheduled_at,omitempty"`    // 定时发送时间（ISO 8601格式，可选）
}

// SendBatchRequest 批量发送消息请求
type SendBatchRequest struct {
	ChannelID      int               `json:"channel_id"`                // 通道ID（必填）
	SignatureName  string            `json:"signature_name"`            // signature_names 中的别名；signature_required=true 时必填
	Receivers      []string          `json:"receivers"`                 // 接收者列表（必填）
	TemplateParams map[string]string `json:"template_params,omitempty"` // 模板参数（可选）
	Attachments    []EmailAttachment `json:"attachments,omitempty"`     // 邮件附件，所有接收者共用（可选）
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

// ListChannelsRequest contains optional catalog filters. Zero values are omitted.
type ListChannelsRequest struct {
	Type     string
	Page     int
	PageSize int
}

// ListChannelsData is a page of enabled channels, ordered by descending ID.
type ListChannelsData struct {
	Items []ChannelData `json:"items"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
}

// ChannelData identifies a channel, its bound system template and its readiness.
type ChannelData struct {
	ID                int              `json:"id"`
	Name              string           `json:"name"`
	Type              string           `json:"type"`
	MessageTemplateID int              `json:"message_template_id"`
	TemplateName      string           `json:"template_name"`
	Readiness         ChannelReadiness `json:"readiness"`
}

// ChannelReadiness describes configuration at query time; sending revalidates it.
type ChannelReadiness struct {
	State        string   `json:"state"`
	BlockerCodes []string `json:"blocker_codes"`
}

// ChannelTemplate is the system template used to build TemplateParams.
type ChannelTemplate struct {
	ID           int      `json:"id"`
	TemplateName string   `json:"template_name"`
	ContentType  string   `json:"content_type"`
	Content      string   `json:"content"`
	Variables    []string `json:"variables"` // nil means invalid configuration; [] means no variables
	Description  string   `json:"description"`
}

// ChannelDetailData contains all configuration needed by a message form.
type ChannelDetailData struct {
	ChannelData
	Template          *ChannelTemplate `json:"template"` // nil if the template is missing or deleted
	SignatureRequired bool             `json:"signature_required"`
	SignatureNames    []string         `json:"signature_names"` // pass the selected alias to SignatureName
}

// Channel readiness states. Both ready and degraded channels can send.
const (
	ChannelReadinessReady    = "ready"
	ChannelReadinessDegraded = "degraded"
	ChannelReadinessBlocked  = "blocked"
)

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
