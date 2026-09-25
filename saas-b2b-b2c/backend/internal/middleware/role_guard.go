package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole returns a middleware that allows the request only when the
// authenticated user's role (set as "role" in the context by AuthMiddleware)
// is present in the provided allow-list. Otherwise it aborts with 403.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role := c.GetString("role")
		if role == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		if _, ok := allowed[role]; !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		// REAUDIT-3: отсутствие tenant = отказ, а не "пропустить проверку".
		// Иначе аккаунт с tenant_id=NULL обходил tenant-гейты (салоны, чеклисты,
		// bulk-операции) и получал доступ к данным других сетей.
		if role != "super_admin" && c.GetString("tenantID") == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "tenant required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
