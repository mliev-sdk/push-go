package mlievpush

import (
	"encoding/base64"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

// NewEmailAttachment 从内存内容创建邮件附件并自动探测 MIME 类型。
// SDK 不限制附件数量和大小，最终限制由服务端配置决定。
func NewEmailAttachment(filename string, content []byte) (EmailAttachment, error) {
	filename, err := validateAttachmentFilename(filename)
	if err != nil {
		return EmailAttachment{}, err
	}
	if len(content) == 0 {
		return EmailAttachment{}, fmt.Errorf("attachment content must not be empty")
	}

	return EmailAttachment{
		Filename:      filename,
		ContentType:   detectAttachmentContentType(filename, content),
		ContentBase64: base64.StdEncoding.EncodeToString(content),
	}, nil
}

// NewEmailAttachmentFromFile 读取本地文件并创建邮件附件。
// 文件名使用路径的 basename，文件读取错误会直接返回。
func NewEmailAttachmentFromFile(path string) (EmailAttachment, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return EmailAttachment{}, fmt.Errorf("read email attachment %q: %w", path, err)
	}
	return NewEmailAttachment(filepath.Base(path), content)
}

func validateEmailAttachments(attachments []EmailAttachment) error {
	for index, attachment := range attachments {
		if _, err := validateAttachmentFilename(attachment.Filename); err != nil {
			return fmt.Errorf("attachment %d filename: %w", index+1, err)
		}
		if attachment.ContentBase64 == "" {
			return fmt.Errorf("attachment %d content_base64 must not be empty", index+1)
		}
		content, err := base64.StdEncoding.DecodeString(attachment.ContentBase64)
		if err != nil {
			return fmt.Errorf("attachment %d content_base64 must be standard Base64", index+1)
		}
		if len(content) == 0 {
			return fmt.Errorf("attachment %d content must not be empty", index+1)
		}
		if err := validateAttachmentContentType(attachment.ContentType); err != nil {
			return fmt.Errorf("attachment %d content_type: %w", index+1, err)
		}
	}
	return nil
}

func validateAttachmentFilename(value string) (string, error) {
	filename := strings.TrimSpace(value)
	if filename == "" {
		return "", fmt.Errorf("must not be empty")
	}
	if !utf8.ValidString(filename) {
		return "", fmt.Errorf("must be valid UTF-8")
	}
	if len([]byte(filename)) > 255 {
		return "", fmt.Errorf("must not exceed 255 bytes")
	}
	if filename == "." || filename == ".." || strings.ContainsAny(filename, "/\\") {
		return "", fmt.Errorf("must be a plain filename without path separators")
	}
	for _, character := range filename {
		if unicode.IsControl(character) {
			return "", fmt.Errorf("must not contain control characters")
		}
	}
	return filename, nil
}

func validateAttachmentContentType(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("must not contain CR or LF characters")
	}
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil || mediaType == "" {
		return fmt.Errorf("must be a valid MIME media type")
	}
	return nil
}

func detectAttachmentContentType(filename string, content []byte) string {
	if contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(filename))); contentType != "" {
		return contentType
	}
	return http.DetectContentType(content)
}
