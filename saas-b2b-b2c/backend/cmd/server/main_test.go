package main

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// RE-AUDIT: сиды демо-аккаунтов запрещены в проде.
func TestIsProdEnv(t *testing.T) {
	prevMode := gin.Mode()
	prevAppEnv := ""
	t.Cleanup(func() {
		gin.SetMode(prevMode)
		t.Setenv("APP_ENV", prevAppEnv)
	})

	gin.SetMode(gin.TestMode)
	t.Setenv("APP_ENV", "")
	require.False(t, isProdEnv(), "dev по умолчанию — не прод")

	t.Setenv("APP_ENV", "production")
	require.True(t, isProdEnv(), "APP_ENV=production — прод")

	t.Setenv("APP_ENV", "")
	gin.SetMode(gin.ReleaseMode)
	require.True(t, isProdEnv(), "release — прод")
}
