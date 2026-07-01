package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nadjamykaela-code/travel/apps/api/service"
)

type AuthHandler struct {
	svc service.AuthService
}

func NewAuthHandler(svc service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) VerifyToken(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"userId": userID})
}
