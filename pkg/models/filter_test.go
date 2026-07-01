package models

import (
	"errors"
	"testing"
	"time"
)

func TestFilterValidate(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		filter  Filter
		wantErr error
	}{
		{
			name: "valid filter",
			filter: Filter{
				Origin:      "GRU",
				Destination: "LIS",
				PriceMax:    3000,
				Passengers:  1,
				StartDate:   now,
				EndDate:     now.Add(7 * 24 * time.Hour),
			},
			wantErr: nil,
		},
		{
			name: "missing origin",
			filter: Filter{
				Destination: "LIS",
				PriceMax:    3000,
				Passengers:  1,
			},
			wantErr: errors.New("origin is required"),
		},
		{
			name: "missing destination",
			filter: Filter{
				Origin:   "GRU",
				PriceMax: 3000,
				Passengers: 1,
			},
			wantErr: errors.New("destination is required"),
		},
		{
			name: "same origin and destination",
			filter: Filter{
				Origin:      "GRU",
				Destination: "GRU",
				PriceMax:    3000,
				Passengers:  1,
			},
			wantErr: errors.New("origin and destination must be different"),
		},
		{
			name: "zero price max",
			filter: Filter{
				Origin:      "GRU",
				Destination: "LIS",
				PriceMax:    0,
				Passengers:  1,
			},
			wantErr: errors.New("priceMax must be positive"),
		},
		{
			name: "zero passengers",
			filter: Filter{
				Origin:      "GRU",
				Destination: "LIS",
				PriceMax:    3000,
			},
			wantErr: errors.New("passengers must be positive"),
		},
		{
			name: "end before start",
			filter: Filter{
				Origin:      "GRU",
				Destination: "LIS",
				PriceMax:    3000,
				Passengers:  1,
				StartDate:   now.Add(7 * 24 * time.Hour),
				EndDate:     now,
			},
			wantErr: errors.New("endDate must be after startDate"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if tt.wantErr == nil && err != nil {
				t.Errorf("Validate() = %v, want nil", err)
			}
			if tt.wantErr != nil && err == nil {
				t.Errorf("Validate() = nil, want %v", tt.wantErr)
			}
			if tt.wantErr != nil && err != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("Validate() = %q, want %q", err.Error(), tt.wantErr.Error())
			}
		})
	}
}

func TestStorageEstimate(t *testing.T) {
	f := Filter{
		ID:               "abc123",
		UserID:           "user-1",
		Origin:           "GRU",
		Destination:      "LIS",
		PreferredAirlines: []string{"TAP", "LATAM"},
		ExcludedAirlines:  []string{"GOL"},
	}
	est := f.StorageEstimate()
	if est <= 0 {
		t.Errorf("StorageEstimate() = %d, want > 0", est)
	}
}
