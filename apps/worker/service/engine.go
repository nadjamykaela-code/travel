package service

import (
	"github.com/nadjamykaela-code/travel/pkg/filters"
	"github.com/nadjamykaela-code/travel/pkg/models"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Apply(itineraries []models.Itinerary, filter models.Filter) []models.Itinerary {
	return filters.FilterResults(itineraries, filter)
}
