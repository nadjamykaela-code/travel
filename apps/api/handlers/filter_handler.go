package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nadjamykaela-code/travel/apps/api/service"
	"github.com/nadjamykaela-code/travel/pkg/models"
)

type FilterHandler struct {
	svc *service.FilterService
}

func NewFilterHandler(svc *service.FilterService) *FilterHandler {
	return &FilterHandler{svc: svc}
}

type filterCreateRequest struct {
	Mode               string   `json:"mode"`
	Origin             string   `json:"origin"`
	Destination        string   `json:"destination"`
	PriceMax           float64  `json:"priceMax"`
	StartDate          string   `json:"startDate"`
	EndDate            string   `json:"endDate"`
	Passengers         int      `json:"passengers"`
	MaxDurationHours   int      `json:"maxDurationHours"`
	MaxStops           int      `json:"maxStops"`
	PreferredDeparture string   `json:"preferredDeparture"`
	PreferredArrival   string   `json:"preferredArrival"`
	PreferredAirlines  []string `json:"preferredAirlines"`
	ExcludedAirlines   []string `json:"excludedAirlines"`
	NotifyEmail        string   `json:"notifyEmail"`
	NotifyPushToken    string   `json:"notifyPushToken"`
}

func userIDFromContext(c *gin.Context) (string, bool) {
	uid, ok := c.Get("userID")
	if !ok {
		return "", false
	}
	return uid.(string), true
}

func (h *FilterHandler) GetFilters(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	filters, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao listar filtros"})
		return
	}
	if filters == nil {
		filters = []models.Filter{}
	}
	c.JSON(http.StatusOK, filters)
}

func (h *FilterHandler) CreateFilter(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	var req filterCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)

	filter := &models.Filter{
		Mode:               models.TravelMode(req.Mode),
		Origin:             req.Origin,
		Destination:        req.Destination,
		PriceMax:           req.PriceMax,
		StartDate:          startDate,
		EndDate:            endDate,
		Passengers:         req.Passengers,
		MaxDurationHours:   req.MaxDurationHours,
		MaxStops:           req.MaxStops,
		PreferredDeparture: req.PreferredDeparture,
		PreferredArrival:   req.PreferredArrival,
		PreferredAirlines:  req.PreferredAirlines,
		ExcludedAirlines:   req.ExcludedAirlines,
		NotifyEmail:        req.NotifyEmail,
		NotifyPushToken:    req.NotifyPushToken,
	}

	created, err := h.svc.Create(c.Request.Context(), userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao criar filtro"})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *FilterHandler) UpdateFilter(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	filterID := c.Param("id")
	if filterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id do filtro é obrigatório"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Update(c.Request.Context(), userID, filterID, updates); err != nil {
		if errors.Is(err, service.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "não autorizado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao atualizar filtro"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "filtro atualizado"})
}

func (h *FilterHandler) DeleteFilter(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	filterID := c.Param("id")
	if filterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id do filtro é obrigatório"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), userID, filterID); err != nil {
		if errors.Is(err, service.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "não autorizado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "falha ao deletar filtro"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "filtro desativado"})
}
