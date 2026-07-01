package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nadjamykaela-code/travel/pkg/models"
	"github.com/sony/gobreaker"
)

type SkyscannerSearcher struct {
	apiKey       string
	baseURL      string
	market       string
	locale       string
	currency     string
	client       *http.Client
	cb           *gobreaker.CircuitBreaker
	maxRetries   int
}

func NewSkyscannerSearcher(apiKey, baseURL string) *SkyscannerSearcher {
	return &SkyscannerSearcher{
		apiKey:   apiKey,
		baseURL:  baseURL,
		market:   "BR",
		locale:   "pt-BR",
		currency: "BRL",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cb: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "skyscanner",
			MaxRequests: 3,
			Interval:    60 * time.Second,
			Timeout:     30 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
				return counts.Requests >= 3 && failureRatio >= 0.6
			},
			OnStateChange: func(name string, from, to gobreaker.State) {
				fmt.Printf("circuit breaker %s: %s -> %s\n", name, from, to)
			},
		}),
		maxRetries: 2,
	}
}

func (s *SkyscannerSearcher) buildQueryLeg(filter models.Filter) models.IndicativeQueryLeg {
	originPlace := &models.PlaceRef{}
	destPlace := &models.PlaceRef{}

	if filter.Origin == "ANY" || filter.Origin == "" {
		originPlace.Anywhere = true
	} else {
		originPlace.QueryPlace = &models.QueryPlace{IATA: filter.Origin}
	}

	if filter.Destination == "ANY" || filter.Destination == "" {
		destPlace.Anywhere = true
	} else {
		destPlace.QueryPlace = &models.QueryPlace{IATA: filter.Destination}
	}

	leg := models.IndicativeQueryLeg{
		OriginPlace:      originPlace,
		DestinationPlace: destPlace,
	}

	if !filter.StartDate.IsZero() && !filter.EndDate.IsZero() {
		leg.DateRange = &models.DateRange{
			StartDate: &models.MonthYear{
				Year:  filter.StartDate.Year(),
				Month: int(filter.StartDate.Month()),
			},
			EndDate: &models.MonthYear{
				Year:  filter.EndDate.Year(),
				Month: int(filter.EndDate.Month()),
			},
		}
	} else if !filter.StartDate.IsZero() {
		leg.FixedDate = &models.FixedDate{
			Year:  filter.StartDate.Year(),
			Month: int(filter.StartDate.Month()),
			Day:   filter.StartDate.Day(),
		}
	} else {
		leg.Anytime = true
	}

	return leg
}

func (s *SkyscannerSearcher) Search(ctx context.Context, filter models.Filter) (*models.SearchResult, error) {
	body := models.IndicativeRequest{
		Query: models.IndicativeQuery{
			Market:    s.market,
			Locale:    s.locale,
			Currency:  s.currency,
			QueryLegs: []models.IndicativeQueryLeg{s.buildQueryLeg(filter)},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := s.baseURL + "/flights/indicative/search"

	result, err := s.cb.Execute(func() (interface{}, error) {
		var lastErr error
		for attempt := 0; attempt <= s.maxRetries; attempt++ {
			if attempt > 0 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(backoff(attempt)):
				}
			}

			reqCtx, reqCancel := context.WithTimeout(ctx, 10*time.Second)
			req, reqErr := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewReader(payload))
			if reqErr != nil {
				reqCancel()
				lastErr = fmt.Errorf("create request: %w", reqErr)
				continue
			}
			req.Header.Set("x-api-key", s.apiKey)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			resp, reqErr := s.client.Do(req)
			reqCancel()
			if reqErr != nil {
				lastErr = fmt.Errorf("http request (attempt %d): %w", attempt+1, reqErr)
				continue
			}

			respBody, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				lastErr = fmt.Errorf("read response (attempt %d): %w", attempt+1, readErr)
				continue
			}

			if resp.StatusCode == http.StatusOK {
				var apiResp models.IndicativeResponse
				if err := json.Unmarshal(respBody, &apiResp); err != nil {
					return nil, fmt.Errorf("decode response: %w", err)
				}
				result := s.mapResponse(apiResp, filter)
				return result, nil
			}

			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return nil, fmt.Errorf("skyscanner client error %d: %s", resp.StatusCode, string(respBody))
			}

			lastErr = fmt.Errorf("skyscanner error %d (attempt %d): %s", resp.StatusCode, attempt+1, string(respBody))
		}
		return nil, lastErr
	})
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	return result.(*models.SearchResult), nil
}

func backoff(attempt int) time.Duration {
	jitter := time.Duration(rand.Int63n(int64(500 * time.Millisecond)))
	delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second
	if delay > 30*time.Second {
		delay = 30 * time.Second
	}
	return delay + jitter
}

func (s *SkyscannerSearcher) mapResponse(apiResp models.IndicativeResponse, filter models.Filter) *models.SearchResult {
	result := &models.SearchResult{
		FilterID:    filter.ID,
		Origin:      filter.Origin,
		Destination: filter.Destination,
		FetchedAt:   time.Now(),
	}

	quotes := apiResp.Content.Results.Quotes
	carriers := apiResp.Content.Results.Carriers
	places := apiResp.Content.Results.Places

	result.TotalFound = len(quotes)

	switch apiResp.Status {
	case "RESULT_STATUS_COMPLETE":
	case "RESULT_STATUS_PARTIAL_ERROR":
		slog.Warn("indicative API returned partial error", "status", apiResp.Status)
	case "RESULT_STATUS_NO_RESULTS":
		return result
	default:
		slog.Warn("indicative API returned unknown status", "status", apiResp.Status)
	}

	i := 0
	for quoteID, quote := range quotes {
		it := models.Itinerary{
			ID:        quoteID,
			Stops:     0,
			IsTrain:   false,
			BookingURL: "",
		}

		if !quote.IsDirect {
			it.Stops = 1
		}

		amt, err := strconv.ParseFloat(quote.MinPrice.Amount, 64)
		if err == nil {
			switch strings.ToUpper(quote.MinPrice.Unit) {
			case "PRICE_UNIT_CENTS":
				amt = amt / 100
			case "PRICE_UNIT_WHOLE", "PRICE_UNIT_UNSPECIFIED", "":
			default:
				slog.Warn("unknown price unit", "unit", quote.MinPrice.Unit, "amount", quote.MinPrice.Amount)
			}
			it.TotalPrice = models.Price{Amount: amt, Currency: s.currency}
		}

		legs := []*models.IndicativeLeg{quote.OutboundLeg}
		if quote.InboundLeg != nil {
			legs = append(legs, quote.InboundLeg)
		}

		for _, leg := range legs {
			if leg == nil {
				continue
			}

			seg := models.Segment{
				Origin:      lookupIATA(places, leg.OriginPlaceID),
				Destination: lookupIATA(places, leg.DestinationPlaceID),
				Carrier:     lookupCarrierName(carriers, leg.MarketingCarrierID),
				CarrierCode: lookupCarrierIATA(carriers, leg.MarketingCarrierID),
				DepartureAt: skyscannerTime(leg.DepartureDateTime),
				Duration:    0,
			}
			it.Segments = append(it.Segments, seg)
			it.TotalDuration += seg.Duration
		}

		if len(it.Segments) > 1 {
			it.Stops = len(it.Segments) - 1
		}

		limit := 100
		if i >= limit {
			break
		}
		result.Itineraries = append(result.Itineraries, it)
		i++
	}

	return result
}

func lookupIATA(places map[string]models.IndicativePlace, id string) string {
	if p, ok := places[id]; ok && p.IATA != "" {
		return p.IATA
	}
	return id
}

func lookupCarrierName(carriers map[string]models.IndicativeCarrier, id string) string {
	if c, ok := carriers[id]; ok {
		return c.Name
	}
	return id
}

func lookupCarrierIATA(carriers map[string]models.IndicativeCarrier, id string) string {
	if c, ok := carriers[id]; ok && c.IATA != "" {
		return c.IATA
	}
	return id
}

func skyscannerTime(dt models.IndicativeDateTime) time.Time {
	return time.Date(dt.Year, time.Month(dt.Month), dt.Day, dt.Hour, dt.Minute, dt.Second, 0, time.UTC)
}
