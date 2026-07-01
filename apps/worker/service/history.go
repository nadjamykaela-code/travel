package service

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/nadjamykaela-code/travel/pkg/models"
)

type HistoryStore struct {
	client *firestore.Client
}

func NewHistoryStore(client *firestore.Client) *HistoryStore {
	return &HistoryStore{client: client}
}

func (s *HistoryStore) Record(ctx context.Context, filter models.Filter, itineraries []models.Itinerary) error {
	if len(itineraries) == 0 {
		return nil
	}
	best := itineraries[0]
	doc := map[string]interface{}{
		"filterId":  filter.ID,
		"userId":    filter.UserID,
		"found":     len(itineraries),
		"bestPrice": best.TotalPrice.Amount,
		"currency":  best.TotalPrice.Currency,
		"itinerary": best,
		"createdAt": time.Now(),
	}
	_, _, err := s.client.Collection("history").Add(ctx, doc)
	if err != nil {
		return fmt.Errorf("add history: %w", err)
	}
	return nil
}
