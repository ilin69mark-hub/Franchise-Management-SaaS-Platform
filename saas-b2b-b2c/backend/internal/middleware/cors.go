package middleware

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

var corsDenyWarnOnce sync.Once

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
		var allowedOrigins []string
		if raw == "" {
			if gin.Mode() == gin.ReleaseMode {
				// F11: fail-closed в проде — никакого Allow-Origin по умолчанию.
				// Раньше тихо разрешался localhost (dev-доступ в прод-API).
				corsDenyWarnOnce.Do(func() {
					log.Printf("WARNING: CORS_ALLOWED_ORIGINS empty in release mode — denying cross-origin requests")
				})
				if c.Request.Method == "OPTIONS" {
					c.AbortWithStatus(204)
					return
				}
				c.Next()
				return
			}
			allowedOrigins = []string{"http://localhost:3000", "http://localhost:8080"}
		} else {
			for _, o := range strings.Split(raw, ",") {
				if o = strings.TrimSpace(o); o != "" {
					allowedOrigins = append(allowedOrigins, o)
				}
			}
		}
		origin := c.Request.Header.Get("Origin")

		matched := false
		if origin != "" {
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
					c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
					c.Writer.Header().Set("Vary", "Origin")
					matched = true
					break
				}
			}
		}

		// F11: служебные заголовки — только доверенному origin.
		// Раньше отдавались всем (разведка preflight для чужих сайтов).
		if matched {
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
			c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
