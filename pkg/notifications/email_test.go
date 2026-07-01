package notifications

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendGridClient_Send(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful send",
			statusCode: http.StatusAccepted,
			wantErr:    false,
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") == "" {
					t.Error("missing Authorization header")
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Error("missing Content-Type header")
				}
				w.WriteHeader(tt.statusCode)
			}))
			defer srv.Close()

			client := &SendGridClient{
				apiKey: "test-key",
				url:    srv.URL,
				client: srv.Client(),
			}

			err := client.Send(context.Background(), "test@example.com", "Subject", "<p>Body</p>")
			if (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewSendGrid(t *testing.T) {
	c := NewSendGrid("key-123")
	if c.apiKey != "key-123" {
		t.Errorf("apiKey = %q; want %q", c.apiKey, "key-123")
	}
	if c.client == nil {
		t.Error("client should not be nil")
	}
	if c.url != "https://api.sendgrid.com/v3/mail/send" {
		t.Errorf("url = %q; want %q", c.url, "https://api.sendgrid.com/v3/mail/send")
	}
}

func TestSendGridClient_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	client := &SendGridClient{
		apiKey: "test-key",
		url:    srv.URL,
		client: srv.Client(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.Send(ctx, "test@example.com", "Sub", "<p>Body</p>")
	if err == nil {
		t.Error("expected error with cancelled context, got nil")
	}
}
