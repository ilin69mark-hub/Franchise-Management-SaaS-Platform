package handlers

import (
	"errors"
	"strings"

	"franchise-saas-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")
var ErrInvalidSession = errors.New("invalid session")

// ErrUserBlocked — аккаунт заблокирован/приостановлен (HTTP 401/403 у хендлеров).
var ErrUserBlocked = errors.New("account is blocked")

// getCurrentUser - загружает пользователя из БД по ID из токена
func getCurrentUser(c *gin.Context) (*models.User, error) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		return nil, ErrInvalidSession
	}

	uid, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return nil, err
	}

	dbVal, exists := c.Get("db")
	if !exists {
		return nil, errors.New("db not found in context")
	}

	db, ok := dbVal.(*gorm.DB)
	if !ok {
		return nil, errors.New("invalid db type")
	}

	var user models.User
	if err := db.First(&user, "id = ?", uid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// REAUDIT-3: блокировка/удаление действует немедленно, а не "на следующем
	// логине" — раньше уже выданный access-токен продолжал работать до 24 часов.
	switch strings.ToLower(strings.TrimSpace(user.Status)) {
	case "blocked", "suspended", "banned":
		return nil, ErrUserBlocked
	}

	return &user, nil
}
