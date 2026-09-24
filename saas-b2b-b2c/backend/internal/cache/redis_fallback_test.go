package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Fallback-путь: в тестовом окружении Redis нет (Client == nil).
func TestIncrLimit_FallbackCountsAndResets(t *testing.T) {
	ctx := context.Background()
	key := "test-limit-fallback"

	require.Equal(t, 1, IncrLimit(ctx, key, time.Minute))
	require.Equal(t, 2, IncrLimit(ctx, key, time.Minute))
	require.Equal(t, 2, PeekLimit(ctx, key))
	ResetLimit(ctx, key)
	require.Equal(t, 0, PeekLimit(ctx, key))
	require.Equal(t, 1, IncrLimit(ctx, key, time.Minute))
	ResetLimit(ctx, key)
}

func TestIncrLimit_FallbackWindowExpiry(t *testing.T) {
	ctx := context.Background()
	key := "test-limit-expiry"
	require.Equal(t, 1, IncrLimit(ctx, key, 30*time.Millisecond))
	time.Sleep(60 * time.Millisecond)
	require.Equal(t, 1, IncrLimit(ctx, key, 30*time.Millisecond), "окно истекло — счёт с 1")
	ResetLimit(ctx, key)
}
