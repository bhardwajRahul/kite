package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// DevCORS returns a CORS middleware that only allows explicitly configured
// origins. In production the SPA is served from the same origin as the API
// (go:embed), so no CORS headers are needed at all.
//
// Developers running the Vite dev server on a different port can set
// CORS_ALLOWED_ORIGINS=http://localhost:5173 to enable cross-origin
// requests during local development.
//
// When allowedOrigins is empty the middleware is a no-op — no CORS
// headers are emitted and browsers enforce same-origin policy.
func DevCORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		o = strings.TrimSpace(o)
		if o != "" {
			allowed[strings.TrimRight(o, "/")] = true
		}
	}

	return func(c *gin.Context) {
		// Fast path: no origins configured → no CORS headers (production)
		if len(allowed) == 0 {
			c.Next()
			return
		}

		origin := c.Request.Header.Get("Origin")
		if origin == "" || !allowed[origin] {
			c.Next()
			return
		}

		// Only set CORS headers for explicitly allowed origins
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers",
			"Content-Type, Authorization, X-Cluster-Name")
		c.Writer.Header().Set("Access-Control-Allow-Methods",
			"GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		// Vary so caches/CDNs don't serve a response with the wrong origin
		c.Writer.Header().Add("Vary", "Origin")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
