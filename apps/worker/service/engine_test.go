package service

import (
	"testing"

	"github.com/nadjamykaela-code/travel/pkg/models"
)

func TestEngine_Apply(t *testing.T) {
	e := NewEngine()
	filter := models.Filter{PriceMax: 3000}

	its := []models.Itinerary{
		{ID: "it-1", TotalPrice: models.Price{Amount: 2000, Currency: "BRL"}, TotalDuration: 300},
		{ID: "it-2", TotalPrice: models.Price{Amount: 5000, Currency: "BRL"}, TotalDuration: 300},
	}

	result := e.Apply(its, filter)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].ID != "it-1" {
		t.Errorf("expected it-1, got %s", result[0].ID)
	}
}

func TestEngine_Apply_Empty(t *testing.T) {
	e := NewEngine()
	result := e.Apply(nil, models.Filter{})
	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}
