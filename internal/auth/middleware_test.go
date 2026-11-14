package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", AuthMiddleware(), func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/admin", AuthMiddleware(), RequireAdmin(), func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func do(r *gin.Engine, path, authHeader string) int {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestAuthMiddleware(t *testing.T) {
	r := newRouter()
	userToken, _ := GenerateToken(1, "user")
	adminToken, _ := GenerateToken(2, "admin")

	cases := []struct {
		name, path, header string
		want               int
	}{
		{"missing header", "/protected", "", http.StatusUnauthorized},
		{"malformed header", "/protected", "Token abc", http.StatusUnauthorized},
		{"garbage token", "/protected", "Bearer not-a-jwt", http.StatusUnauthorized},
		{"valid user", "/protected", "Bearer " + userToken, http.StatusOK},
		{"user on admin route", "/admin", "Bearer " + userToken, http.StatusForbidden},
		{"admin on admin route", "/admin", "Bearer " + adminToken, http.StatusOK},
	}
	for _, tc := range cases {
		if got := do(r, tc.path, tc.header); got != tc.want {
			t.Errorf("%s: got %d, want %d", tc.name, got, tc.want)
		}
	}
}
