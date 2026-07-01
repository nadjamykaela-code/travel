package service

import (
	"context"
)

type mockEmailSender struct {
	sentTo string
	err    error
}

func (m *mockEmailSender) Send(_ context.Context, to, _, _ string) error {
	m.sentTo = to
	return m.err
}

type mockPushSender struct {
	sent bool
	err  error
}

func (m *mockPushSender) Send(_ context.Context, _, _, _ string) error {
	m.sent = true
	return m.err
}

type assertAnError struct {
	msg string
}

func (e assertAnError) Error() string {
	if e.msg != "" {
		return e.msg
	}
	return "error"
}


