package handlers

import (
    "franchise-saas-backend/internal/models"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

// getCurrentUser - загружает пользователя из БД по ID из токена
func getCurrentUser(c *gin.Context) *models.User {
    // 1. Достаем ID из контекста (проверен middleware)
    userIDStr, exists := c.Get("userID")
    if !exists {
        return nil
    }

    uid, err := uuid.Parse(userIDStr.(string))
    if err != nil {
        return nil
    }

    // 2. Достаем DB из контекста (добавлено в main.go)
    dbVal, exists := c.Get("db")
    if !exists {
        return nil
    }

    db, ok := dbVal.(*gorm.DB)
    if !ok {
        return nil
    }

    // 3. Загружаем пользователя из БД
    // Теперь мы получаем корректный SalonID!
    var user models.User
    if err := db.First(&user, "id = ?", uid).Error; err != nil {
        return nil
    }

    return &user
}
