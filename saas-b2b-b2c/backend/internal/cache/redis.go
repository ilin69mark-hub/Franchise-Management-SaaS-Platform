package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func ConnectRedis() (*redis.Client, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	password := os.Getenv("REDIS_PASSWORD")

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	Client = client
	return client, nil
}

func GetRedis() *redis.Client {
	return Client
}

func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return Client.Set(ctx, key, value, expiration).Err()
}

func Get(ctx context.Context, key string) (string, error) {
	if Client == nil {
		return "", fmt.Errorf("redis client not initialized")
	}
	return Client.Get(ctx, key).Result()
}

func Delete(ctx context.Context, key string) error {
	if Client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return Client.Del(ctx, key).Err()
}

func SetWithIncrement(ctx context.Context, key string, expiration time.Duration, limit int) (int, error) {
	if Client == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}

	pipe := Client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	count := int(incr.Val())
	return count, nil
}

// F9: счётчики лимитов с fallback (Redis-first, память — при недоступности).
// Никогда не возвращают ошибку наружу как отказ: при полном отказе хранилища
// Peek даёт 0 (не блокируем легитимных), Incr продолжает считать локально.
var (
	limMu sync.Mutex
	lim   = make(map[string]*limEntry)
)

type limEntry struct {
	count int
	reset time.Time
}

const limMaxEntries = 100000

func limGet(key string, window time.Duration) (int, bool) {
	now := time.Now()
	limMu.Lock()
	defer limMu.Unlock()
	e, ok := lim[key]
	if !ok {
		return 0, false
	}
	if !now.Before(e.reset) {
		delete(lim, key)
		return 0, false
	}
	return e.count, true
}

func limIncr(key string, window time.Duration) int {
	now := time.Now()
	limMu.Lock()
	defer limMu.Unlock()
	if len(lim) >= limMaxEntries {
		for k, e := range lim {
			if !now.Before(e.reset) {
				delete(lim, k)
			}
		}
	}
	e, ok := lim[key]
	if !ok || !now.Before(e.reset) {
		lim[key] = &limEntry{count: 1, reset: now.Add(window)}
		return 1
	}
	e.count++
	return e.count
}

func limDel(key string) {
	limMu.Lock()
	defer limMu.Unlock()
	delete(lim, key)
}

// RE-AUDIT: одноразовость refresh-jti и kill-chain.
// Check-then-act парой IsTokenRevoked+RevokeRefreshToken гонялся:
// два параллельных refresh с одним токеном давали две валидные пары.
// ClaimJTI — атомарный claim (Redis SET NX / мьютекс локально).
var (
	consumedMu sync.Mutex
	consumed   = make(map[string]time.Time)
)

const consumedTTL = 7 * 24 * time.Hour

// ClaimJTI возвращает true, если jti claimed впервые (токен свеж).
// Повторный вызов с тем же jti → false (reuse/гонка).
func ClaimJTI(ctx context.Context, tokenID string) bool {
	if tokenID == "" {
		return false
	}
	if Client != nil {
		ok, err := Client.SetNX(ctx, "consumed_token:"+tokenID, "1", consumedTTL).Result()
		if err == nil {
			return ok
		}
		log.Printf("WARNING: redis ClaimJTI failed (%v) — using instance-local claim", err)
	}
	consumedMu.Lock()
	defer consumedMu.Unlock()
	now := time.Now()
	if len(consumed) >= fbMaxEntries {
		for k, exp := range consumed {
			if !now.Before(exp) {
				delete(consumed, k)
			}
		}
	}
	if exp, ok := consumed[tokenID]; ok && now.Before(exp) {
		return false
	}
	consumed[tokenID] = now.Add(consumedTTL)
	return true
}

// RevokeChain помечает всю refresh-цепочку как убитую (подозрение на кражу).
func RevokeChain(ctx context.Context, chainID string) {
	if chainID == "" {
		return
	}
	if Client != nil {
		if err := Client.Set(ctx, "revoked_chain:"+chainID, "1", maxRefreshChainLifetimeCache).Err(); err != nil {
			log.Printf("WARNING: redis RevokeChain failed (%v)", err)
		}
	}
	fbRevoke("chain:"+chainID, maxRefreshChainLifetimeCache)
}

// IsChainRevoked — убита ли цепочка.
func IsChainRevoked(ctx context.Context, chainID string) bool {
	if chainID == "" {
		return false
	}
	if Client != nil {
		if n, err := Client.Exists(ctx, "revoked_chain:"+chainID).Result(); err == nil && n > 0 {
			return true
		}
	}
	return fbIsRevoked("chain:" + chainID)
}

const maxRefreshChainLifetimeCache = 30 * 24 * time.Hour

// IncrLimit — инкремент счётчика за окно (для rate-limit/lockout).
// Возвращает текущее значение. Redis-first; без Redis — процесс-локально
// (multi-instance требует Redis — см. предупреждение в RevokeRefreshToken).
func IncrLimit(ctx context.Context, key string, window time.Duration) int {
	if Client != nil {
		pipe := Client.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
		if _, err := pipe.Exec(ctx); err == nil {
			return int(incr.Val())
		}
		log.Printf("WARNING: redis IncrLimit failed — using instance-local counter for %q", key)
	}
	return limIncr(key, window)
}

// PeekLimit — текущее значение без инкремента (0 при недоступности).
func PeekLimit(ctx context.Context, key string) int {
	if Client != nil {
		if n, err := Client.Get(ctx, key).Int(); err == nil {
			return n
		}
	}
	if n, ok := limGet(key, 0); ok {
		return n
	}
	return 0
}

// ResetLimit — сброс счётчика (успешный логин). Best-effort.
func ResetLimit(ctx context.Context, key string) {
	if Client != nil {
		_ = Client.Del(ctx, key).Err()
	}
	limDel(key)
}

func SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if Client == nil {
		return false, fmt.Errorf("redis client not initialized")
	}
	return Client.SetNX(ctx, key, value, expiration).Result()
}

// F1: отзыв сессий должен работать и без Redis (single-instance fail-safe).
// Когда Redis недоступен, пишем в процесс-локальный fallback с TTL.
// ВНИМАНИЕ: при горизонтальном масштабировании (N реплик) fallback виден
// только своему инстансу — для multi-instance Redis обязателен
// (health уже отдаёт redis:down; см. main.go). Без fallback старый код
// молча возвращал success/false — logout был театром.
var (
	fbMu      sync.Mutex
	fbRevoked = make(map[string]time.Time)
)

const fbMaxEntries = 100000

func fbRevoke(tokenID string, ttl time.Duration) {
	if tokenID == "" {
		return
	}
	fbMu.Lock()
	defer fbMu.Unlock()
	if len(fbRevoked) >= fbMaxEntries {
		now := time.Now()
		for k, exp := range fbRevoked {
			if !now.Before(exp) {
				delete(fbRevoked, k)
			}
		}
	}
	fbRevoked[tokenID] = time.Now().Add(ttl)
}

func fbIsRevoked(tokenID string) bool {
	if tokenID == "" {
		return false
	}
	fbMu.Lock()
	defer fbMu.Unlock()
	exp, ok := fbRevoked[tokenID]
	if !ok {
		return false
	}
	if !time.Now().Before(exp) {
		delete(fbRevoked, tokenID)
		return false
	}
	return true
}

func RevokeRefreshToken(ctx context.Context, tokenID string) error {
	if tokenID == "" {
		return fmt.Errorf("empty token id")
	}
	if Client == nil {
		log.Printf("WARNING: redis unavailable — refresh revocation is instance-local (jti=%.8s)", tokenID)
		fbRevoke(tokenID, 7*24*time.Hour)
		return nil
	}
	key := "revoked_token:" + tokenID
	if err := Client.Set(ctx, key, "1", 7*24*time.Hour).Err(); err != nil {
		// Redis отвалился между проверками — дублируем локально, ошибку наружу
		fbRevoke(tokenID, 7*24*time.Hour)
		return err
	}
	return nil
}

// RevokeAccessTokenByJti — отзыв access-токена с TTL до его exp (fallback-aware).
func RevokeAccessTokenByJti(ctx context.Context, tokenID string, ttl time.Duration) error {
	if tokenID == "" {
		return fmt.Errorf("empty token id")
	}
	if ttl <= 0 || ttl > 24*time.Hour {
		ttl = 24 * time.Hour
	}
	if Client == nil {
		log.Printf("WARNING: redis unavailable — access revocation is instance-local (jti=%.8s)", tokenID)
		fbRevoke(tokenID, ttl)
		return nil
	}
	key := "revoked_token:" + tokenID
	if err := Client.Set(ctx, key, "1", ttl).Err(); err != nil {
		fbRevoke(tokenID, ttl)
		return err
	}
	return nil
}

func IsTokenRevoked(ctx context.Context, tokenID string) bool {
	if tokenID == "" {
		return false
	}
	if Client != nil {
		key := "revoked_token:" + tokenID
		if exists, err := Client.Exists(ctx, key).Result(); err == nil {
			if exists > 0 {
				return true
			}
		} else {
			log.Printf("WARNING: redis Exists failed (%v) — checking local revocation fallback", err)
		}
	}
	return fbIsRevoked(tokenID)
}

func CacheUserSession(ctx context.Context, userID string, data interface{}, expiration time.Duration) error {
	if Client == nil {
		return nil
	}
	key := "session:" + userID
	return Client.Set(ctx, key, data, expiration).Err()
}

func GetUserSession(ctx context.Context, userID string) (string, error) {
	if Client == nil {
		return "", fmt.Errorf("redis client not initialized")
	}
	key := "session:" + userID
	return Client.Get(ctx, key).Result()
}

func InvalidateUserSession(ctx context.Context, userID string) error {
	if Client == nil {
		return nil
	}
	key := "session:" + userID
	return Client.Del(ctx, key).Err()
}

func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}
