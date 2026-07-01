package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/nadjamykaela-code/travel/pkg/models"
)

type FilterStore interface {
	GetActiveFilters(ctx context.Context, userID string) ([]models.Filter, error)
	CreateFilter(ctx context.Context, filter *models.Filter) (string, error)
	GetFilterByID(ctx context.Context, filterID string) (*models.Filter, error)
	UpdateFilter(ctx context.Context, filterID string, updates map[string]interface{}) error
	SoftDeleteFilter(ctx context.Context, filterID string) error
}

type FilterStoreFirestore struct {
	client *firestore.Client
}

func NewFilterStoreFirestore(client *firestore.Client) *FilterStoreFirestore {
	return &FilterStoreFirestore{client: client}
}

func (s *FilterStoreFirestore) GetActiveFilters(ctx context.Context, userID string) ([]models.Filter, error) {
	iter := s.client.Collection("filters").
		Where("userId", "==", userID).
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
			return nil, fmt.Errorf("decode filter: %w", err)
		}
		f.ID = doc.Ref.ID
		filters = append(filters, f)
	}
	return filters, nil
}

func (s *FilterStoreFirestore) CreateFilter(ctx context.Context, filter *models.Filter) (string, error) {
	filter.CreatedAt = time.Now()
	filter.UpdatedAt = time.Now()
	filter.IsActive = true

	ref, _, err := s.client.Collection("filters").Add(ctx, filter)
	if err != nil {
		return "", fmt.Errorf("create filter: %w", err)
	}
	return ref.ID, nil
}

func (s *FilterStoreFirestore) GetFilterByID(ctx context.Context, filterID string) (*models.Filter, error) {
	doc, err := s.client.Collection("filters").Doc(filterID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get filter %s: %w", filterID, err)
	}
	var f models.Filter
	if err := doc.DataTo(&f); err != nil {
		return nil, fmt.Errorf("decode filter: %w", err)
	}
	f.ID = doc.Ref.ID
	return &f, nil
}

func mapToFirestoreUpdates(m map[string]interface{}) []firestore.Update {
	updates := make([]firestore.Update, 0, len(m))
	for k, v := range m {
		updates = append(updates, firestore.Update{Path: k, Value: v})
	}
	sort.Slice(updates, func(i, j int) bool {
		return updates[i].Path < updates[j].Path
	})
	return updates
}

func (s *FilterStoreFirestore) UpdateFilter(ctx context.Context, filterID string, updates map[string]interface{}) error {
	updates["updatedAt"] = time.Now()
	_, err := s.client.Collection("filters").Doc(filterID).Update(ctx, mapToFirestoreUpdates(updates))
	if err != nil {
		return fmt.Errorf("update filter %s: %w", filterID, err)
	}
	return nil
}

func (s *FilterStoreFirestore) SoftDeleteFilter(ctx context.Context, filterID string) error {
	_, err := s.client.Collection("filters").Doc(filterID).Update(ctx, mapToFirestoreUpdates(map[string]interface{}{
		"isActive":  false,
		"updatedAt": time.Now(),
	}))
	if err != nil {
		return fmt.Errorf("soft delete filter %s: %w", filterID, err)
	}
	return nil
}

type FilterService struct {
	store FilterStore
}

func NewFilterService(store FilterStore) *FilterService {
	return &FilterService{store: store}
}

func (s *FilterService) List(ctx context.Context, userID string) ([]models.Filter, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: userID is required", ErrInvalidInput)
	}
	filters, err := s.store.GetActiveFilters(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list filters: %w", err)
	}
	return filters, nil
}

func (s *FilterService) Create(ctx context.Context, userID string, filter *models.Filter) (*models.Filter, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: userID is required", ErrInvalidInput)
	}
	filter.UserID = userID
	id, err := s.store.CreateFilter(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("create filter: %w", err)
	}
	filter.ID = id
	return filter, nil
}

var allowedUpdateFields = map[string]bool{
	"origin": true, "destination": true, "priceMax": true, "mode": true,
	"startDate": true, "endDate": true, "passengers": true,
	"maxDurationHours": true, "maxStops": true,
	"preferredDeparture": true, "preferredArrival": true,
	"preferredAirlines": true, "excludedAirlines": true,
	"notifyEmail": true, "notifyPushToken": true, "isActive": true,
}

func sanitizeUpdates(updates map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{}, len(updates))
	for k, v := range updates {
		if allowedUpdateFields[k] {
			sanitized[k] = v
		}
	}
	return sanitized
}

func (s *FilterService) Update(ctx context.Context, userID, filterID string, updates map[string]interface{}) error {
	if userID == "" || filterID == "" {
		return fmt.Errorf("%w: userID and filterID are required", ErrInvalidInput)
	}
	existing, err := s.store.GetFilterByID(ctx, filterID)
	if err != nil {
		return fmt.Errorf("update filter: %w", err)
	}
	if existing.UserID != userID {
		return fmt.Errorf("%w: cannot update filter %s", ErrForbidden, filterID)
	}
	if err := s.store.UpdateFilter(ctx, filterID, sanitizeUpdates(updates)); err != nil {
		return fmt.Errorf("update filter: %w", err)
	}
	return nil
}

func (s *FilterService) Delete(ctx context.Context, userID, filterID string) error {
	if userID == "" || filterID == "" {
		return fmt.Errorf("%w: userID and filterID are required", ErrInvalidInput)
	}
	existing, err := s.store.GetFilterByID(ctx, filterID)
	if err != nil {
		return fmt.Errorf("delete filter: %w", err)
	}
	if existing.UserID != userID {
		return fmt.Errorf("%w: cannot delete filter %s", ErrForbidden, filterID)
	}
	if err := s.store.SoftDeleteFilter(ctx, filterID); err != nil {
		return fmt.Errorf("delete filter: %w", err)
	}
	return nil
}
