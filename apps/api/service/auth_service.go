package service

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type AuthService interface {
	VerifyToken(ctx context.Context, tokenString string) (string, error)
}

type AuthServiceFirebase struct {
	client *auth.Client
}

func NewAuthServiceFirebase(ctx context.Context, credsPath string) (*AuthServiceFirebase, error) {
	opt := option.WithCredentialsFile(credsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase app init: %w", err)
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase auth init: %w", err)
	}
	return &AuthServiceFirebase{client: client}, nil
}

func (a *AuthServiceFirebase) VerifyToken(ctx context.Context, tokenString string) (string, error) {
	token, err := a.client.VerifyIDToken(ctx, tokenString)
	if err != nil {
		return "", fmt.Errorf("verify firebase token: %w", err)
	}
	return token.UID, nil
}

type NoopAuthService struct{}

func NewNoopAuthService() *NoopAuthService {
	return &NoopAuthService{}
}

func (n *NoopAuthService) VerifyToken(_ context.Context, _ string) (string, error) {
	return "", fmt.Errorf("auth service not configured")
}
