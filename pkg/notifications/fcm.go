package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type PushSender interface {
	Send(ctx context.Context, token, title, body string) error
}

type FCMNotifier struct {
	client *messaging.Client
	logger *slog.Logger
}

func NewFCMNotifier(credsPath string) (*FCMNotifier, error) {
	opt := option.WithCredentialsFile(credsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase app: %w", err)
	}
	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("messaging client: %w", err)
	}
	return &FCMNotifier{
		client: client,
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}, nil
}

func (f *FCMNotifier) Send(ctx context.Context, token, title, body string) error {
	if f.client == nil {
		return fmt.Errorf("fcm client not initialized")
	}
	msg := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: map[string]string{
			"click_action": "FLUTTER_NOTIFICATION_CLICK",
		},
	}
	resp, err := f.client.Send(ctx, msg)
	if err != nil {
		return fmt.Errorf("fcm send: %w", err)
	}
	f.logger.Info("push sent", "response", resp)
	return nil
}
