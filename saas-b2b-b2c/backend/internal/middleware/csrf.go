package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	// CSRFHeader — заголовок double-submit токена.
	CSRFHeader = "X-CSRF-Token"
	// CSRFCookie — читаемая JS cookie с тем же значением (НЕ httpOnly — иначе фронт не прочитает).
	CSRFCookie = "csrf_token"
	csrfLength = 32
)

// GenerateCSRFToken — 32 случайных байта, base64url (F7).
func GenerateCSRFToken() (string, error) {
	b := make([]byte, csrfLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// CSRF — double-submit защита для cookie-сессий (F7).
// Правила:
//   - GET/HEAD/OPTIONS — пропуск (без побочных эффектов);
//   - запрос с Authorization: Bearer — пропуск (не ambient credential,
//     браузер сам его не подставит, CSRF невозможен);
//   - POST/PUT/PATCH/DELETE с cookie-авторизацией — требуют X-CSRF-Token,
//     равный csrf_token cookie (сравнение за константное время).
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}

		if c.GetHeader("Authorization") != "" {
			c.Next()
			return
		}

		cookie, err := c.Cookie(CSRFCookie)
		if err != nil || cookie == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token missing"})
			c.Abort()
			return
		}
		header := c.GetHeader(CSRFHeader)
		if header == "" || subtle.ConstantTimeCompare([]byte(header), []byte(cookie)) != 1 {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token invalid"})
			c.Abort()
			return
		}
		c.Next()
	}
}
