package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nadjamykaela-code/travel/pkg/clients"
)

type PlaceHandler struct {
	client *clients.PlaceSearchClient
}

func NewPlaceHandler(client *clients.PlaceSearchClient) *PlaceHandler {
	return &PlaceHandler{client: client}
}

func (h *PlaceHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parâmetro 'q' é obrigatório"})
		return
	}

	if len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mínimo de 2 caracteres"})
		return
	}

	places, err := h.client.Search(c.Request.Context(), query)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "place search failed", "query", query, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar lugares"})
		return
	}

	if places == nil {
		places = []clients.Place{}
	}

	c.JSON(http.StatusOK, places)
}
