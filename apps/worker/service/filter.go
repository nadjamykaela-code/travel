package service

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/nadjamykaela-code/travel/pkg/models"
)

type FilterRepo struct {
	client *firestore.Client
}

func NewFilterRepo(client *firestore.Client) *FilterRepo {
	return &FilterRepo{client: client}
}

func (r *FilterRepo) ListActive(ctx context.Context) ([]models.Filter, error) {
	iter := r.client.Collection("filters").
		Where("isActive", "==", true).
		Documents(ctx)

	var filters []models.Filter
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		var f models.Filter
		if err := doc.DataTo(&f); err != nil {
			return nil, fmt.Errorf("decode filter %s: %w", doc.Ref.ID, err)
		}
		f.ID = doc.Ref.ID
		filters = append(filters, f)
	}
	return filters, nil
}
