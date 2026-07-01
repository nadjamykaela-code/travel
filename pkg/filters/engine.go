package filters

import (
	"strings"
	"time"

	"github.com/nadjamykaela-code/travel/pkg/models"
)

// ApplyFilters verifica se um itinerário atende a todos os critérios do filtro
func ApplyFilters(itinerary models.Itinerary, filter models.Filter) bool {
	// 1. Preço máximo
	if itinerary.TotalPrice.Amount > filter.PriceMax {
		return false
	}

	// 2. Número de escalas
	if filter.MaxStops >= 0 && itinerary.Stops > filter.MaxStops {
		return false
	}

	// 3. Duração máxima (convertendo minutos para horas)
	if filter.MaxDurationHours > 0 {
		durationHours := itinerary.TotalDuration / 60
		if durationHours > filter.MaxDurationHours {
			return false
		}
	}

	// 4. Companhias excluídas
	if len(filter.ExcludedAirlines) > 0 {
		for _, seg := range itinerary.Segments {
			for _, excluded := range filter.ExcludedAirlines {
				if strings.EqualFold(seg.Carrier, excluded) || strings.EqualFold(seg.CarrierCode, excluded) {
					return false
				}
			}
		}
	}

	// 5. Companhias preferidas (se definidas, pelo menos uma deve aparecer)
	if len(filter.PreferredAirlines) > 0 {
		foundPreferred := false
		for _, seg := range itinerary.Segments {
			for _, pref := range filter.PreferredAirlines {
				if strings.EqualFold(seg.Carrier, pref) || strings.EqualFold(seg.CarrierCode, pref) {
					foundPreferred = true
					break
				}
			}
			if foundPreferred {
				break
			}
		}
		if !foundPreferred {
			return false
		}
	}

	// 6. Horário de partida (se definido)
	if filter.PreferredDeparture != "" && len(itinerary.Segments) > 0 {
		depTime, err := time.Parse("15:04", filter.PreferredDeparture)
		if err == nil {
			firstSeg := itinerary.Segments[0]
			hour, min, _ := firstSeg.DepartureAt.Clock()
			segDep := time.Date(0, 1, 1, hour, min, 0, 0, time.UTC)
			// Margem de 1 hora
			if segDep.Before(depTime.Add(-1*time.Hour)) || segDep.After(depTime.Add(1*time.Hour)) {
				return false
			}
		}
	}

	// 7. Horário de chegada (se definido)
	if filter.PreferredArrival != "" && len(itinerary.Segments) > 0 {
		arrTime, err := time.Parse("15:04", filter.PreferredArrival)
		if err == nil {
			lastSeg := itinerary.Segments[len(itinerary.Segments)-1]
			hour, min, _ := lastSeg.ArrivalAt.Clock()
			segArr := time.Date(0, 1, 1, hour, min, 0, 0, time.UTC)
			if segArr.Before(arrTime.Add(-1*time.Hour)) || segArr.After(arrTime.Add(1*time.Hour)) {
				return false
			}
		}
	}

	return true
}

// FilterResults filtra uma lista inteira de itinerários
func FilterResults(results []models.Itinerary, filter models.Filter) []models.Itinerary {
	var filtered []models.Itinerary
	for _, it := range results {
		if ApplyFilters(it, filter) {
			filtered = append(filtered, it)
		}
	}
	return filtered
}

// TruncateResults limita a quantidade de itinerários retornados.
// Firestore salva cada resultado no histórico; truncar reduz reads/writes.
func TruncateResults(results []models.Itinerary, max int) []models.Itinerary {
	if max <= 0 || len(results) <= max {
		return results
	}
	return results[:max]
}
