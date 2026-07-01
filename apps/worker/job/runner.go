package job

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/nadjamykaela-code/travel/pkg/models"
)

type FilterLister interface {
	ListActive(ctx context.Context) ([]models.Filter, error)
}

type FlightSearcher interface {
	Search(ctx context.Context, filter models.Filter) (*models.SearchResult, error)
}

type ResultFilter interface {
	Apply(itineraries []models.Itinerary, filter models.Filter) []models.Itinerary
}

type Notifier interface {
	Send(ctx context.Context, filter models.Filter, itineraries []models.Itinerary) error
}

type HistoryRecorder interface {
	Record(ctx context.Context, filter models.Filter, itineraries []models.Itinerary) error
}

type APILimiter interface {
	Allow() bool
	Remaining() int
}

type Runner struct {
	filters    FilterLister
	searcher   FlightSearcher
	resultFilt ResultFilter
	notifier   Notifier
	history    HistoryRecorder
	limiter    APILimiter
	logger     *slog.Logger
}

func NewRunner(
	filters FilterLister,
	searcher FlightSearcher,
	resultFilt ResultFilter,
	notifier Notifier,
	history HistoryRecorder,
	limiter APILimiter,
	logger *slog.Logger,
) *Runner {
	return &Runner{
		filters:    filters,
		searcher:   searcher,
		resultFilt: resultFilt,
		notifier:   notifier,
		history:    history,
		limiter:    limiter,
		logger:     logger,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	active, err := r.filters.ListActive(ctx)
	if err != nil {
		return fmt.Errorf("list active filters: %w", err)
	}

	if len(active) == 0 {
		r.logger.Info("no active filters found")
		return nil
	}

	r.logger.Info("processing filters", "count", len(active))

	var errs []error
	for _, filter := range active {
		if err := r.processFilter(ctx, filter); err != nil {
			errs = append(errs, fmt.Errorf("filter %s: %w", filter.ID, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("completed with %d error(s): %w", len(errs), errors.Join(errs...))
	}
	return nil
}

func (r *Runner) processFilter(ctx context.Context, filter models.Filter) error {
	log := r.logger.With("filter_id", filter.ID, "origin", filter.Origin, "dest", filter.Destination)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if !r.limiter.Allow() {
		log.Warn("API rate limit reached, skipping filter")
		return nil
	}

	log.Info("searching flights", "remaining", r.limiter.Remaining())

	result, err := r.searcher.Search(ctx, filter)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	matched := r.resultFilt.Apply(result.Itineraries, filter)
	if len(matched) == 0 {
		log.Info("no matching itineraries found")
		return nil
	}

	log.Info("matching itineraries found", "count", len(matched),
		"best_price", matched[0].TotalPrice.Amount,
		"currency", matched[0].TotalPrice.Currency)

	if err := r.notifier.Send(ctx, filter, matched); err != nil {
		return fmt.Errorf("notify: %w", err)
	}

	if err := r.history.Record(ctx, filter, matched); err != nil {
		return fmt.Errorf("record history: %w", err)
	}

	return nil
}
