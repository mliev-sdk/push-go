package mlievpush

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// sortParams 按 key 排序参数并返回 JSON 字符串
// 如果 params 为空或 nil，返回空字符串
func sortParams(params map[string]interface{}) string {
	if len(params) == 0 {
		return ""
	}

	// 提取所有 key 并排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 按排序后的 key 构建新的 map
	sortedMap := make(map[string]interface{})
	for _, k := range keys {
		sortedMap[k] = params[k]
	}

	// 序列化为 JSON（不转义 Unicode，无空格）
	result, _ := json.Marshal(sortedMap)
	return string(result)
}

// generateSignature 生成请求签名
// 签名算法: HMAC-SHA256(method + path + sorted_params + timestamp + nonce, app_secret)
func generateSignature(method, path string, params map[string]interface{}, timestamp, nonce, appSecret string) string {
	sortedParams := sortParams(params)

	// 构造签名内容: method + path + sorted_params + timestamp + nonce
	signContent := method + path + sortedParams + timestamp + nonce

	// HMAC-SHA256 计算
	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write([]byte(signContent))

	// 十六进制编码（小写）
	return hex.EncodeToString(mac.Sum(nil))
}
