package middleware

import (
    "log"
    "net/http"

    "franchise-saas-backend/internal/models"

    "github.com/gin-gonic/gin"
)

func SuperAdminMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        roleVal, exists := c.Get("role")

        // Логирование для отладки
        log.Printf("SuperAdmin Check | Role: %v | Exists: %v", roleVal, exists)

        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Role not found in context"})
            c.Abort()
            return
        }

        role, ok := roleVal.(string)
        if !ok {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid role type"})
            c.Abort()
            return
        }

        if role != "super_admin" && role != string(models.RoleSuperAdmin) {
            log.Printf("Access Denied for role: %s", role)
            c.JSON(http.StatusForbidden, gin.H{
                "error":   "Access denied",
                "message": "You do not have super admin privileges",
            })
            c.Abort()
            return
        }

        log.Printf("Access Granted to Super Admin")
        c.Next()
    }
}
