package filters

import (
	"testing"
	"time"

	"github.com/nadjamykaela-code/travel/pkg/models"
)

func refTime(h, m int) time.Time {
	return time.Date(2026, 7, 1, h, m, 0, 0, time.UTC)
}

func TestApplyFilters(t *testing.T) {
	tests := []struct {
		name      string
		itinerary models.Itinerary
		filter    models.Filter
		want      bool
	}{
		{
			name: "passes all criteria",
			itinerary: models.Itinerary{
				Segments: []models.Segment{{
					Carrier:     "TAP",
					CarrierCode: "TP",
					DepartureAt: refTime(8, 0),
					ArrivalAt:   refTime(18, 0),
				}},
				TotalPrice:    models.Price{Amount: 2500, Currency: "BRL"},
				TotalDuration: 600,
				Stops:         0,
			},
			filter: models.Filter{
				PriceMax:         3000,
				MaxStops:         1,
				MaxDurationHours: 12,
			},
			want: true,
		},
		{
			name: "price exceeds max",
			itinerary: models.Itinerary{
				TotalPrice: models.Price{Amount: 4000, Currency: "BRL"},
			},
			filter: models.Filter{PriceMax: 3000},
			want:   false,
		},
		{
			name: "too many stops",
			itinerary: models.Itinerary{
				TotalPrice: models.Price{Amount: 2000, Currency: "BRL"},
				Stops:      2,
			},
			filter: models.Filter{PriceMax: 3000, MaxStops: 1},
			want:   false,
		},
		{
			name: "ignores stops when MaxStops < 0",
			itinerary: models.Itinerary{
				TotalPrice: models.Price{Amount: 2000, Currency: "BRL"},
				Stops:      5,
			},
			filter: models.Filter{PriceMax: 3000, MaxStops: -1},
			want:   true,
		},
		{
			name: "duration exceeds max",
			itinerary: models.Itinerary{
				TotalPrice:    models.Price{Amount: 2000, Currency: "BRL"},
				TotalDuration: 720,
			},
			filter: models.Filter{PriceMax: 3000, MaxDurationHours: 10},
			want:   false,
		},
		{
			name: "excluded airline matches carrier name",
			itinerary: models.Itinerary{
				Segments:    []models.Segment{{Carrier: "GOL", CarrierCode: "G3"}},
				TotalPrice:  models.Price{Amount: 2000, Currency: "BRL"},
				TotalDuration: 300,
			},
			filter: models.Filter{PriceMax: 3000, ExcludedAirlines: []string{"gol"}},
			want:   false,
		},
		{
			name: "excluded airline matches carrier code",
			itinerary: models.Itinerary{
				Segments:    []models.Segment{{Carrier: "GOL Linhas Aereas", CarrierCode: "G3"}},
				TotalPrice:  models.Price{Amount: 2000, Currency: "BRL"},
				TotalDuration: 300,
			},
			filter: models.Filter{PriceMax: 3000, ExcludedAirlines: []string{"g3"}},
			want:   false,
		},
		{
			name: "preferred airline found",
			itinerary: models.Itinerary{
				Segments:    []models.Segment{{Carrier: "LATAM", CarrierCode: "LA"}},
				TotalPrice:  models.Price{Amount: 2000, Currency: "BRL"},
				TotalDuration: 300,
			},
			filter: models.Filter{PriceMax: 3000, PreferredAirlines: []string{"LATAM"}},
			want:   true,
		},
		{
			name: "preferred airline not found",
			itinerary: models.Itinerary{
				Segments:    []models.Segment{{Carrier: "AZUL", CarrierCode: "AD"}},
				TotalPrice:  models.Price{Amount: 2000, Currency: "BRL"},
				TotalDuration: 300,
			},
			filter: models.Filter{PriceMax: 3000, PreferredAirlines: []string{"LATAM"}},
			want:   false,
		},
		{
			name: "preferred departure time within 1h",
			itinerary: models.Itinerary{
				Segments: []models.Segment{{
					DepartureAt: refTime(9, 30),
					ArrivalAt:   refTime(18, 0),
				}},
				TotalPrice:    models.Price{Amount: 2000, Currency: "BRL"},
				TotalDuration: 300,
			},
			filter: models.Filter{PriceMax: 3000, PreferredDeparture: "10:00"},
			want:   true,
		},
		{
			name: "preferred departure time outside 1h",
			itinerary: models.Itinerary{
				Segments: []models.Segment{{
					DepartureAt: refTime(7, 0),
					ArrivalAt:   refTime(18, 0),
				}},
				TotalPrice:    models.Price{Amount: 2000, Currency: "BRL"},
				TotalDuration: 300,
			},
			filter: models.Filter{PriceMax: 3000, PreferredDeparture: "10:00"},
			want:   false,
		},
		{
			name: "preferred arrival time within 1h",
			itinerary: models.Itinerary{
				Segments: []models.Segment{{
					DepartureAt: refTime(8, 0),
					ArrivalAt:   refTime(18, 0),
				}},
				TotalPrice:    models.Price{Amount: 2000, Currency: "BRL"},
				TotalDuration: 600,
			},
			filter: models.Filter{PriceMax: 3000, PreferredArrival: "17:30"},
			want:   true,
		},
		{
			name: "empty itinerary segments with preferred time should not panic",
			itinerary: models.Itinerary{
				Segments:      []models.Segment{},
				TotalPrice:    models.Price{Amount: 2000, Currency: "BRL"},
				TotalDuration: 300,
			},
			filter: models.Filter{PriceMax: 3000, PreferredDeparture: "10:00"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyFilters(tt.itinerary, tt.filter)
			if got != tt.want {
				t.Errorf("ApplyFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterResults(t *testing.T) {
	its := []models.Itinerary{
		{ID: "it-1", TotalPrice: models.Price{Amount: 1000, Currency: "BRL"}, TotalDuration: 300},
		{ID: "it-2", TotalPrice: models.Price{Amount: 5000, Currency: "BRL"}, TotalDuration: 300},
		{ID: "it-3", TotalPrice: models.Price{Amount: 2000, Currency: "BRL"}, TotalDuration: 300},
	}
	got := FilterResults(its, models.Filter{PriceMax: 2500})
	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %d", len(got))
	}
	if got[0].ID != "it-1" || got[1].ID != "it-3" {
		t.Errorf("unexpected selection: %+v", got)
	}
}

func TestTruncateResults(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		max      int
		expected int
	}{
		{"no truncation needed", 3, 5, 3},
		{"truncates to max", 10, 3, 3},
		{"zero max returns all", 5, 0, 5},
		{"negative max returns all", 5, -1, 5},
		{"empty slice", 0, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := make([]models.Itinerary, tt.input)
			for i := range results {
				results[i] = models.Itinerary{ID: string(rune('a' + i))}
			}
			got := TruncateResults(results, tt.max)
			if len(got) != tt.expected {
				t.Errorf("TruncateResults() len = %d, want %d", len(got), tt.expected)
			}
		})
	}
}
