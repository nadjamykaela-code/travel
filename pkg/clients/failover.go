package clients

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/nadjamykaela-code/travel/pkg/models"
)

type FailoverSearcher struct {
	livePricing *LivePricingClient
	indicative  *SkyscannerSearcher
}

func NewFailoverSearcher(livePricing *LivePricingClient, indicative *SkyscannerSearcher) *FailoverSearcher {
	return &FailoverSearcher{
		livePricing: livePricing,
		indicative:  indicative,
	}
}

func (f *FailoverSearcher) Search(ctx context.Context, filter models.Filter) (*models.SearchResult, error) {
	result, err := f.livePricing.Search(ctx, filter)
	if err == nil {
		slog.Info("live pricing search succeeded",
			"filter_id", filter.ID,
			"itineraries", result.TotalFound,
		)
		return result, nil
	}

	slog.Warn("live pricing failed, falling back to indicative",
		"filter_id", filter.ID,
		"error", err,
	)

	result, fallbackErr := f.indicative.Search(ctx, filter)
	if fallbackErr != nil {
		return nil, fmt.Errorf("live pricing: %w; indicative fallback: %w", err, fallbackErr)
	}

	slog.Info("indicative fallback search succeeded",
		"filter_id", filter.ID,
		"itineraries", result.TotalFound,
	)
	return result, nil
}

func (f *FailoverSearcher) IsLivePricingAvailable(ctx context.Context, filter models.Filter) bool {
	_, err := f.livePricing.Search(ctx, filter)
	return err == nil
}

var _ FlightSearcher = (*FailoverSearcher)(nil)

type FlightSearcher interface {
	Search(ctx context.Context, filter models.Filter) (*models.SearchResult, error)
}

func IsNoResults(err error) bool {
	if err == nil {
		return false
	}
	var partialErr interface{ NoResults() bool }
	if errors.As(err, &partialErr) {
		return partialErr.NoResults()
	}
	return false
}
