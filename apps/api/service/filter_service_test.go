package service

import (
	"context"
	"errors"
	"testing"

	"github.com/nadjamykaela-code/travel/pkg/models"
)

type mockFilterStore struct {
	filters []models.Filter
	err     error
}

func (m *mockFilterStore) GetActiveFilters(_ context.Context, userID string) ([]models.Filter, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []models.Filter
	for _, f := range m.filters {
		if f.UserID == userID && f.IsActive {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *mockFilterStore) CreateFilter(_ context.Context, filter *models.Filter) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return "new-id", nil
}

func (m *mockFilterStore) GetFilterByID(_ context.Context, filterID string) (*models.Filter, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, f := range m.filters {
		if f.ID == filterID {
			return &f, nil
		}
	}
		return nil, errors.New("not found")
}

func (m *mockFilterStore) UpdateFilter(_ context.Context, filterID string, updates map[string]interface{}) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *mockFilterStore) SoftDeleteFilter(_ context.Context, filterID string) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func TestFilterService_List(t *testing.T) {
	tests := []struct {
		name    string
		store   *mockFilterStore
		userID  string
		want    int
		wantErr bool
	}{
		{
			name: "lists active filters for user",
			store: &mockFilterStore{
				filters: []models.Filter{
					{ID: "1", UserID: "u1", IsActive: true},
					{ID: "2", UserID: "u1", IsActive: false},
					{ID: "3", UserID: "u2", IsActive: true},
				},
			},
			userID: "u1",
			want:   1,
		},
		{
			name:    "empty userID returns error",
			store:   &mockFilterStore{err: errors.New("empty userID")},
			wantErr: true,
		},
		{
			name: "store error propagates",
			store: &mockFilterStore{
				err: errors.New("db error"),
			},
			userID:  "u1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewFilterService(tt.store)
			filters, err := svc.List(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("List() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(filters) != tt.want {
				t.Errorf("List() returned %d filters, want %d", len(filters), tt.want)
			}
		})
	}
}

func TestFilterService_Create(t *testing.T) {
	store := &mockFilterStore{}
	svc := NewFilterService(store)

	f, err := svc.Create(context.Background(), "u1", &models.Filter{Origin: "GRU"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if f.ID != "new-id" {
		t.Errorf("ID = %q; want %q", f.ID, "new-id")
	}
	if f.UserID != "u1" {
		t.Errorf("UserID = %q; want %q", f.UserID, "u1")
	}
}

func TestFilterService_Update(t *testing.T) {
	store := &mockFilterStore{
		filters: []models.Filter{
			{ID: "f1", UserID: "u1"},
		},
	}
	svc := NewFilterService(store)

	err := svc.Update(context.Background(), "u1", "f1", map[string]interface{}{"origin": "GRU"})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	err = svc.Update(context.Background(), "u2", "f1", map[string]interface{}{})
	if err == nil {
		t.Error("expected forbidden error, got nil")
	}
}

func TestFilterService_Delete(t *testing.T) {
	store := &mockFilterStore{
		filters: []models.Filter{
			{ID: "f1", UserID: "u1"},
		},
	}
	svc := NewFilterService(store)

	err := svc.Delete(context.Background(), "u1", "f1")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	err = svc.Delete(context.Background(), "u2", "f1")
	if err == nil {
		t.Error("expected forbidden error, got nil")
	}
}
