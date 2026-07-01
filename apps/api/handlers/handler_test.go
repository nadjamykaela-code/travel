package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nadjamykaela-code/travel/apps/api/service"
	"github.com/nadjamykaela-code/travel/pkg/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

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
	return "new-id-123", nil
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
	return nil, service.ErrNotFound
}

func (m *mockFilterStore) UpdateFilter(_ context.Context, _ string, _ map[string]interface{}) error {
	return m.err
}

func (m *mockFilterStore) SoftDeleteFilter(_ context.Context, _ string) error {
	return m.err
}

type mockAuthService struct {
	userID string
	err    error
}

func (m *mockAuthService) VerifyToken(_ context.Context, _ string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.userID, nil
}

func setupRouter(store *mockFilterStore, auth *mockAuthService) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	filterSvc := service.NewFilterService(store)
	filterHandler := NewFilterHandler(filterSvc)

	api := r.Group("/api")
	api.Use(newTestAuthMiddleware(auth))
	{
		api.GET("/filters", filterHandler.GetFilters)
		api.POST("/filters", filterHandler.CreateFilter)
		api.PUT("/filters/:id", filterHandler.UpdateFilter)
		api.DELETE("/filters/:id", filterHandler.DeleteFilter)
	}
	return r
}

func newTestAuthMiddleware(auth *mockAuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, err := auth.VerifyToken(c.Request.Context(), "")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
			return
		}
		c.Set("userID", uid)
		c.Next()
	}
}

func TestFilterHandler_GetFilters(t *testing.T) {
	tests := []struct {
		name       string
		store      *mockFilterStore
		auth       *mockAuthService
		wantStatus int
		wantLen    int
	}{
		{
			name: "returns active filters for authenticated user",
			store: &mockFilterStore{
				filters: []models.Filter{
					{ID: "f1", UserID: "user-1", Origin: "GRU", Destination: "LIS", PriceMax: 3000, IsActive: true},
					{ID: "f2", UserID: "user-1", Origin: "POA", Destination: "MIA", PriceMax: 2000, IsActive: true},
					{ID: "f3", UserID: "user-2", Origin: "GIG", Destination: "MAD", PriceMax: 4000, IsActive: true},
				},
			},
			auth:       &mockAuthService{userID: "user-1"},
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name: "returns empty list when no filters",
			store: &mockFilterStore{
				filters: []models.Filter{},
			},
			auth:       &mockAuthService{userID: "user-3"},
			wantStatus: http.StatusOK,
			wantLen:    0,
		},
		{
			name:       "returns 401 when not authenticated",
			store:      &mockFilterStore{},
			auth:       &mockAuthService{err: service.ErrForbidden},
			wantStatus: http.StatusUnauthorized,
			wantLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRouter(tt.store, tt.auth)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/filters", nil)
			if tt.auth.err == nil {
				req.Header.Set("Authorization", "Bearer test-token")
			}
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d", w.Code, tt.wantStatus)
			}
			if w.Code == http.StatusOK {
				var filters []models.Filter
				if err := json.Unmarshal(w.Body.Bytes(), &filters); err != nil {
					t.Fatalf("unmarshal response: %v", err)
				}
				if len(filters) != tt.wantLen {
					t.Errorf("len(filters) = %d; want %d", len(filters), tt.wantLen)
				}
			}
		})
	}
}

func TestFilterHandler_CreateFilter(t *testing.T) {
	tests := []struct {
		name       string
		store      *mockFilterStore
		auth       *mockAuthService
		body       string
		wantStatus int
	}{
		{
			name:  "creates filter successfully",
			store: &mockFilterStore{},
			auth:  &mockAuthService{userID: "user-1"},
			body: `{
				"origin": "GRU",
				"destination": "LIS",
				"priceMax": 3000,
				"passengers": 2
			}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:  "returns 400 on invalid body",
			store: &mockFilterStore{},
			auth:  &mockAuthService{userID: "user-1"},
			body: `{invalid json`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 401 when not authenticated",
			store:      &mockFilterStore{},
			auth:       &mockAuthService{err: service.ErrForbidden},
			body:       `{"origin": "GRU"}`,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRouter(tt.store, tt.auth)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/filters",
				strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			if tt.auth.err == nil {
				req.Header.Set("Authorization", "Bearer test-token")
			}
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestFilterHandler_UpdateFilter(t *testing.T) {
	tests := []struct {
		name       string
		store      *mockFilterStore
		auth       *mockAuthService
		filterID   string
		body       string
		wantStatus int
	}{
		{
			name: "updates own filter successfully",
			store: &mockFilterStore{
				filters: []models.Filter{
					{ID: "f1", UserID: "user-1", Origin: "GRU", IsActive: true},
				},
			},
			auth:       &mockAuthService{userID: "user-1"},
			filterID:   "f1",
			body:       `{"priceMax": 2500}`,
			wantStatus: http.StatusOK,
		},
		{
			name: "returns 403 when updating another user's filter",
			store: &mockFilterStore{
				filters: []models.Filter{
					{ID: "f2", UserID: "user-2", Origin: "POA", IsActive: true},
				},
			},
			auth:       &mockAuthService{userID: "user-1"},
			filterID:   "f2",
			body:       `{"priceMax": 1500}`,
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRouter(tt.store, tt.auth)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPut, "/api/filters/"+tt.filterID,
				strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-token")
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestFilterHandler_DeleteFilter(t *testing.T) {
	tests := []struct {
		name       string
		store      *mockFilterStore
		auth       *mockAuthService
		filterID   string
		wantStatus int
	}{
		{
			name: "deletes own filter successfully",
			store: &mockFilterStore{
				filters: []models.Filter{
					{ID: "f1", UserID: "user-1", Origin: "GRU", IsActive: true},
				},
			},
			auth:       &mockAuthService{userID: "user-1"},
			filterID:   "f1",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupRouter(tt.store, tt.auth)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodDelete, "/api/filters/"+tt.filterID, nil)
			req.Header.Set("Authorization", "Bearer test-token")
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
