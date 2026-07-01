package notifications

import (
	"context"
	"testing"
)

func TestPushSenderInterface(t *testing.T) {
	var _ PushSender = (*FCMNotifier)(nil)
}

func TestNewFCMNotifier_InvalidCreds(t *testing.T) {
	_, err := NewFCMNotifier("/nonexistent/creds.json")
	if err == nil {
		t.Error("expected error with invalid credentials, got nil")
	}
}

func TestFCMNotifier_Send_NilClient(t *testing.T) {
	n := &FCMNotifier{client: nil}
	err := n.Send(context.Background(), "token", "title", "body")
	if err == nil {
		t.Error("expected error with nil client, got nil")
	}
}
