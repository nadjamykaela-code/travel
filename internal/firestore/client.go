package firestore

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

type Client struct {
	inner *firestore.Client

	mu sync.RWMutex
}

func New(ctx context.Context, projectID string, credsPath string) (*Client, error) {
	cfg := (&firebase.Config{ProjectID: projectID})
	app, err := firebase.NewApp(ctx, cfg)
	if err != nil {
		if credsPath == "" {
			return nil, fmt.Errorf("firebase default app: %w", err)
		}
		slog.Warn("firebase default app failed, trying local credentials", "error", err)
		opt := option.WithCredentialsFile(credsPath)
		app, err = firebase.NewApp(ctx, cfg, opt)
		if err != nil {
			return nil, fmt.Errorf("firebase app with credentials: %w", err)
		}
	}

	inner, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("firestore client: %w", err)
	}

	return &Client{inner: inner}, nil
}

func (c *Client) Firestore() *firestore.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.inner
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.inner == nil {
		return nil
	}
	if err := c.inner.Close(); err != nil {
		return fmt.Errorf("close firestore: %w", err)
	}
	c.inner = nil
	return nil
}
