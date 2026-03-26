package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zxh326/kite/pkg/common"
)

func TestSetCookieSecure(t *testing.T) {
	originalHost := common.Host
	t.Cleanup(func() {
		common.Host = originalHost
	})

	t.Run("secure when host is https", func(t *testing.T) {
		common.Host = "https://kite.example.com"
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "http://kite.local", nil)

		setCookieSecure(c, "auth_token", "value", 600)

		cookie := rec.Result().Cookies()[0]
		if !cookie.Secure {
			t.Fatalf("cookie.Secure = false, want true")
		}
		if !cookie.HttpOnly {
			t.Fatalf("cookie.HttpOnly = false, want true")
		}
		if cookie.SameSite != http.SameSiteLaxMode {
			t.Fatalf("cookie.SameSite = %v, want %v", cookie.SameSite, http.SameSiteLaxMode)
		}
		if got := rec.Header().Get("Set-Cookie"); !strings.Contains(got, "SameSite=Lax") {
			t.Fatalf("Set-Cookie header = %q, want SameSite=Lax", got)
		}
	})

	t.Run("secure when forwarded proto is https", func(t *testing.T) {
		common.Host = ""
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		req := httptest.NewRequest(http.MethodGet, "http://kite.local", nil)
		req.Header.Set("X-Forwarded-Proto", "https")
		c.Request = req

		setCookieSecure(c, "auth_token", "value", 600)

		if !rec.Result().Cookies()[0].Secure {
			t.Fatalf("cookie.Secure = false, want true")
		}
	})

	t.Run("not secure on plain http", func(t *testing.T) {
		common.Host = ""
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodGet, "http://kite.local", nil)

		setCookieSecure(c, "auth_token", "value", 600)

		if rec.Result().Cookies()[0].Secure {
			t.Fatalf("cookie.Secure = true, want false")
		}
	})
}
