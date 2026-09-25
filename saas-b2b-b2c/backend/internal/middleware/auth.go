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
	"github.com/spf13/viper"
)

// IdentityResolver — актуальная роль/tenant пользователя из БД (REAUDIT-4).
// Возвращает ok=false, если аккаунт удалён или заблокирован.
type IdentityResolver func(ctx context.Context, userID string) (role string, tenantID string, ok bool)

var identityResolver IdentityResolver

// SetIdentityResolver — внедряется в main (nil отключает проверку, только для тестов).
func SetIdentityResolver(fn IdentityResolver) { identityResolver = fn }

// currentIdentity — актуальные роль/tenant из БД. Если резолвер не внедрён
// (юнит-тесты), возвращается ok=true с пустыми значениями — вызывающий код
// тогда оставляет claims как есть.
func currentIdentity(ctx context.Context, userID string) (role string, tenantID string, ok bool) {
	if identityResolver != nil {
		return identityResolver(ctx, userID)
	}
	return "", "", true
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
		}, jwt.WithValidMethods([]string{"HS256"}))

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// REAUDIT-3: refresh-токен не должен работать как bearer.
			if use, _ := claims["token_use"].(string); use != "access" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}
			// F12: jti обязателен — токены без jti нельзя отозвать (бессмертные).
			jti, _ := claims["jti"].(string)
			if jti == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing jti"})
				c.Abort()
				return
			}
			// revocation check via jti (Redis + instance-local fallback)
			ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
			revoked := cache.IsTokenRevoked(ctx, jti)
			cancel()
			if revoked {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token revoked"})
				c.Abort()
				return
			}
			userID, ok := claims["user_id"].(string)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing user_id"})
				c.Abort()
				return
			}

			// REAUDIT-4: обязательный iat (иначе токен нельзя привязать к эпохе
			// отзыва, а проверка 0 <= revokedAfter трактовала его как мёртвый/живой
			// в зависимости от наличия маркера).
			iatClaim, hasIat := claims["iat"].(float64)
			if !hasIat || iatClaim <= 0 {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing iat"})
				c.Abort()
				return
			}

			// REAUDIT-3: после смены пароля/роли/блокировки все ранее выданные
			// токены этого пользователя недействительны (эпоха в Redis).
			ctxR, cancelR := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
			revokedAfter := cache.UserSessionsRevokedAfter(ctxR, userID)
			cancelR()
			if revokedAfter > 0 && int64(iatClaim) <= revokedAfter {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Session revoked"})
				c.Abort()
				return
			}

			userEmail, _ := claims["email"].(string)

			// REAUDIT-4: роль и tenant берём из БД, а не из токена. Иначе понижение
			// роли/перенос в другой tenant действовали только с момента refresh.
			curRole, curTenantID, ok := currentIdentity(c.Request.Context(), userID)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Account not found or disabled"})
				c.Abort()
				return
			}
			if curRole == "" {
				curRole, _ = claims["role"].(string)
			}
			if curTenantID == "" {
				curTenantID, _ = claims["tenant_id"].(string)
			}

			c.Set("userID", userID)
			c.Set("email", userEmail)
			c.Set("role", curRole)
			c.Set("tenantID", curTenantID)
			c.Set("jti", jti)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
		c.Next()
	}
}
