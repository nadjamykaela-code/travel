package models

import (
	"errors"
	"time"
)

type TravelMode string

const (
	ModeFlight TravelMode = "flight"
	ModeTrain  TravelMode = "train"
)

type Filter struct {
	ID                  string     `firestore:"id" json:"id"`
	UserID              string     `firestore:"userId" json:"userId"`
	Mode                TravelMode `firestore:"mode" json:"mode"`
	Origin              string     `firestore:"origin" json:"origin"`
	Destination         string     `firestore:"destination" json:"destination"`
	PriceMax            float64    `firestore:"priceMax" json:"priceMax"`
	StartDate           time.Time  `firestore:"startDate" json:"startDate"`
	EndDate             time.Time  `firestore:"endDate" json:"endDate"`
	Passengers          int        `firestore:"passengers" json:"passengers"`
	MaxDurationHours    int        `firestore:"maxDurationHours" json:"maxDurationHours"`
	MaxStops            int        `firestore:"maxStops" json:"maxStops"`
	PreferredDeparture  string     `firestore:"preferredDeparture" json:"preferredDeparture"`
	PreferredArrival    string     `firestore:"preferredArrival" json:"preferredArrival"`
	PreferredAirlines   []string   `firestore:"preferredAirlines" json:"preferredAirlines"`
	ExcludedAirlines    []string   `firestore:"excludedAirlines" json:"excludedAirlines"`
	NotifyEmail         string     `firestore:"notifyEmail" json:"notifyEmail"`
	NotifyPushToken     string     `firestore:"notifyPushToken" json:"notifyPushToken"`
	LastRunAt           time.Time  `firestore:"lastRunAt" json:"lastRunAt"`
	CreatedAt           time.Time  `firestore:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time  `firestore:"updatedAt" json:"updatedAt"`
	IsActive            bool       `firestore:"isActive" json:"isActive"`
}

func (f Filter) Validate() error {
	if f.Origin == "" {
		return errors.New("origin is required")
	}
	if f.Destination == "" {
		return errors.New("destination is required")
	}
	if f.Origin == f.Destination {
		return errors.New("origin and destination must be different")
	}
	if f.PriceMax <= 0 {
		return errors.New("priceMax must be positive")
	}
	if f.Passengers <= 0 {
		return errors.New("passengers must be positive")
	}
	if !f.StartDate.IsZero() && !f.EndDate.IsZero() && f.EndDate.Before(f.StartDate) {
		return errors.New("endDate must be after startDate")
	}
	return nil
}

func (f Filter) StorageEstimate() int {
	size := len(f.ID) + len(f.UserID) + len(f.Origin) + len(f.Destination)
	size += len(f.PreferredDeparture) + len(f.PreferredArrival) + len(f.NotifyEmail) + len(f.NotifyPushToken)
	for _, a := range f.PreferredAirlines {
		size += len(a)
	}
	for _, a := range f.ExcludedAirlines {
		size += len(a)
	}
	return size
}
