package job

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/nadjamykaela-code/travel/pkg/models"
)

type mockFilterLister struct {
	filters []models.Filter
	err     error
}

func (m *mockFilterLister) ListActive(_ context.Context) ([]models.Filter, error) {
	return m.filters, m.err
}

type mockSearcher struct {
	result *models.SearchResult
	err    error
}

func (m *mockSearcher) Search(_ context.Context, _ models.Filter) (*models.SearchResult, error) {
	return m.result, m.err
}

type mockResultFilter struct {
	result []models.Itinerary
}

func (m *mockResultFilter) Apply(_ []models.Itinerary, _ models.Filter) []models.Itinerary {
	return m.result
}

type mockNotifier struct {
	called int
	err    error
}

func (m *mockNotifier) Send(_ context.Context, _ models.Filter, _ []models.Itinerary) error {
	m.called++
	return m.err
}

type mockHistoryRecorder struct {
	called int
	err    error
}

func (m *mockHistoryRecorder) Record(_ context.Context, _ models.Filter, _ []models.Itinerary) error {
	m.called++
	return m.err
}

type mockLimiter struct {
	allow     bool
	remaining int
}

func (m *mockLimiter) Allow() bool         { return m.allow }
func (m *mockLimiter) Remaining() int      { return m.remaining }

type discardHandler struct{}

func (discardHandler) Enabled(_ context.Context, _ slog.Level) bool  { return false }
func (discardHandler) Handle(_ context.Context, _ slog.Record) error  { return nil }
func (d discardHandler) WithAttrs(_ []slog.Attr) slog.Handler         { return d }
func (d discardHandler) WithGroup(_ string) slog.Handler              { return d }

func logger() *slog.Logger {
	return slog.New(discardHandler{})
}

func now() time.Time {
	return time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
}

var testFilter = models.Filter{
	ID:          "f1",
	UserID:      "user-1",
	Origin:      "GRU",
	Destination: "LIS",
	PriceMax:    3000,
	IsActive:    true,
}

var testItinerary = models.Itinerary{
	ID: "it-1",
	Segments: []models.Segment{{
		Origin:      "GRU",
		Destination: "LIS",
		DepartureAt: now(),
		ArrivalAt:   now().Add(10 * time.Hour),
		Carrier:     "TAP",
		Duration:    600,
	}},
	TotalPrice:    models.Price{Amount: 2500, Currency: "BRL"},
	TotalDuration: 600,
	Stops:         0,
	BookingURL:    "https://example.com/book",
}

func TestRunner_Run(t *testing.T) {
	tests := []struct {
		name    string
		filters *mockFilterLister
		searcher *mockSearcher
		filter  *mockResultFilter
		notifier *mockNotifier
		history *mockHistoryRecorder
		limiter *mockLimiter
		wantErr bool
	}{
		{
			name: "processes filter and finds matches",
			filters: &mockFilterLister{
				filters: []models.Filter{testFilter},
			},
			searcher: &mockSearcher{
				result: &models.SearchResult{
					Itineraries: []models.Itinerary{testItinerary},
				},
			},
			filter:  &mockResultFilter{result: []models.Itinerary{testItinerary}},
			notifier: &mockNotifier{},
			history: &mockHistoryRecorder{},
			limiter: &mockLimiter{allow: true, remaining: 99},
		},
		{
			name:    "no active filters",
			filters: &mockFilterLister{},
			searcher: &mockSearcher{},
			filter:  &mockResultFilter{},
			notifier: &mockNotifier{},
			history: &mockHistoryRecorder{},
			limiter: &mockLimiter{allow: true, remaining: 100},
		},
		{
			name: "skips filter when rate limited",
			filters: &mockFilterLister{
				filters: []models.Filter{testFilter},
			},
			searcher: &mockSearcher{},
			filter:  &mockResultFilter{},
			notifier: &mockNotifier{},
			history: &mockHistoryRecorder{},
			limiter: &mockLimiter{allow: false, remaining: 0},
		},
		{
			name: "no matches after filtering",
			filters: &mockFilterLister{
				filters: []models.Filter{testFilter},
			},
			searcher: &mockSearcher{
				result: &models.SearchResult{
					Itineraries: []models.Itinerary{testItinerary},
				},
			},
			filter:  &mockResultFilter{}, // empty result
			notifier: &mockNotifier{},
			history: &mockHistoryRecorder{},
			limiter: &mockLimiter{allow: true, remaining: 99},
		},
		{
			name: "filter listing error",
			filters: &mockFilterLister{
				err: assertAnError{},
			},
			searcher: &mockSearcher{},
			filter:  &mockResultFilter{},
			notifier: &mockNotifier{},
			history: &mockHistoryRecorder{},
			limiter: &mockLimiter{allow: true, remaining: 100},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewRunner(
				tt.filters, tt.searcher, tt.filter,
				tt.notifier, tt.history, tt.limiter,
				logger(),
			)

			err := runner.Run(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

type assertAnError struct{}

func (assertAnError) Error() string { return "test error" }
