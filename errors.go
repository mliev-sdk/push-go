package mlievpush

import "fmt"

// APIError API错误
type APIError struct {
	Code    int    // 错误码
	Message string // 错误消息
}

// Error 实现 error 接口
func (e *APIError) Error() string {
	return fmt.Sprintf("API error [%d]: %s", e.Code, e.Message)
}

// NewAPIError 创建API错误
func NewAPIError(code int, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// IsAPIError 判断是否为API错误
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// 错误码常量定义

// 请求错误 (1xxxx)
const (
	ErrCodeInvalidParams   = 10001 // 请求参数错误
	ErrCodeInvalidJSON     = 10002 // JSON格式错误
	ErrCodeMissingParams   = 10003 // 缺少必填参数
	ErrCodeInvalidValue    = 10004 // 参数值非法
	ErrCodeInvalidReceiver = 10005 // 接收者格式错误
	ErrCodeInvalidTemplate = 10006 // 模板参数错误
)

// 鉴权错误 (2xxxx)
const (
	ErrCodeUnauthorized     = 20001 // 未授权
	ErrCodeInvalidAppID     = 20002 // 无效的AppID
	ErrCodeInvalidSignature = 20003 // 签名验证失败
	ErrCodeInvalidTimestamp = 20004 // 时间戳无效
	ErrCodeIPNotAllowed     = 20005 // IP不在白名单
	ErrCodeAppDisabled      = 20006 // 应用已禁用
)

// 业务错误 (3xxxx)
const (
	ErrCodeRateLimitExceeded  = 30001 // 超出速率限制
	ErrCodeQuotaExceeded      = 30002 // 超出配额限制
	ErrCodeChannelNotFound    = 30003 // 通道不存在
	ErrCodeChannelDisabled    = 30004 // 通道已禁用
	ErrCodeTemplateNotFound   = 30005 // 模板不存在
	ErrCodeNoAvailableChannel = 30006 // 无可用通道
	ErrCodeTaskNotFound       = 30007 // 任务不存在
	ErrCodeBatchNotFound      = 30008 // 批量任务不存在
)

// 系统错误 (4xxxx)
const (
	ErrCodeInternalError  = 40001 // 内部错误
	ErrCodeDatabaseError  = 40002 // 数据库错误
	ErrCodeRedisError     = 40003 // Redis错误
	ErrCodeQueueError     = 40004 // 队列错误
	ErrCodeProviderError  = 40005 // 服务商错误
	ErrCodeNetworkTimeout = 40006 // 网络超时
	ErrCodeCircuitOpen    = 40007 // 熔断器打开
)

// ErrorCodeMessages 错误码对应的消息
var ErrorCodeMessages = map[int]string{
	// 请求错误
	ErrCodeInvalidParams:   "请求参数错误",
	ErrCodeInvalidJSON:     "JSON格式错误",
	ErrCodeMissingParams:   "缺少必填参数",
	ErrCodeInvalidValue:    "参数值非法",
	ErrCodeInvalidReceiver: "接收者格式错误",
	ErrCodeInvalidTemplate: "模板参数错误",

	// 鉴权错误
	ErrCodeUnauthorized:     "未授权",
	ErrCodeInvalidAppID:     "无效的AppID",
	ErrCodeInvalidSignature: "签名验证失败",
	ErrCodeInvalidTimestamp: "时间戳无效",
	ErrCodeIPNotAllowed:     "IP不在白名单",
	ErrCodeAppDisabled:      "应用已禁用",

	// 业务错误
	ErrCodeRateLimitExceeded:  "超出速率限制",
	ErrCodeQuotaExceeded:      "超出配额限制",
	ErrCodeChannelNotFound:    "通道不存在",
	ErrCodeChannelDisabled:    "通道已禁用",
	ErrCodeTemplateNotFound:   "模板不存在",
	ErrCodeNoAvailableChannel: "无可用通道",
	ErrCodeTaskNotFound:       "任务不存在",
	ErrCodeBatchNotFound:      "批量任务不存在",

	// 系统错误
	ErrCodeInternalError:  "内部错误",
	ErrCodeDatabaseError:  "数据库错误",
	ErrCodeRedisError:     "Redis错误",
	ErrCodeQueueError:     "队列错误",
	ErrCodeProviderError:  "服务商错误",
	ErrCodeNetworkTimeout: "网络超时",
	ErrCodeCircuitOpen:    "熔断器打开",
}

// GetErrorMessage 根据错误码获取错误消息
func GetErrorMessage(code int) string {
	if msg, ok := ErrorCodeMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
