package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nadjamykaela-code/travel/pkg/models"
	"github.com/nadjamykaela-code/travel/pkg/notifications"
)

type Notifier struct {
	email  notifications.EmailSender
	push   notifications.PushSender
	logger *slog.Logger
}

func NewNotifier(email notifications.EmailSender, push notifications.PushSender, logger *slog.Logger) *Notifier {
	return &Notifier{email: email, push: push, logger: logger}
}

func (n *Notifier) Send(ctx context.Context, filter models.Filter, itineraries []models.Itinerary) error {
	if len(itineraries) == 0 {
		return nil
	}
	best := itineraries[0]
	message := fmt.Sprintf(
		"Encontramos %d ofertas para %s → %s! (melhor preço: %.2f %s)",
		len(itineraries), filter.Origin, filter.Destination,
		best.TotalPrice.Amount, best.TotalPrice.Currency,
	)

	if filter.NotifyEmail != "" {
		html := fmt.Sprintf(htmlTemplate,
			filter.Origin, filter.Destination,
			best.TotalPrice.Amount, best.TotalPrice.Currency,
			best.Stops, best.TotalDuration, best.BookingURL)
		if err := n.email.Send(ctx, filter.NotifyEmail, "Nova oferta de viagem!", html); err != nil {
			return fmt.Errorf("send email: %w", err)
		}
		n.logger.Info("email sent", "to", filter.NotifyEmail)
	}

	if filter.NotifyPushToken != "" && n.push != nil {
		if err := n.push.Send(ctx, filter.NotifyPushToken, "Nova oferta!", message); err != nil {
			return fmt.Errorf("send push: %w", err)
		}
		n.logger.Info("push sent")
	}

	return nil
}

const htmlTemplate = `<h2>Ofertas encontradas</h2>
<p><strong>%s → %s</strong></p>
<ul>
	<li>Preço: %.2f %s</li>
	<li>Escalas: %d</li>
	<li>Duração: %d min</li>
</ul>
<a href="%s">Ver detalhes</a>`
