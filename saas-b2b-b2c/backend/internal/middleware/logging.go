package middleware

import (
    "franchise-saas-backend/internal/models"
    "franchise-saas-backend/internal/repository"
    "log"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// LoggingMiddleware записывает активность пользователей
func LoggingMiddleware(userRepo *repository.UserRepository) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Пропускаем лишнее
        if c.Request.Method == http.MethodOptions || c.Request.URL.Path == "/health" || strings.Contains(c.Request.URL.Path, "/static") {
            c.Next()
            return
        }

        path := c.Request.URL.Path
        
        // Записываем лог асинхронно только если пользователь авторизован — копируем значения до горутины
        userVal, exists := c.Get("currentUser")
        if exists {
            user := userVal.(*models.User)
            ip := c.ClientIP()
            ua := c.Request.UserAgent()
            // не логируем Authorization/Cookie — защита от утечки секретов в user_logs/user_agent
            if ua != "" && len(ua) > 500 {
                ua = ua[:500]
            }
            uid := user.ID
            tid := user.TenantID
            act := "api_request"
            if strings.Contains(path, "/auth/login") {
                act = "login"
            }
            // truncate UA, already done
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
            }(uid, tid, act, ip, ua)
        }

        c.Next()
    }
}
