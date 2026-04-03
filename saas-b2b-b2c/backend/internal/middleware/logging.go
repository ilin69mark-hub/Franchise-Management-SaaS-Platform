package middleware

import (
    "franchise-saas-backend/internal/models"
    "franchise-saas-backend/internal/repository"
    "log"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
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
        
        // Записываем лог асинхронно только если пользователь авторизован
        userVal, exists := c.Get("currentUser")
        if exists {
            user := userVal.(*models.User)
            
            go func() {
                action := "api_request"
                if strings.Contains(path, "/auth/login") {
                    action = "login"
                }

                logEntry := models.UserLog{
                    UserID:    &user.ID,
                    TenantID:  user.TenantID,
                    Action:    action,
                    IPAddress: c.ClientIP(),
                    UserAgent: c.Request.UserAgent(),
                }
                
                if err := userRepo.CreateLog(&logEntry); err != nil {
                    log.Printf("LoggingMiddleware error: %v", err)
                }
            }()
        }

        c.Next()
    }
}
