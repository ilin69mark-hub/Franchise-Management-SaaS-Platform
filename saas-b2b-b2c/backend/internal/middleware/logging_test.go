package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// F13: логгер обязан писать user_logs по userID из контекста (раньше читал
// несуществующий ключ currentUser и молчал).
func TestLoggingMiddleware_WritesUserLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	// :memory: + пул = каждая новая коннекция видит пустую БД.
	// Без этого тест флапает (поймано CI, локально везло).
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	// PG-теги модели (uuid/gen_random_uuid) sqlite не понимает — DDL вручную.
	require.NoError(t, db.Exec(`CREATE TABLE user_logs (
		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		user_id TEXT, tenant_id TEXT, action TEXT,
		ip_address TEXT, user_agent TEXT, created_at DATETIME)`).Error)
	userRepo := repository.NewUserRepository(db)

	uid := uuid.New()
	tid := uuid.New()
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", uid.String())
		c.Set("tenantID", tid.String())
		c.Next()
	})
	r.Use(LoggingMiddleware(userRepo))
	r.GET("/api/v1/goals", func(c *gin.Context) { c.Status(200) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/goals", nil)
	req.Header.Set("User-Agent", "test-agent")
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var logs []models.UserLog
	require.Eventually(t, func() bool {
		var n int64
		db.Model(&models.UserLog{}).Count(&n)
		if n == 0 {
			return false
		}
		db.Find(&logs)
		return len(logs) > 0
	}, 3*time.Second, 20*time.Millisecond, "user_logs должен пополниться")

	require.Equal(t, uid, *logs[0].UserID)
	require.NotNil(t, logs[0].TenantID)
	require.Equal(t, tid, *logs[0].TenantID)
	require.Equal(t, "test-agent", logs[0].UserAgent)
}

func TestLoggingMiddleware_SanitizesCRLF(t *testing.T) {
	require.Equal(t, "a b c", sanitizeUA("a\nb\rc"))
	require.LessOrEqual(t, len(sanitizeUA(string(make([]byte, 600)))), 500)
}
