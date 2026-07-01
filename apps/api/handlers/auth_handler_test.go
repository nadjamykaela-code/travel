package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthHandler_VerifyToken(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		wantStatus int
	}{
		{
			name:       "valid user in context",
			userID:     "user-123",
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing user in context",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tt.userID != "" {
				c.Set("userID", tt.userID)
			}

			h := &AuthHandler{}
			h.VerifyToken(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
