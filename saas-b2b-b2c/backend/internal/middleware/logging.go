package middleware

import (
	"log"
	"net/http"
	"strings"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// sanitizeUA режет CRLF (log injection, CWE-117) и длину.
func sanitizeUA(ua string) string {
	ua = strings.ReplaceAll(ua, "\n", " ")
	ua = strings.ReplaceAll(ua, "\r", " ")
	if len(ua) > 500 {
		ua = ua[:500]
	}
	return ua
}

// LoggingMiddleware записывает активность пользователей.
// F13: раньше читал c.Get("currentUser") — такой ключ никто не ставит
// (AuthMiddleware кладёт userID/role/tenantID), поэтому user_logs были
// ПУСТЫ, а инциденты невосстановимы. Теперь — userID/tenantID из контекста,
// без похода в БД (значения уже проверены AuthMiddleware выше по цепочке).
func LoggingMiddleware(userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Пропускаем лишнее
		if c.Request.Method == http.MethodOptions || c.Request.URL.Path == "/health" || strings.Contains(c.Request.URL.Path, "/static") {
			c.Next()
			return
		}

		path := c.Request.URL.Path

		// Записываем лог асинхронно только если пользователь авторизован — копируем значения до горутины
		uidStr, exists := c.Get("userID")
		if exists {
			uidParsed, err := uuid.Parse(uidStr.(string))
			if err == nil {
				var tid *uuid.UUID
				if tidStr, ok := c.Get("tenantID"); ok {
					if s, ok := tidStr.(string); ok && s != "" {
						if parsed, err := uuid.Parse(s); err == nil {
							tid = &parsed
						}
					}
				}
				ip := c.ClientIP()
				// не логируем Authorization/Cookie — защита от утечки секретов в user_logs/user_agent
				ua := sanitizeUA(c.Request.UserAgent())
				act := "api_request"
				if strings.Contains(path, "/auth/login") {
					act = "login"
				}
				go func(userID uuid.UUID, tenantID *uuid.UUID, action, ipAddr, userAgent string) {
					logEntry := models.UserLog{
						UserID:    &userID,
						TenantID:  tenantID,
						Action:    action,
						IPAddress: ipAddr,
						UserAgent: userAgent,
					}
					if err := userRepo.CreateLog(&logEntry); err != nil {
						log.Printf("LoggingMiddleware error: %v", err)
					}
				}(uidParsed, tid, act, ip, ua)
			}
		}

		c.Next()
	}
}
