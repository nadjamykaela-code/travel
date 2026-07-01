package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/nadjamykaela-code/travel/pkg/models"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func TestNotifier_Send(t *testing.T) {
	tests := []struct {
		name      string
		filter    models.Filter
		emailErr  error
		pushErr   error
		noPush    bool
		wantErr   bool
	}{
		{
			name: "sends email and push",
			filter: models.Filter{
				NotifyEmail:     "test@example.com",
				NotifyPushToken: "push-token",
			},
		},
		{
			name: "sends email only",
			filter: models.Filter{
				NotifyEmail: "test@example.com",
			},
		},
		{
			name: "sends push only",
			filter: models.Filter{
				NotifyPushToken: "push-token",
			},
		},
		{
			name:   "no notifications configured",
			filter: models.Filter{},
		},
		{
			name: "email error propagates",
			filter: models.Filter{
				NotifyEmail: "test@example.com",
			},
			emailErr: assertAnError{"sendgrid error"},
			wantErr:  true,
		},
		{
			name: "no push notifier does not block",
			filter: models.Filter{
				NotifyEmail:     "test@example.com",
				NotifyPushToken: "push-token",
			},
			noPush: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email := &mockEmailSender{err: tt.emailErr}
			var n *Notifier
			if tt.noPush {
				n = NewNotifier(email, nil, discardLogger())
			} else {
				n = NewNotifier(email, &mockPushSender{err: tt.pushErr}, discardLogger())
			}

			itineraries := []models.Itinerary{{
				ID: "it-1",
				TotalPrice: models.Price{Amount: 1000, Currency: "BRL"},
			}}

			err := n.Send(context.Background(), tt.filter, itineraries)
			if (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestNotifier_Send_NoItineraries(t *testing.T) {
	email := &mockEmailSender{}
	n := NewNotifier(email, nil, discardLogger())
	err := n.Send(context.Background(), models.Filter{
		NotifyEmail: "test@example.com",
	}, nil)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if email.sentTo != "" {
		t.Error("email should not have been sent")
	}
}
