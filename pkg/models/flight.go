package models

import "time"

type Price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type Segment struct {
	Origin        string    `json:"origin"`
	Destination   string    `json:"destination"`
	DepartureAt   time.Time `json:"departureAt"`
	ArrivalAt     time.Time `json:"arrivalAt"`
	Carrier       string    `json:"carrier"`
	CarrierCode   string    `json:"carrierCode"`
	Duration      int       `json:"duration"`
}

type Itinerary struct {
	ID            string    `json:"id"`
	Segments      []Segment `json:"segments"`
	TotalPrice    Price     `json:"totalPrice"`
	TotalDuration int       `json:"totalDuration"`
	Stops         int       `json:"stops"`
	IsTrain       bool      `json:"isTrain"`
	BookingURL    string    `json:"bookingUrl"`
}

type SearchResult struct {
	QueryID      string      `json:"queryId"`
	FilterID     string      `json:"filterId"`
	Origin       string      `json:"origin"`
	Destination  string      `json:"destination"`
	Itineraries  []Itinerary `json:"itineraries"`
	FetchedAt    time.Time   `json:"fetchedAt"`
	TotalFound   int         `json:"totalFound"`
}

// --- Skyscanner Indicative API v3 request types ---

type IndicativeRequest struct {
	Query IndicativeQuery `json:"query"`
}

type IndicativeQuery struct {
	Market               string                `json:"market"`
	Locale               string                `json:"locale"`
	Currency             string                `json:"currency"`
	QueryLegs            []IndicativeQueryLeg  `json:"queryLegs"`
	DateTimeGroupingType string                `json:"dateTimeGroupingType,omitempty"`
}

type IndicativeQueryLeg struct {
	OriginPlace      *PlaceRef  `json:"originPlace,omitempty"`
	DestinationPlace *PlaceRef  `json:"destinationPlace,omitempty"`
	DateRange        *DateRange `json:"dateRange,omitempty"`
	FixedDate        *FixedDate `json:"fixedDate,omitempty"`
	Anytime          bool       `json:"anytime"`
}

type PlaceRef struct {
	QueryPlace     *QueryPlace `json:"queryPlace,omitempty"`
	Anywhere       bool        `json:"anywhere"`
	AnywhereByCity bool        `json:"anywhereByCity"`
}

type QueryPlace struct {
	IATA     string `json:"iata,omitempty"`
	EntityID string `json:"entityId,omitempty"`
}

type DateRange struct {
	StartDate *MonthYear `json:"startDate,omitempty"`
	EndDate   *MonthYear `json:"endDate,omitempty"`
}

type FixedDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

type MonthYear struct {
	Year  int `json:"year"`
	Month int `json:"month"`
}

// --- Skyscanner Live Pricing API v3 request types ---

type LivePricingCreateRequest struct {
	Query LivePricingQuery `json:"query"`
}

type LivePricingQuery struct {
	Market      string               `json:"market"`
	Locale      string               `json:"locale"`
	Currency    string               `json:"currency"`
	QueryLegs   []LivePricingQueryLeg `json:"queryLegs"`
	CabinClass  string               `json:"cabinClass,omitempty"`
	Adults      int                  `json:"adults,omitempty"`
	ChildrenAges []int               `json:"childrenAges,omitempty"`
}

type LivePricingQueryLeg struct {
	OriginPlaceID      string `json:"originPlaceId"`
	DestinationPlaceID string `json:"destinationPlaceId"`
	Date               string `json:"date"`
}

type LivePricingCreateResponse struct {
	SessionToken string `json:"sessionToken"`
	Status       string `json:"status"`
}

// --- Skyscanner Live Pricing API v3 poll response types ---

type LivePricingPollResponse struct {
	Status     string                 `json:"status"`
	Content    LivePricingContent     `json:"content"`
}

type LivePricingContent struct {
	Results    LivePricingResults     `json:"results"`
}

type LivePricingResults struct {
	Itineraries  map[string]LiveItinerary  `json:"itineraries"`
	Carriers     map[string]IndicativeCarrier `json:"carriers"`
	Places       map[string]IndicativePlace   `json:"places"`
	Legs         map[string]LiveLeg           `json:"legs"`
	Stats        *LiveStats                   `json:"stats,omitempty"`
	SortingOptions []LiveSortingOption         `json:"sortingOptions,omitempty"`
}

type LiveItinerary struct {
	PricingOptions []LivePricingOption `json:"pricingOptions"`
	OutboundLegID  string              `json:"outboundLegId"`
	InboundLegID   string              `json:"inboundLegId,omitempty"`
}

type LivePricingOption struct {
	Price          LivePrice     `json:"price"`
	Items          []LiveItem    `json:"items"`
	AgentIDs       []string      `json:"agentIds"`
	DeepLink       string        `json:"deepLink,omitempty"`
}

type LivePrice struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Unit     string  `json:"unit,omitempty"`
}

type LiveItem struct {
	SegmentIDs []string `json:"segmentIds"`
	BookingURL string   `json:"bookingUrl,omitempty"`
	DeepLink   string   `json:"deepLink,omitempty"`
}

type LiveLeg struct {
	OriginPlaceID      string             `json:"originPlaceId"`
	DestinationPlaceID string             `json:"destinationPlaceId"`
	DepartureDateTime  IndicativeDateTime `json:"departureDateTime"`
	ArrivalDateTime    IndicativeDateTime `json:"arrivalDateTime"`
	MarketingCarrierID string             `json:"marketingCarrierId"`
	Duration           int                `json:"duration"`
	SegmentIDs         []string           `json:"segmentIds"`
	StopCount          int                `json:"stopCount"`
}

type LiveSegment struct {
	OriginPlaceID         string             `json:"originPlaceId"`
	DestinationPlaceID    string             `json:"destinationPlaceId"`
	DepartureDateTime     IndicativeDateTime `json:"departureDateTime"`
	ArrivalDateTime       IndicativeDateTime `json:"arrivalDateTime"`
	MarketingCarrierID    string             `json:"marketingCarrierId"`
	OperatingCarrierID    string             `json:"operatingCarrierId,omitempty"`
	Duration              int                `json:"duration"`
	FlightNumber          string             `json:"flightNumber"`
}

type LiveStats struct {
	MinPrice        LivePrice `json:"minPrice"`
	MaxPrice        LivePrice `json:"maxPrice"`
	MinDuration     int       `json:"minDuration"`
	MaxDuration     int       `json:"maxDuration"`
	ItineraryCount  int       `json:"itineraryCount"`
}

type LiveSortingOption struct {
	Type  string `json:"type"`
	Label string `json:"label"`
}

// --- Skyscanner Indicative API v3 response types ---

type IndicativeResponse struct {
	Status  string            `json:"status"`
	Content IndicativeContent `json:"content"`
}

type IndicativeContent struct {
	Results IndicativeResults `json:"results"`
}

type IndicativeResults struct {
	Quotes   map[string]IndicativeQuote   `json:"quotes"`
	Carriers map[string]IndicativeCarrier `json:"carriers"`
	Places   map[string]IndicativePlace   `json:"places"`
}

type IndicativeQuote struct {
	MinPrice    IndicativePrice `json:"minPrice"`
	IsDirect    bool            `json:"isDirect"`
	OutboundLeg *IndicativeLeg  `json:"outboundLeg,omitempty"`
	InboundLeg  *IndicativeLeg  `json:"inboundLeg,omitempty"`
}

type IndicativePrice struct {
	Amount string `json:"amount"`
	Unit   string `json:"unit"`
}

type IndicativeLeg struct {
	OriginPlaceID          string              `json:"originPlaceId"`
	DestinationPlaceID    string              `json:"destinationPlaceId"`
	DepartureDateTime     IndicativeDateTime  `json:"departureDateTime"`
	MarketingCarrierID    string              `json:"marketingCarrierId"`
	QuoteCreationTimestamp string             `json:"quoteCreationTimestamp,omitempty"`
}

type IndicativeDateTime struct {
	Year   int `json:"year"`
	Month  int `json:"month"`
	Day    int `json:"day"`
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
	Second int `json:"second"`
}

type IndicativeCarrier struct {
	Name        string `json:"name"`
	IATA        string `json:"iata"`
	ImageURL    string `json:"imageUrl,omitempty"`
	Icao        string `json:"icao,omitempty"`
	DisplayCode string `json:"displayCode,omitempty"`
}

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type IndicativePlace struct {
	Name        string      `json:"name"`
	IATA        string      `json:"iata"`
	Type        string      `json:"type"`
	EntityID    string      `json:"entityId,omitempty"`
	ParentID    string      `json:"parentId,omitempty"`
	Coordinates *Coordinates `json:"coordinates,omitempty"`
}
