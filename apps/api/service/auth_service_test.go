package service

import (
	"context"
	"testing"
)

func TestNewAuthServiceFirebase_InvalidCredentials(t *testing.T) {
	_, err := NewAuthServiceFirebase(context.Background(), "/nonexistent/creds.json")
	if err == nil {
		t.Error("expected error with invalid creds path, got nil")
	}
}

func TestNewNoopAuthService(t *testing.T) {
	svc := NewNoopAuthService()
	_, err := svc.VerifyToken(context.Background(), "anything")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestAuthServiceInterface(t *testing.T) {
	var _ AuthService = (*AuthServiceFirebase)(nil)
	var _ AuthService = (*NoopAuthService)(nil)
}
