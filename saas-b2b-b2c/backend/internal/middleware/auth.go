package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"franchise-saas-backend/internal/cache"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// IdentityResolver — актуальная роль/tenant пользователя из БД (REAUDIT-4).
// Возвращает ok=false, если аккаунт удалён или заблокирован.
type IdentityResolver func(ctx context.Context, userID, sessionID string, authVersion int64) (role string, tenantID string, ok bool)

var identityResolver IdentityResolver

// SetIdentityResolver — внедряется в main (nil отключает проверку, только для тестов).
func SetIdentityResolver(fn IdentityResolver) { identityResolver = fn }

// currentIdentity — актуальные роль/tenant из БД. Если резолвер не внедрён
// (юнит-тесты), возвращается ok=false.
func currentIdentity(ctx context.Context, userID, sessionID string, authVersion int64) (role string, tenantID string, ok bool) {
	if identityResolver == nil {
		return "", "", false
	}
	return identityResolver(ctx, userID, sessionID, authVersion)
}

// AuthMiddleware - основная проверка токена (поддерживает httpOnly cookie + Authorization)
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			if len(authHeader) >= 7 && strings.HasPrefix(strings.ToUpper(authHeader), "BEARER ") {
				tokenString = strings.TrimSpace(authHeader[7:])
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
				c.Abort()
				return
			}
		} else {
			// S9: только access_token — мёртвая __Host-ветка удалена
			// (__Host-cookie бэкенд никогда не ставит; приём несуществующего
			// имени создавал ложное чувство __Host-строгости).
			if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
				tokenString = cookie
			}
		}
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		secret := viper.GetString("jwt_secret")
		if secret == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if use, _ := claims["token_use"].(string); use != "access" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}
			jti, _ := claims["jti"].(string)
			if _, err := uuid.Parse(jti); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing jti"})
				c.Abort()
				return
			}
			userID, _ := claims["user_id"].(string)
			if _, err := uuid.Parse(userID); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing user_id"})
				c.Abort()
				return
			}
			iatClaim, hasIat := claims["iat"].(float64)
			if !hasIat || iatClaim <= 0 {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing iat"})
				c.Abort()
				return
			}
			sessionID, _ := claims["sid"].(string)
			if _, err := uuid.Parse(sessionID); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing session"})
				c.Abort()
				return
			}
			authVersion, hasAuthVersion := claims["av"].(float64)
			if !hasAuthVersion || authVersion <= 0 {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing auth version"})
				c.Abort()
				return
			}

			curRole, curTenantID, ok := currentIdentity(c.Request.Context(), userID, sessionID, int64(authVersion))
			if !ok || curRole == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Account or session not found"})
				c.Abort()
				return
			}
			if curRole != "super_admin" && curTenantID == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant is required"})
				c.Abort()
				return
			}
			revocationCtx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
			revoked := cache.IsTokenRevoked(revocationCtx, jti)
			cancel()
			if revoked {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token revoked"})
				c.Abort()
				return
			}
			userEmail, _ := claims["email"].(string)
			c.Set("userID", userID)
			c.Set("email", userEmail)
			c.Set("role", curRole)
			c.Set("tenantID", curTenantID)
			c.Set("sessionID", sessionID)
			c.Set("jti", jti)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
		c.Next()
	}
}
