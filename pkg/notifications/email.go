package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type EmailSender interface {
	Send(ctx context.Context, to, subject, htmlContent string) error
}

type SendGridClient struct {
	apiKey string
	url    string
	client *http.Client
}

func NewSendGrid(apiKey string) *SendGridClient {
	return &SendGridClient{
		apiKey: apiKey,
		url:    "https://api.sendgrid.com/v3/mail/send",
		client: &http.Client{},
	}
}

func (s *SendGridClient) Send(ctx context.Context, to, subject, htmlContent string) error {
	payload := map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{
				"to": []map[string]string{{"email": to}},
			},
		},
		"from":    map[string]string{"email": "noreply@travelbot.com"},
		"subject": subject,
		"content": []map[string]string{
			{"type": "text/html", "value": htmlContent},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sendgrid error: %d", resp.StatusCode)
	}
	return nil
}
