package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockAuthService struct{}

func (m *mockAuthService) VerifyToken(_ context.Context, token string) (string, error) {
	if token == "valid-token" {
		return "user-123", nil
	}
	return "", assertAnError{"invalid"}
}

type assertAnError struct {
	msg string
}

func (e assertAnError) Error() string { return e.msg }

func setupRouter(mw *AuthMiddleware) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", mw.RequireAuth(), func(c *gin.Context) {
		uid, _ := c.Get("userID")
		c.JSON(http.StatusOK, gin.H{"userId": uid})
	})
	return r
}

func TestRequireAuth(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "valid token",
			authHeader: "Bearer valid-token",
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing token",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid format",
			authHeader: "Invalid-format",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid token",
			authHeader: "Bearer bad-token",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := NewAuthMiddleware(&mockAuthService{})
			router := setupRouter(mw)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tt.authHeader)
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d; want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && w.Body.String() != tt.wantBody {
				t.Errorf("body = %q; want %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}
