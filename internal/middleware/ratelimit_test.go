package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(3, time.Hour)
	for i := 0; i < 3; i++ {
		if !rl.Allow("1.1.1.1") {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	if rl.Allow("1.1.1.1") {
		t.Error("4th request should be blocked")
	}
	if !rl.Allow("2.2.2.2") {
		t.Error("another IP must have its own bucket")
	}
}

func TestRateLimiterWindowReset(t *testing.T) {
	rl := NewRateLimiter(1, 10*time.Millisecond)
	rl.Allow("ip")
	if rl.Allow("ip") {
		t.Fatal("second request inside window should be blocked")
	}
	time.Sleep(15 * time.Millisecond)
	if !rl.Allow("ip") {
		t.Error("request after the window should be allowed again")
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/login", NewRateLimiter(2, time.Hour).Middleware(), func(c *gin.Context) { c.Status(http.StatusOK) })

	codes := []int{}
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "9.9.9.9:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		codes = append(codes, w.Code)
	}
	if codes[0] != http.StatusOK || codes[1] != http.StatusOK || codes[2] != http.StatusTooManyRequests {
		t.Errorf("unexpected status sequence: %v", codes)
	}
}
