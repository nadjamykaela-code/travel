package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nadjamykaela-code/travel/pkg/models"
)

func TestSkyscannerSearcher_Search(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   interface{}
		filter     models.Filter
		wantErr    bool
		checkFn    func(*testing.T, *models.SearchResult)
	}{
		{
			name:       "successful search with quote",
			statusCode: http.StatusOK,
			response: models.IndicativeResponse{
				Status: "RESULT_STATUS_COMPLETE",
				Content: models.IndicativeContent{
					Results: models.IndicativeResults{
						Quotes: map[string]models.IndicativeQuote{
							"q1": {
								MinPrice: models.IndicativePrice{Amount: "2500.00", Unit: "PRICE_UNIT_UNSPECIFIED"},
								IsDirect: true,
								OutboundLeg: &models.IndicativeLeg{
									OriginPlaceID:      "place-1",
									DestinationPlaceID: "place-2",
									DepartureDateTime:  models.IndicativeDateTime{Year: 2026, Month: 8, Day: 15, Hour: 10, Minute: 0},
									MarketingCarrierID: "carrier-1",
								},
							},
						},
						Carriers: map[string]models.IndicativeCarrier{
							"carrier-1": {Name: "TAP Portugal", IATA: "TP"},
						},
						Places: map[string]models.IndicativePlace{
							"place-1": {Name: "São Paulo", IATA: "GRU", Type: "PLACE_TYPE_AIRPORT"},
							"place-2": {Name: "Lisbon", IATA: "LIS", Type: "PLACE_TYPE_AIRPORT"},
						},
					},
				},
			},
			filter:  models.Filter{ID: "f1", Origin: "GRU", Destination: "LIS", StartDate: time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)},
			wantErr: false,
			checkFn: func(t *testing.T, r *models.SearchResult) {
				if r.FilterID != "f1" {
					t.Errorf("FilterID = %q; want %q", r.FilterID, "f1")
				}
				if r.Origin != "GRU" {
					t.Errorf("Origin = %q; want %q", r.Origin, "GRU")
				}
				if len(r.Itineraries) != 1 {
					t.Fatalf("expected 1 itinerary, got %d", len(r.Itineraries))
				}
				if r.Itineraries[0].TotalPrice.Amount != 2500.0 {
					t.Errorf("price = %f; want 2500.0", r.Itineraries[0].TotalPrice.Amount)
				}
				if len(r.Itineraries[0].Segments) != 1 {
					t.Errorf("expected 1 segment, got %d", len(r.Itineraries[0].Segments))
				}
				if r.FetchedAt.IsZero() {
					t.Error("FetchedAt should not be zero")
				}
			},
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			response:   map[string]string{"error": "internal"},
			filter:     models.Filter{ID: "f1", Origin: "GRU", Destination: "LIS"},
			wantErr:    true,
		},
		{
			name:       "empty quotes",
			statusCode: http.StatusOK,
			response: models.IndicativeResponse{
				Status: "RESULT_STATUS_COMPLETE",
				Content: models.IndicativeContent{
					Results: models.IndicativeResults{
						Quotes:   map[string]models.IndicativeQuote{},
						Carriers: map[string]models.IndicativeCarrier{},
						Places:   map[string]models.IndicativePlace{},
					},
				},
			},
			filter:  models.Filter{ID: "f1", Origin: "GRU", Destination: "LIS"},
			wantErr: false,
			checkFn: func(t *testing.T, r *models.SearchResult) {
				if len(r.Itineraries) != 0 {
					t.Errorf("expected 0 itineraries, got %d", len(r.Itineraries))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("x-api-key") == "" {
					t.Error("missing x-api-key header")
				}
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}
				w.WriteHeader(tt.statusCode)
				switch v := tt.response.(type) {
				case string:
					w.Write([]byte(v))
				default:
					json.NewEncoder(w).Encode(v)
				}
			}))
			defer srv.Close()

			s := NewSkyscannerSearcher("test-key", srv.URL)
			result, err := s.Search(context.Background(), tt.filter)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Search() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.checkFn != nil && result != nil {
				tt.checkFn(t, result)
			}
		})
	}
}

func TestNewSkyscannerSearcher(t *testing.T) {
	s := NewSkyscannerSearcher("key", "https://partners.api.skyscanner.net/apiservices/v3")
	if s.apiKey != "key" {
		t.Errorf("apiKey = %q; want %q", s.apiKey, "key")
	}
	if s.client.Timeout != 30*time.Second {
		t.Errorf("timeout = %v; want %v", s.client.Timeout, 30*time.Second)
	}
	if s.market != "BR" {
		t.Errorf("market = %q; want BR", s.market)
	}
}

func TestSkyscannerSearcher_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.IndicativeResponse{})
	}))
	defer srv.Close()

	s := NewSkyscannerSearcher("key", srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := s.Search(ctx, models.Filter{ID: "f1", Origin: "GRU", Destination: "LIS"})
	if err == nil {
		t.Error("expected context deadline error, got nil")
	}
}
