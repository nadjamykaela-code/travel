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
	"time"

	"github.com/nadjamykaela-code/travel/pkg/models"
	"github.com/sony/gobreaker"
)

type LivePricingClient struct {
	apiKey     string
	baseURL    string
	market     string
	locale     string
	currency   string
	client     *http.Client
	cb         *gobreaker.CircuitBreaker
	maxRetries int
	pollDelay  time.Duration
	pollMax    int
}

func NewLivePricingClient(apiKey, baseURL string) *LivePricingClient {
	return &LivePricingClient{
		apiKey:   apiKey,
		baseURL:  baseURL,
		market:   "BR",
		locale:   "pt-BR",
		currency: "BRL",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cb: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "live-pricing",
			MaxRequests: 3,
			Interval:    60 * time.Second,
			Timeout:     30 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
				return counts.Requests >= 3 && failureRatio >= 0.6
			},
			OnStateChange: func(name string, from, to gobreaker.State) {
				slog.Info("live pricing circuit breaker state change", "from", from, "to", to)
			},
		}),
		maxRetries: 2,
		pollDelay:  1 * time.Second,
		pollMax:    30,
	}
}

func (c *LivePricingClient) Search(ctx context.Context, filter models.Filter) (*models.SearchResult, error) {
	createReq := c.buildCreateRequest(filter)
	payload, err := json.Marshal(createReq)
	if err != nil {
		return nil, fmt.Errorf("marshal create request: %w", err)
	}

	createURL := c.baseURL + "/flights/live/search/create"

	sessionToken, err := c.createSession(ctx, createURL, payload)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	pollURL := c.baseURL + "/flights/live/search/poll/" + sessionToken

	pollResp, err := c.pollResults(ctx, pollURL)
	if err != nil {
		return nil, fmt.Errorf("poll results: %w", err)
	}

	result := c.mapResponse(pollResp, filter)
	return result, nil
}

func (c *LivePricingClient) buildCreateRequest(filter models.Filter) models.LivePricingCreateRequest {
	date := ""
	if !filter.StartDate.IsZero() {
		date = filter.StartDate.Format("2006-01-02")
	}

	return models.LivePricingCreateRequest{
		Query: models.LivePricingQuery{
			Market:    c.market,
			Locale:    c.locale,
			Currency:  c.currency,
			CabinClass: "ECONOMY",
			Adults:    filter.Passengers,
			QueryLegs: []models.LivePricingQueryLeg{
				{
					OriginPlaceID:      filter.Origin,
					DestinationPlaceID: filter.Destination,
					Date:               date,
				},
			},
		},
	}
}

func (c *LivePricingClient) createSession(ctx context.Context, url string, payload []byte) (string, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		var lastErr error
		for attempt := 0; attempt <= c.maxRetries; attempt++ {
			if attempt > 0 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(lpBackoff(attempt)):
				}
			}

			reqCtx, reqCancel := context.WithTimeout(ctx, 10*time.Second)
			req, reqErr := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewReader(payload))
			if reqErr != nil {
				reqCancel()
				lastErr = fmt.Errorf("create request: %w", reqErr)
				continue
			}
			req.Header.Set("x-api-key", c.apiKey)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			resp, reqErr := c.client.Do(req)
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
				var createResp models.LivePricingCreateResponse
				if err := json.Unmarshal(respBody, &createResp); err != nil {
					return nil, fmt.Errorf("decode create response: %w", err)
				}
				if createResp.SessionToken == "" {
					return nil, fmt.Errorf("empty session token in response")
				}
				return createResp.SessionToken, nil
			}

			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return nil, fmt.Errorf("live pricing client error %d: %s", resp.StatusCode, string(respBody))
			}

			lastErr = fmt.Errorf("live pricing error %d (attempt %d): %s", resp.StatusCode, attempt+1, string(respBody))
		}
		return nil, lastErr
	})
	if err != nil {
		return "", fmt.Errorf("create session failed: %w", err)
	}
	return result.(string), nil
}

func (c *LivePricingClient) pollResults(ctx context.Context, url string) (*models.LivePricingPollResponse, error) {
	for i := 0; i < c.pollMax; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(c.pollDelay):
		}

		reqCtx, reqCancel := context.WithTimeout(ctx, 10*time.Second)
		req, reqErr := http.NewRequestWithContext(reqCtx, http.MethodPost, url, nil)
		if reqErr != nil {
			reqCancel()
			return nil, fmt.Errorf("create poll request: %w", reqErr)
		}
		req.Header.Set("x-api-key", c.apiKey)
		req.Header.Set("Accept", "application/json")

		resp, reqErr := c.client.Do(req)
		reqCancel()
		if reqErr != nil {
			return nil, fmt.Errorf("poll request failed: %w", reqErr)
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read poll response: %w", readErr)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("poll error %d: %s", resp.StatusCode, string(respBody))
		}

		var pollResp models.LivePricingPollResponse
		if err := json.Unmarshal(respBody, &pollResp); err != nil {
			return nil, fmt.Errorf("decode poll response: %w", err)
		}

		switch pollResp.Status {
		case "RESULT_STATUS_COMPLETE":
			return &pollResp, nil
		case "RESULT_STATUS_PENDING":
			continue
		case "RESULT_STATUS_NO_RESULTS":
			return &pollResp, nil
		case "RESULT_STATUS_PARTIAL_ERROR":
			slog.Warn("live pricing partial error during poll")
			return &pollResp, nil
		default:
			slog.Warn("live pricing unknown poll status", "status", pollResp.Status)
			return &pollResp, nil
		}
	}

	return nil, fmt.Errorf("poll timeout after %d attempts", c.pollMax)
}

func (c *LivePricingClient) mapResponse(pollResp *models.LivePricingPollResponse, filter models.Filter) *models.SearchResult {
	result := &models.SearchResult{
		FilterID:    filter.ID,
		Origin:      filter.Origin,
		Destination: filter.Destination,
		FetchedAt:   time.Now(),
	}

	if pollResp.Content.Results.Stats != nil {
		result.TotalFound = pollResp.Content.Results.Stats.ItineraryCount
	}

	carriers := pollResp.Content.Results.Carriers
	places := pollResp.Content.Results.Places
	legs := pollResp.Content.Results.Legs

	i := 0
	for itineraryID, itin := range pollResp.Content.Results.Itineraries {
		it := models.Itinerary{
			ID:      itineraryID,
			Stops:   0,
			IsTrain: false,
		}

		if len(itin.PricingOptions) > 0 {
			opt := itin.PricingOptions[0]
			priceAmt := opt.Price.Amount
			if opt.Price.Unit == "PRICE_UNIT_CENTS" {
				priceAmt = priceAmt / 100
			}
			it.TotalPrice = models.Price{Amount: priceAmt, Currency: opt.Price.Currency}
			if it.TotalPrice.Currency == "" {
				it.TotalPrice.Currency = c.currency
			}

			if len(opt.Items) > 0 {
				it.BookingURL = opt.Items[0].BookingURL
				if it.BookingURL == "" {
					it.BookingURL = opt.Items[0].DeepLink
				}
			}
			if it.BookingURL == "" {
				it.BookingURL = opt.DeepLink
			}
		}

		legIDs := []string{itin.OutboundLegID}
		if itin.InboundLegID != "" {
			legIDs = append(legIDs, itin.InboundLegID)
		}

		for _, legID := range legIDs {
			leg, ok := legs[legID]
			if !ok {
				continue
			}

			seg := models.Segment{
				Origin:      lookupIATA(places, leg.OriginPlaceID),
				Destination: lookupIATA(places, leg.DestinationPlaceID),
				Carrier:     lookupCarrierName(carriers, leg.MarketingCarrierID),
				CarrierCode: lookupCarrierIATA(carriers, leg.MarketingCarrierID),
				DepartureAt: skyscannerTime(leg.DepartureDateTime),
				ArrivalAt:   skyscannerTime(leg.ArrivalDateTime),
				Duration:    leg.Duration,
			}
			it.Segments = append(it.Segments, seg)
			it.TotalDuration += leg.Duration
			it.Stops += leg.StopCount
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

func lpBackoff(attempt int) time.Duration {
	jitter := time.Duration(rand.Int63n(int64(500 * time.Millisecond)))
	delay := time.Duration(math.Pow(2, float64(attempt))) * time.Second
	if delay > 30*time.Second {
		delay = 30 * time.Second
	}
	return delay + jitter
}
