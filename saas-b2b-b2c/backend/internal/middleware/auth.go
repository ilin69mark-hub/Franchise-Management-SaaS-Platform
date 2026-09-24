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
			// fallback на httpOnly cookie (миграция с localStorage)
			if cookie, err := c.Cookie("access_token"); err == nil && cookie != "" {
				tokenString = cookie
			} else if cookie, err := c.Cookie("__Host-access_token"); err == nil && cookie != "" {
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

			userEmail, _ := claims["email"].(string)
			userRole, _ := claims["role"].(string)

			// tenant_id может отсутствовать у super_admin
			tenantID, _ := claims["tenant_id"].(string)

			c.Set("userID", userID)
			c.Set("email", userEmail)
			c.Set("role", userRole)
			c.Set("tenantID", tenantID)
			c.Set("jti", jti)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
		c.Next()
	}
}
