package sender

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/smtp"
	"os"
	"path/filepath"

	"substack-to-kindle/pkg/converter"
)

// EmailConfig contains email configuration
type EmailConfig struct {
	From     string
	To       string
	Password string
	SMTPHost string
	SMTPPort string
}

// SendToKindle sends a converted file to a Kindle device via email
func SendToKindle(result *converter.ConversionResult, config EmailConfig) error {
	// Create buffer for message
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// Set email headers
	buffer.WriteString(fmt.Sprintf("From: %s\r\n", config.From))
	buffer.WriteString(fmt.Sprintf("To: %s\r\n", config.To))
	buffer.WriteString(fmt.Sprintf("Subject: %s\r\n", result.Title))
	buffer.WriteString("MIME-Version: 1.0\r\n")
	buffer.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", writer.Boundary()))
	buffer.WriteString("\r\n")

	// Add text part
	textPart, err := writer.CreatePart(map[string][]string{
		"Content-Type": {"text/plain; charset=UTF-8"},
	})
	if err != nil {
		return fmt.Errorf("failed to create text part: %w", err)
	}

	textContent := fmt.Sprintf("Sending '%s' by %s to your Kindle.\r\n",
		result.Title, result.Author)
	_, err = textPart.Write([]byte(textContent))
	if err != nil {
		return fmt.Errorf("failed to write text content: %w", err)
	}

	// Add attachment part
	file, err := os.Open(result.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	attachmentPart, err := writer.CreatePart(map[string][]string{
		"Content-Type":              {"application/octet-stream"},
		"Content-Transfer-Encoding": {"base64"},
		"Content-Disposition":       {fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(result.FilePath))},
	})
	if err != nil {
		return fmt.Errorf("failed to create attachment part: %w", err)
	}

	// Read file and encode as base64
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	encoder := base64.NewEncoder(base64.StdEncoding, attachmentPart)
	_, err = encoder.Write(fileContent)
	if err != nil {
		return fmt.Errorf("failed to encode file: %w", err)
	}
	encoder.Close()

	// Close multipart writer
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	// Authentication
	auth := smtp.PlainAuth("", config.From, config.Password, config.SMTPHost)

	// Send email
	err = smtp.SendMail(
		config.SMTPHost+":"+config.SMTPPort,
		auth,
		config.From,
		[]string{config.To},
		buffer.Bytes(),
	)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// LoadEmailConfigFromEnv loads email configuration from environment variables
func LoadEmailConfigFromEnv() EmailConfig {
	return EmailConfig{
		From:     os.Getenv("EMAIL_FROM"),
		To:       os.Getenv("EMAIL_TO"),
		Password: os.Getenv("EMAIL_PASSWORD"),
		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: os.Getenv("SMTP_PORT"),
	}
}
