package mlievpush

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewEmailAttachmentEncodesContentAndDetectsMIME(t *testing.T) {
	attachment, err := NewEmailAttachment("报告.txt", []byte("hello"))
	if err != nil {
		t.Fatalf("NewEmailAttachment() error = %v", err)
	}
	if attachment.Filename != "报告.txt" {
		t.Fatalf("Filename = %q", attachment.Filename)
	}
	if !strings.HasPrefix(attachment.ContentType, "text/plain") {
		t.Fatalf("ContentType = %q, want text/plain", attachment.ContentType)
	}
	if attachment.ContentBase64 != base64.StdEncoding.EncodeToString([]byte("hello")) {
		t.Fatalf("ContentBase64 = %q", attachment.ContentBase64)
	}
}

func TestNewEmailAttachmentFromFileUsesBaseName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invoice.pdf")
	content := []byte("%PDF-1.4\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	attachment, err := NewEmailAttachmentFromFile(path)
	if err != nil {
		t.Fatalf("NewEmailAttachmentFromFile() error = %v", err)
	}
	if attachment.Filename != "invoice.pdf" || attachment.ContentBase64 != base64.StdEncoding.EncodeToString(content) {
		t.Fatalf("attachment = %+v", attachment)
	}
	if attachment.ContentType != "application/pdf" {
		t.Fatalf("ContentType = %q, want application/pdf", attachment.ContentType)
	}
}

func TestEmailAttachmentValidationRejectsUnsafeInput(t *testing.T) {
	tests := []struct {
		name       string
		attachment EmailAttachment
		want       string
	}{
		{name: "path", attachment: EmailAttachment{Filename: "../secret.txt", ContentBase64: "eA=="}, want: "path separators"},
		{name: "header injection", attachment: EmailAttachment{Filename: "safe.txt\r\nBcc: bad@example.com", ContentBase64: "eA=="}, want: "control characters"},
		{name: "empty", attachment: EmailAttachment{Filename: "safe.txt"}, want: "must not be empty"},
		{name: "invalid base64", attachment: EmailAttachment{Filename: "safe.txt", ContentBase64: "%%%"}, want: "standard Base64"},
		{name: "invalid MIME", attachment: EmailAttachment{Filename: "safe.txt", ContentType: "text/plain\r\nX-Test: bad", ContentBase64: "eA=="}, want: "CR or LF"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmailAttachments([]EmailAttachment{tt.attachment})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}
