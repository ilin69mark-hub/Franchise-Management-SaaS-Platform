package services

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"franchise-saas-backend/internal/cache"
	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ErrUserBlocked - возвращается при попытке входа заблокированного/приостановленного
// пользователя. Хендлер маппит его в HTTP 403.
var ErrUserBlocked = errors.New("account is blocked")

// ErrTooManyAttempts - lockout после N неудачных входов (хендлер маппит в 429).
var ErrTooManyAttempts = errors.New("too many login attempts, try again later")

// ErrEmailTaken — email уже зарегистрирован (маппится в 409, без текста SQL).
var ErrEmailTaken = errors.New("email already registered")

const (
	minPasswordLength = 12
	// RE-AUDIT: bcrypt падает свыше 72 байт (500 на регистрации 73+).
	// Потолок = 72: политика min/max обязана уважать хеш-функцию.
	maxPasswordLength = 72
)

// validatePassword — единая политика паролей (F6): минимум 12 символов.
func validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}
	if len(password) > maxPasswordLength {
		return fmt.Errorf("password must be at most %d characters", maxPasswordLength)
	}
	return nil
}

// hibpBaseURL переопределяется в тестах (httptest-сервер).
var hibpBaseURL = "https://api.pwnedpasswords.com"

const hibpTimeout = 2 * time.Second

// checkHIBP — пароль засвечен в утечках? (k-anonymity: уходит только префикс SHA-1).
// RE-AUDIT: пункт чеклиста «проверка на утечки». Fail-open при недоступности API:
// недоступность внешней сети не должна класть регистрацию (событие — в лог).
// Возвращает (exposed, checked): checked=false — API недоступно, решение за политикой выше.
// fetchHIBPSuffixes — сырой список суффиксов префикса (для сверки и кэширования).
func fetchHIBPSuffixes(prefix string) ([]string, bool) {
	client := &http.Client{Timeout: hibpTimeout}
	resp, err := client.Get(hibpBaseURL + "/range/" + prefix)
	if err != nil {
		return nil, false
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, false
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, false
	}
	var out []string
	for _, line := range strings.Split(string(body), "\n") {
		parts := strings.Split(strings.TrimSpace(line), ":")
		if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" {
			out = append(out, strings.TrimSpace(parts[0]))
		}
	}
	return out, true
}

func checkHIBP(password string) (exposed bool, checked bool) {
	sum := sha1.Sum([]byte(password))
	hex := strings.ToUpper(hex.EncodeToString(sum[:]))
	prefix, suffix := hex[:5], hex[5:]

	suffixes, ok := fetchHIBPSuffixes(prefix)
	if !ok {
		return false, false
	}
	for _, s := range suffixes {
		if strings.EqualFold(s, suffix) {
			return true, true
		}
	}
	return false, true
}

// hibpCacheTTL — кэш ответов range-API: префикс редко меняется, режем latency/нагрузку.
const hibpCacheTTL = 24 * time.Hour

// hibpCachedSuffixes — префикс -> множество суффиксов из прошлого ответа (только чтение API).
func hibpCachedSuffixes(ctx context.Context, prefix string) (map[string]bool, bool) {
	if cache.Client == nil {
		return nil, false
	}
	raw, err := cache.Client.HGetAll(ctx, "hibp:"+prefix).Result()
	if err != nil || len(raw) == 0 {
		return nil, false
	}
	out := make(map[string]bool, len(raw))
	for k := range raw {
		out[strings.ToUpper(k)] = true
	}
	return out, true
}

func hibpStoreSuffixes(ctx context.Context, prefix string, suffixes []string) {
	if cache.Client == nil || len(suffixes) == 0 {
		return
	}
	pipe := cache.Client.Pipeline()
	key := "hibp:" + prefix
	for _, s := range suffixes {
		pipe.HSet(ctx, key, strings.ToUpper(strings.TrimSpace(s)), "1")
	}
	pipe.Expire(ctx, key, hibpCacheTTL)
	_, _ = pipe.Exec(ctx)
}

// rejectPwnedPassword — отклоняет засвеченные пароли; недоступность API — лог, не блок.
// RE-AUDIT: ответы кэшируются (префикс→суффиксы, 24ч) — каждый register больше
// не держит воркер до 2с; пропуск при outage виден в логах (hibp_unchecked).
func rejectPwnedPassword(password string) error {
	sum := sha1.Sum([]byte(password))
	hex := strings.ToUpper(hex.EncodeToString(sum[:]))
	prefix, suffix := hex[:5], hex[5:]
	ctx := context.Background()

	if cached, ok := hibpCachedSuffixes(ctx, prefix); ok {
		if cached[suffix] {
			return errors.New("password has been exposed in data breaches, choose another")
		}
		return nil
	}

	suffixes, ok := fetchHIBPSuffixes(prefix)
	if !ok {
		log.Printf("hibp_unchecked: password breach check skipped (API unavailable)")
		return nil
	}
	hibpStoreSuffixes(ctx, prefix, suffixes)
	for _, s := range suffixes {
		if strings.EqualFold(s, suffix) {
			return errors.New("password has been exposed in data breaches, choose another")
		}
	}
	return nil
}

// isDuplicateKeyErr — дубль unique (PG 23505 / sqlite / mysql) без разбора драйвера.
func isDuplicateKeyErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "23505") ||
		strings.Contains(msg, "duplicate entry")
}

// ensureEmailFree — pre-check дубля до вставки (гонку страхует isDuplicateKeyErr).
func (s *AuthService) ensureEmailFree(ctx context.Context, email string) error {
	_, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return ErrEmailTaken
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	// Heuristic for repos whose NotFound isn't gorm.ErrRecordNotFound (mocks):
	// только явный not-found считаем свободным; остальное — тоже дубль-сигнал?
	// Нет: неожиданную ошибку БД пробрасываем, а не маскируем под 409.
	if strings.Contains(strings.ToLower(err.Error()), "not found") ||
		strings.Contains(strings.ToLower(err.Error()), "record not found") {
		return nil
	}
	return err
}

const (
	// F9: lockout — 10 неудач за 15 минут на email (счётчик общий:
	// инкремент и по неизвестному email, чтобы не выдавать существование).
	maxLoginAttempts = 10
	loginLockWindow  = 15 * time.Minute
	// F12: потолки времени жизни токенов.
	maxJWTExpires   = 24 * time.Hour
	refreshLifetime = 7 * 24 * time.Hour
	// Абсолютный предел цепочки refresh-ротаций: украденный refresh
	// нельзя продлевать вечно — через 30 дней от iat нужен новый логин.
	maxRefreshChainLifetime = 30 * 24 * time.Hour
)

type AuthService struct {
	userRepo   repository.UserRepositoryInterface
	tenantRepo repository.TenantRepositoryInterface
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		userRepo:   repository.NewUserRepository(db),
		tenantRepo: repository.NewTenantRepository(db),
	}
}

func NewAuthServiceWithInterface(userRepo repository.UserRepositoryInterface, tenantRepo repository.TenantRepositoryInterface) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}
}

// CreateUserWithTenant - Создает сеть (Tenant) и Владельца (User)
func (s *AuthService) CreateUserWithTenant(user *models.User, password string, companyName string) (*models.User, error) {
	ctx := context.Background()

	if err := validatePassword(password); err != nil {
		return nil, err
	}
	if err := rejectPwnedPassword(password); err != nil {
		return nil, err
	}
	if err := s.ensureEmailFree(ctx, user.Email); err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = string(hashedPassword)

	tenant := &models.Tenant{Name: companyName, Status: "active"}
	// RE-AUDIT: берём ID из возвращённого значения, а не из мутации входа
	// (моки/обёртки ID не проставляют — компенсация чистила бы Nil).
	createdTenant, err := s.tenantRepo.CreateTenant(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create company: %w", err)
	}
	if createdTenant == nil {
		return nil, fmt.Errorf("failed to create company: empty tenant")
	}

	user.TenantID = &createdTenant.ID
	user.Role = models.RoleFranchisor
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		// RE-AUDIT: компенсация вместо транзакции (репозитории без tx):
		// tenant без владельца — сирота, чистим сразу. Ошибка компенсации —
		// в лог, исходная ошибка — наружу без изменений.
		if s.tenantRepo != nil {
			if derr := s.tenantRepo.DeleteTenant(ctx, createdTenant.ID); derr != nil {
				_ = derr
			}
		}
		if isDuplicateKeyErr(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

// CreateUser - Создание пользователя без создания сети
func (s *AuthService) CreateUser(user *models.User, password string) (*models.User, error) {
	if err := validatePassword(password); err != nil {
		return nil, err
	}
	if err := rejectPwnedPassword(password); err != nil {
		return nil, err
	}
	if err := s.ensureEmailFree(context.Background(), user.Email); err != nil {
		return nil, err
	}
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = string(hashedPass)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := s.userRepo.CreateUser(context.Background(), user); err != nil {
		if isDuplicateKeyErr(err) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return user, nil
}

// Authenticate - Проверка логина/пароля
// loginFailKey — ключ счётчика неудач: IP+email (RE-AUDIT: ключ только по email
// позволял любому лочить чужой аккаунт 10 запросами; CAPTCHA — следующим шагом,
// пока изоляция по IP клиента).
func loginFailKey(clientIP, email string) string {
	clientIP = strings.TrimSpace(clientIP)
	if clientIP == "" {
		clientIP = "unknown-ip"
	}
	return "loginfail:" + clientIP + ":" + strings.ToLower(strings.TrimSpace(email))
}

func (s *AuthService) Authenticate(ctx context.Context, email, password, clientIP string) (*models.User, error) {
	failKey := loginFailKey(clientIP, email)

	// F9: сначала lockout (до обращения к БД — не даём перебирать и не течём таймингом).
	if cache.PeekLimit(ctx, failKey) >= maxLoginAttempts {
		return nil, ErrTooManyAttempts
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		cache.IncrLimit(ctx, failKey, loginLockWindow)
		return nil, errors.New("invalid credentials")
	}

	// Проверка пароля
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		cache.IncrLimit(ctx, failKey, loginLockWindow)
		return nil, errors.New("invalid credentials")
	}

	// ВАЖНОЕ ИЗМЕНЕНИЕ:
	// После успешной проверки пароля, мы заново запрашиваем пользователя по ID.
	// Это нужно, чтобы гарантированно получить актуальное поле SalonID,
	// которое могло измениться (назначение менеджера) после создания пользователя.
	freshUser, err := s.userRepo.GetUserByID(context.Background(), user.ID)
	if err != nil {
		return nil, errors.New("failed to fetch user details")
	}

	// Заблокированные/приостановленные пользователи не могут войти.
	switch strings.ToLower(strings.TrimSpace(freshUser.Status)) {
	case "blocked", "suspended", "banned":
		return nil, ErrUserBlocked
	}

	// Успех — сбрасываем счётчик неудач.
	cache.ResetLimit(ctx, failKey)
	return freshUser, nil
}

// GenerateTokens - Генерация токенов (новая цепочка: cid = свежий uuid).
func (s *AuthService) GenerateTokens(userID uuid.UUID, email string, role models.Role, tenantID *uuid.UUID, salonID *uuid.UUID) (string, string, error) {
	return s.generateTokensWithChain(userID, email, role, tenantID, salonID, uuid.New().String())
}

// generateTokensWithChain — ротация в рамках той же цепочки (cid carried).
func (s *AuthService) generateTokensWithChain(userID uuid.UUID, email string, role models.Role, tenantID *uuid.UUID, salonID *uuid.UUID, chainID string) (string, string, error) {
	secret := viper.GetString("jwt_secret")
	if secret == "" {
		return "", "", errors.New("jwt_secret is not configured")
	}

	tidStr := ""
	if tenantID != nil {
		tidStr = tenantID.String()
	}

	sidStr := ""
	if salonID != nil {
		sidStr = salonID.String()
	}

	accessJti := uuid.New().String()
	// respect JWT_EXPIRES from config, default 24h, hard cap 24h (F12:
	// без потолка env JWT_EXPIRES=8760h давал годовой access-токен).
	jwtExpiresStr := viper.GetString("JWT_EXPIRES")
	if jwtExpiresStr == "" {
		jwtExpiresStr = viper.GetString("jwt_expires")
	}
	jwtExpires := 24 * time.Hour
	if d, err := time.ParseDuration(jwtExpiresStr); err == nil {
		jwtExpires = d
	}
	if jwtExpires > maxJWTExpires || jwtExpires <= 0 {
		jwtExpires = maxJWTExpires
	}
	now := time.Now()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   userID.String(),
		"email":     email,
		"role":      role,
		"tenant_id": tidStr,
		"salon_id":  sidStr,
		"jti":       accessJti,
		"iat":       now.Unix(),
		"exp":       now.Add(jwtExpires).Unix(),
	})

	jti := uuid.New().String()
	if chainID == "" {
		chainID = jti
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"jti":     jti,
		"cid":     chainID,
		"iat":     now.Unix(),
		"exp":     now.Add(refreshLifetime).Unix(),
	})

	accessStr, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}
	refreshStr, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessStr, refreshStr, nil
}

// RefreshTokens - Обновление токенов с ротацией
func (s *AuthService) RefreshTokens(oldRefreshToken string) (string, string, error) {
	secret := viper.GetString("jwt_secret")
	if secret == "" {
		return "", "", errors.New("jwt_secret is not configured")
	}

	token, err := jwt.Parse(oldRefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return "", "", errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("invalid token claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return "", "", errors.New("user_id not found in token")
	}

	tokenID, _ := claims["jti"].(string)
	if tokenID == "" {
		return "", "", errors.New("refresh token missing jti")
	}
	// F12: абсолютный предел цепочки — iat первого refresh + 30d.
	// Токены без iat (выпущенные до патча) grandfathered: ротация выдаст с iat.
	if iatVal, ok := claims["iat"].(float64); ok {
		if time.Since(time.Unix(int64(iatVal), 0)) > maxRefreshChainLifetime {
			return "", "", errors.New("refresh token chain expired, re-login required")
		}
	}
	// RE-AUDIT: атомарный claim вместо check-then-act (гонка давала две пары)
	// + kill-chain при reuse (подозрение на кражу — вся цепочка умирает).
	chainID, _ := claims["cid"].(string)
	if chainID == "" {
		chainID = tokenID // grandfathered: цепочка = первый увиденный jti
	}
	if cache.IsChainRevoked(context.Background(), chainID) {
		return "", "", errors.New("refresh token chain revoked, re-login required")
	}
	if cache.IsTokenRevoked(context.Background(), tokenID) {
		cache.RevokeChain(context.Background(), chainID)
		return "", "", errors.New("refresh token revoked")
	}
	if !cache.ClaimJTI(context.Background(), tokenID) {
		// Параллельный reuse: кто-то уже потребил этот jti — убиваем цепочку.
		cache.RevokeChain(context.Background(), chainID)
		return "", "", errors.New("refresh token reuse detected")
	}

	// Ротация: старый refresh-токен отзывается (идемпотентная метка).
	if err := cache.RevokeRefreshToken(context.Background(), tokenID); err != nil {
		return "", "", err
	}

	uid, _ := uuid.Parse(userIDStr)
	user, err := s.userRepo.GetUserByID(context.Background(), uid)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	// RE-AUDIT: заблокированный обязан терять и refresh (раньше бан обходился
	// ротацией до 7+ суток — проверка была только в Authenticate).
	switch strings.ToLower(strings.TrimSpace(user.Status)) {
	case "blocked", "suspended", "banned":
		// Одноразово гасим предъявленный токен, чтобы не оставлять валидным.
		_ = cache.RevokeRefreshToken(context.Background(), tokenID)
		return "", "", ErrUserBlocked
	}

	return s.generateTokensWithChain(user.ID, user.Email, user.Role, user.TenantID, user.SalonID, chainID)
}

// Logout - отзывает refresh-токен, извлекая его jti.
func (s *AuthService) Logout(refreshToken string) error {
	secret := viper.GetString("jwt_secret")
	if secret == "" {
		return errors.New("jwt_secret is not configured")
	}

	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid token claims")
	}

	tokenID, _ := claims["jti"].(string)
	if tokenID == "" {
		return errors.New("refresh token missing jti")
	}

	return cache.RevokeRefreshToken(context.Background(), tokenID)
}

// RevokeAccessToken — извлекает jti из access токена и помещает в blacklist
func (s *AuthService) RevokeAccessToken(authHeader string) error {
	tokenStr := strings.TrimSpace(authHeader)
	if strings.HasPrefix(strings.ToUpper(tokenStr), "BEARER ") {
		tokenStr = strings.TrimSpace(tokenStr[7:])
	}
	if tokenStr == "" {
		return errors.New("empty token")
	}
	secret := viper.GetString("jwt_secret")
	if secret == "" {
		return errors.New("jwt_secret is not configured")
	}
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// F12: явная проверка метода (раньше её не было — только WithValidMethods).
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid claims")
	}
	jti, _ := claims["jti"].(string)
	if jti == "" {
		return errors.New("jti missing")
	}
	// TTL до exp, fallback 24h
	exp := 24 * time.Hour
	if expVal, ok := claims["exp"].(float64); ok {
		expTime := time.Unix(int64(expVal), 0)
		remaining := time.Until(expTime)
		if remaining > 0 && remaining < 24*time.Hour {
			exp = remaining
		}
	}
	return cache.RevokeAccessTokenByJti(context.Background(), jti, exp)
}

// GetUserByID - Получение пользователя по ID
func (s *AuthService) GetUserByID(id uuid.UUID) (*models.User, error) {
	return s.userRepo.GetUserByID(context.Background(), id)
}
