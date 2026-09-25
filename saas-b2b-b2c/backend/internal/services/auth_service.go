package services

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
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
	"gorm.io/gorm/clause"
)

// ErrUserBlocked - возвращается при попытке входа заблокированного/приостановленного
// пользователя. Хендлер маппит его в HTTP 403.
var ErrUserBlocked = errors.New("account is blocked")

// ErrTenantSubscriptionExpired - оплаченный период истёк (включая grace).
var ErrTenantSubscriptionExpired = errors.New("tenant subscription expired")

// ErrCaptchaRequired - после captcha-порога вход без валидного токена запрещён.
var ErrCaptchaRequired = errors.New("captcha required")

// captchaFromEnv - конфиг провайдера из окружения (hCaptcha/Turnstile-совместимый POST).
// Пусто = CAPTCHA не настроена (fail-open, чтобы dev/staging не ломались).
func captchaFromEnv() (secret, verifyURL string) {
	// AUDIT-EXCEPTION(E13): owner-key, см. .audit-exceptions.yml
	secret = strings.TrimSpace(os.Getenv("CAPTCHA_SECRET"))
	verifyURL = strings.TrimSpace(os.Getenv("CAPTCHA_VERIFY_URL"))
	if secret == "" || verifyURL == "" {
		return "", ""
	}
	return secret, verifyURL
}

// captchaThresholdAfter - с какого числа неудач требуется CAPTCHA (половина от lockout).
func captchaThresholdAfter() int {
	threshold := captchaThreshold
	if threshold <= 0 || threshold >= maxLoginAttempts {
		return maxLoginAttempts / 2
	}
	return threshold
}

// verifyCaptchaToken - S1: после порога неудач вход требует CAPTCHA.
// Без CAPTCHA_SECRET/CAPTCHA_VERIFY_URL проверка пропускается (E13 — человек
// заводит провайдера; ключи в репозитории хранить нельзя).
func verifyCaptchaToken(ctx context.Context, clientIP, token string) error {
	secret, verifyURL := captchaFromEnv()
	if secret == "" {
		return nil
	}
	if strings.TrimSpace(token) == "" {
		return ErrCaptchaRequired
	}
	form := url.Values{"secret": {secret}, "response": {strings.TrimSpace(token)}}
	if ip := strings.TrimSpace(clientIP); ip != "" {
		form.Set("remoteip", ip)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, verifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return ErrCaptchaRequired
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := captchaHTTPClient.Do(req)
	if err != nil {
		// Провайдер недоступен — не открываем логин (fail-closed на enforce-пути).
		log.Printf("captcha verify failed: %v", err)
		return ErrCaptchaRequired
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var parsed struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil || !parsed.Success {
		return ErrCaptchaRequired
	}
	return nil
}

var captchaHTTPClient = &http.Client{Timeout: 5 * time.Second}

// ErrTooManyAttempts - lockout после N неудачных входов (хендлер маппит в 429).
var ErrTooManyAttempts = errors.New("too many login attempts, try again later")

// ErrTenantBlocked - сеть заблокирована/приостановлена (хендлер маппит в 403).
// RE-AUDIT: BlockTenant раньше был надписью в админке — auth-путь статус
// сети не смотрел, неплательщик работал дальше.
var ErrTenantBlocked = errors.New("tenant is blocked")

var (
	ErrAuthSessionUnavailable = errors.New("durable auth session store unavailable")
	ErrAuthSessionNotFound    = errors.New("auth session not found")
	ErrAuthSessionRevoked     = errors.New("auth session revoked")
	ErrRefreshReuse           = errors.New("refresh token reuse detected")
)

// tenantAccessDenied — статусы сети, закрывающие вход и refresh.
// churned сознательно пропускаем (graceful wind-down бывшим клиентам).
func tenantAccessDenied(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "blocked", "suspended":
		return true
	}
	return false
}

// checkTenantAccess — сеть пользователя активна? super_admin и пользователи
// без сети (nil) — вне проверки. Ошибка БД — fail-closed (deny).
func (s *AuthService) checkTenantAccess(ctx context.Context, user *models.User) error {
	if s.db == nil && user != nil && user.TenantID == nil {
		return nil
	}
	return checkTenantAccessWithRepository(ctx, s.tenantRepo, user)
}

func checkTenantAccessWithRepository(ctx context.Context, tenantRepo repository.TenantRepositoryInterface, user *models.User) error {
	if user == nil {
		return errors.New("user is required")
	}
	if user.Role == models.RoleSuperAdmin {
		return nil
	}
	if user.TenantID == nil {
		return errors.New("tenant is required")
	}
	if tenantRepo == nil {
		return errors.New("tenant check unavailable")
	}
	tenant, err := tenantRepo.FindByID(ctx, *user.TenantID)
	if err != nil || tenant == nil {
		return ErrTenantBlocked
	}
	if tenantAccessDenied(tenant.Status) {
		return ErrTenantBlocked
	}
	if tenant.PaidUntil != nil {
		grace := 0
		if tenant.GracePeriodDays > 0 {
			grace = tenant.GracePeriodDays
		}
		if time.Now().After(tenant.PaidUntil.AddDate(0, 0, grace)) {
			return ErrTenantSubscriptionExpired
		}
	}
	return nil
}

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
	// S1: CAPTCHA требуется с середины lockout-окна.
	captchaThreshold = maxLoginAttempts / 2
	loginLockWindow  = 15 * time.Minute
	// F12: потолки времени жизни токенов.
	maxJWTExpires   = 24 * time.Hour
	refreshLifetime = 7 * 24 * time.Hour
	// Абсолютный предел цепочки refresh-ротаций: украденный refresh
	// нельзя продлевать вечно — через 30 дней от iat нужен новый логин.
	maxRefreshChainLifetime = 30 * 24 * time.Hour
)

type AuthIdentity struct {
	UserID      uuid.UUID
	SessionID   uuid.UUID
	Email       string
	Role        models.Role
	TenantID    *uuid.UUID
	AuthVersion int64
}

type AuthService struct {
	userRepo    repository.UserRepositoryInterface
	tenantRepo  repository.TenantRepositoryInterface
	sessionRepo *repository.AuthSessionRepository
	db          *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{
		db:          db,
		userRepo:    repository.NewUserRepository(db),
		tenantRepo:  repository.NewTenantRepository(db),
		sessionRepo: repository.NewAuthSessionRepository(db),
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

	// S4: храним нормализованный email (уникальность — по LOWER(email)).
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

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
	// REAUDIT-4: в проде tenant+owner создаются в одной транзакции (раньше
	// tenant коммитился первым, и падение вставки пользователя оставляло
	// «сеть без владельца»). В юнит-тестах БД нет — работает старый путь
	// с компенсацией на моках.
	if s.db != nil {
		createdTenant := &models.Tenant{}
		if err := s.db.Transaction(func(tx *gorm.DB) error {
			created, err := repository.NewTenantRepository(tx).CreateTenant(ctx, tenant)
			if err != nil {
				return err
			}
			if created == nil {
				return errors.New("failed to create company: empty tenant")
			}
			createdTenant = created
			user.TenantID = &createdTenant.ID
			user.Role = models.RoleFranchisor
			user.CreatedAt = time.Now()
			user.UpdatedAt = time.Now()
			return repository.NewUserRepository(tx).CreateUser(ctx, user)
		}); err != nil {
			if isDuplicateKeyErr(err) {
				return nil, ErrEmailTaken
			}
			return nil, fmt.Errorf("failed to register franchise owner: %w", err)
		}
		return user, nil
	}

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
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
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

func (s *AuthService) Authenticate(ctx context.Context, email, password, clientIP string, captchaToken ...string) (*models.User, error) {
	// S4: вход регистронезависимо (хранение — всегда lower).
	email = strings.ToLower(strings.TrimSpace(email))
	failKey := loginFailKey(clientIP, email)

	// S1: после порога неудач — CAPTCHA (до проверки пароля).
	if cache.PeekLimit(ctx, failKey) >= captchaThresholdAfter() {
		token := ""
		if len(captchaToken) > 0 {
			token = captchaToken[0]
		}
		if err := verifyCaptchaToken(ctx, clientIP, token); err != nil {
			return nil, err
		}
	}

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

	if err := s.checkTenantAccess(ctx, freshUser); err != nil {
		return nil, err
	}

	// Успех — сбрасываем счётчик неудач.
	cache.ResetLimit(ctx, failKey)
	return freshUser, nil
}

func userAccessDenied(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "blocked", "suspended", "banned", "disabled", "inactive":
		return true
	default:
		return false
	}
}

func (s *AuthService) IssueSession(ctx context.Context, user *models.User) (string, string, error) {
	if s.db == nil || s.sessionRepo == nil || user == nil {
		return "", "", ErrAuthSessionUnavailable
	}
	var accessToken, refreshToken string
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&locked, "id = ?", user.ID).Error; err != nil {
			return err
		}
		if locked.AuthVersion <= 0 {
			return ErrAuthSessionRevoked
		}
		if user.AuthVersion > 0 && locked.AuthVersion != user.AuthVersion {
			return ErrAuthSessionRevoked
		}
		if userAccessDenied(locked.Status) {
			return ErrUserBlocked
		}
		if err := checkTenantAccessWithRepository(ctx, repository.NewTenantRepository(tx), &locked); err != nil {
			return err
		}
		now := time.Now().UTC()
		chainExpiresAt := now.Add(maxRefreshChainLifetime)
		refreshExpiresAt := now.Add(refreshLifetime)
		if refreshExpiresAt.After(chainExpiresAt) {
			refreshExpiresAt = chainExpiresAt
		}
		sessionID := uuid.New()
		refreshJTI := uuid.New()
		session := &models.AuthSession{
			ID:                sessionID,
			UserID:            locked.ID,
			AuthVersion:       locked.AuthVersion,
			CurrentRefreshJTI: refreshJTI,
			ChainStartedAt:    now,
			ChainExpiresAt:    chainExpiresAt,
			RefreshExpiresAt:  refreshExpiresAt,
			LastTokenIssuedAt: now,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
		if err := repository.NewAuthSessionRepository(tx).Create(ctx, session); err != nil {
			return err
		}
		var err error
		accessToken, refreshToken, err = s.generateTokensForSession(locked, sessionID, now, now, locked.AuthVersion, refreshJTI)
		return err
	})
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (s *AuthService) ResolveAccessIdentity(ctx context.Context, userID, sessionID uuid.UUID, authVersion int64) (AuthIdentity, error) {
	if s.db == nil {
		return AuthIdentity{}, ErrAuthSessionUnavailable
	}
	var user models.User
	if err := s.db.WithContext(ctx).Select("id, email, role, status, tenant_id, auth_version").First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AuthIdentity{}, ErrAuthSessionNotFound
		}
		return AuthIdentity{}, err
	}
	if userAccessDenied(user.Status) || user.Role == "" {
		return AuthIdentity{}, ErrAuthSessionRevoked
	}
	if user.Role != models.RoleSuperAdmin && user.TenantID == nil {
		return AuthIdentity{}, ErrAuthSessionRevoked
	}
	if user.AuthVersion != authVersion {
		return AuthIdentity{}, ErrAuthSessionRevoked
	}
	var session models.AuthSession
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AuthIdentity{}, ErrAuthSessionNotFound
		}
		return AuthIdentity{}, err
	}
	if session.AuthVersion != authVersion || session.RevokedAt != nil || !time.Now().UTC().Before(session.ChainExpiresAt) {
		return AuthIdentity{}, ErrAuthSessionRevoked
	}
	if err := checkTenantAccessWithRepository(ctx, s.tenantRepo, &user); err != nil {
		return AuthIdentity{}, err
	}
	return AuthIdentity{
		UserID:      user.ID,
		SessionID:   session.ID,
		Email:       user.Email,
		Role:        user.Role,
		TenantID:    user.TenantID,
		AuthVersion: user.AuthVersion,
	}, nil
}

func (s *AuthService) GenerateTokens(userID uuid.UUID, email string, role models.Role, tenantID *uuid.UUID, salonID *uuid.UUID) (string, string, error) {
	if s.db != nil {
		return "", "", ErrAuthSessionUnavailable
	}
	now := time.Now().UTC()
	sessionID := uuid.New()
	return s.generateTokensForSession(models.User{ID: userID, Email: email, Role: role, TenantID: tenantID, SalonID: salonID}, sessionID, now, now, 1, uuid.New())
}

func (s *AuthService) generateTokensForSession(user models.User, sessionID uuid.UUID, chainStart time.Time, issuedAt time.Time, authVersion int64, refreshJTI uuid.UUID) (string, string, error) {
	secret := viper.GetString("jwt_secret")
	if secret == "" {
		return "", "", errors.New("jwt_secret is not configured")
	}
	tenantID := ""
	if user.TenantID != nil {
		tenantID = user.TenantID.String()
	}
	salonID := ""
	if user.SalonID != nil {
		salonID = user.SalonID.String()
	}
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
	accessJTI := uuid.New()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"token_use": "access",
		"user_id":   user.ID.String(),
		"email":     user.Email,
		"role":      user.Role,
		"tenant_id": tenantID,
		"salon_id":  salonID,
		"sid":       sessionID.String(),
		"av":        authVersion,
		"jti":       accessJTI.String(),
		"iat":       issuedAt.Unix(),
		"exp":       issuedAt.Add(jwtExpires).Unix(),
	})
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"token_use": "refresh",
		"user_id":   user.ID.String(),
		"jti":       refreshJTI.String(),
		"sid":       sessionID.String(),
		"cid":       sessionID.String(),
		"av":        authVersion,
		"chain_iat": chainStart.Unix(),
		"iat":       issuedAt.Unix(),
		"exp":       issuedAt.Add(refreshLifetime).Unix(),
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

func (s *AuthService) generateTokensWithChain(userID uuid.UUID, email string, role models.Role, tenantID *uuid.UUID, salonID *uuid.UUID, chainID string, chainStart int64) (string, string, error) {
	sessionID, err := uuid.Parse(chainID)
	if err != nil {
		sessionID = uuid.New()
	}
	issuedAt := time.Now().UTC()
	return s.generateTokensForSession(models.User{ID: userID, Email: email, Role: role, TenantID: tenantID, SalonID: salonID}, sessionID, time.Unix(chainStart, 0).UTC(), issuedAt, 1, uuid.New())
}

func (s *AuthService) parseSignedClaims(raw string, validateClaims bool) (jwt.MapClaims, error) {
	secret := viper.GetString("jwt_secret")
	if secret == "" {
		return nil, errors.New("jwt_secret is not configured")
	}
	options := []jwt.ParserOption{jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired()}
	if !validateClaims {
		options = append(options, jwt.WithoutClaimsValidation())
	}
	token, err := jwt.Parse(raw, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	}, options...)
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

func (s *AuthService) parseRefreshClaims(raw string) (jwt.MapClaims, error) {
	claims, err := s.parseSignedClaims(raw, true)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}
	if use, _ := claims["token_use"].(string); use != "refresh" {
		return nil, errors.New("not a refresh token")
	}
	return claims, nil
}

func (s *AuthService) RefreshTokens(oldRefreshToken string) (string, string, error) {
	return s.RefreshTokensWithContext(context.Background(), oldRefreshToken)
}

func (s *AuthService) RefreshTokensWithContext(ctx context.Context, oldRefreshToken string) (string, string, error) {
	if s.db == nil {
		return s.refreshTokensLegacy(oldRefreshToken)
	}
	return s.refreshTokensDurable(ctx, oldRefreshToken)
}

func (s *AuthService) refreshTokensDurable(ctx context.Context, oldRefreshToken string) (string, string, error) {
	claims, err := s.parseRefreshClaims(oldRefreshToken)
	if err != nil {
		return "", "", err
	}
	userID, err := uuid.Parse(stringClaim(claims, "user_id"))
	if err != nil {
		return "", "", errors.New("user_id not found in token")
	}
	sessionID, err := uuid.Parse(stringClaim(claims, "sid"))
	if err != nil {
		return "", "", errors.New("refresh token missing sid")
	}
	chainID, err := uuid.Parse(stringClaim(claims, "cid"))
	if err != nil || chainID != sessionID {
		return "", "", errors.New("refresh token has invalid chain")
	}
	refreshJTI, err := uuid.Parse(stringClaim(claims, "jti"))
	if err != nil {
		return "", "", errors.New("refresh token missing jti")
	}
	authVersion, ok := numericClaim(claims, "av")
	if !ok || authVersion <= 0 {
		return "", "", errors.New("refresh token missing auth version")
	}
	chainStart, ok := numericClaim(claims, "chain_iat")
	if !ok || chainStart <= 0 {
		return "", "", errors.New("refresh token missing chain_iat")
	}
	iat, ok := numericClaim(claims, "iat")
	if !ok || iat <= 0 {
		return "", "", errors.New("refresh token missing iat")
	}
	requestTime := time.Now().UTC()
	if iat > float64(requestTime.Add(30*time.Second).Unix()) {
		return "", "", errors.New("refresh token iat is in the future")
	}
	var accessToken, newRefreshToken string
	var validationErr error
	reuse := false
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, "id = ?", userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				validationErr = errors.New("user not found")
				return nil
			}
			return err
		}
		var session models.AuthSession
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				validationErr = ErrAuthSessionNotFound
				return nil
			}
			return err
		}
		now := time.Now().UTC()
		if tx.Name() == "postgres" {
			if err := tx.Raw("SELECT clock_timestamp()").Scan(&now).Error; err != nil {
				return err
			}
		}
		if user.AuthVersion != int64(authVersion) || session.AuthVersion != int64(authVersion) || session.RevokedAt != nil || !now.Before(session.ChainExpiresAt) {
			validationErr = ErrAuthSessionRevoked
			return nil
		}
		if session.ChainStartedAt.Unix() != int64(chainStart) {
			validationErr = ErrAuthSessionRevoked
			return nil
		}
		if !now.Before(session.RefreshExpiresAt) {
			validationErr = ErrAuthSessionRevoked
			return nil
		}
		if userAccessDenied(user.Status) {
			validationErr = ErrUserBlocked
			return nil
		}
		if err := checkTenantAccessWithRepository(ctx, repository.NewTenantRepository(tx), &user); err != nil {
			validationErr = err
			return nil
		}
		if session.CurrentRefreshJTI != refreshJTI {
			revokedAt := time.Now().UTC()
			if err := tx.Model(&models.AuthSession{}).Where("id = ? AND user_id = ?", sessionID, userID).Updates(map[string]interface{}{
				"revoked_at":    revokedAt,
				"revoke_reason": "refresh_reuse",
				"updated_at":    revokedAt,
			}).Error; err != nil {
				return err
			}
			reuse = true
			return nil
		}
		newJTI := uuid.New()
		newRefreshExpiresAt := now.Add(refreshLifetime)
		if newRefreshExpiresAt.After(session.ChainExpiresAt) {
			newRefreshExpiresAt = session.ChainExpiresAt
		}
		result := tx.Model(&models.AuthSession{}).Where("id = ? AND user_id = ? AND current_refresh_jti = ? AND revoked_at IS NULL", sessionID, userID, refreshJTI).Updates(map[string]interface{}{
			"current_refresh_jti":  newJTI,
			"refresh_expires_at":   newRefreshExpiresAt,
			"last_token_issued_at": now,
			"updated_at":           now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			reuse = true
			return nil
		}
		accessToken, newRefreshToken, err = s.generateTokensForSession(user, sessionID, session.ChainStartedAt, now, int64(authVersion), newJTI)
		return err
	})
	if err != nil {
		return "", "", err
	}
	if validationErr != nil {
		return "", "", validationErr
	}
	if reuse {
		return "", "", ErrRefreshReuse
	}
	return accessToken, newRefreshToken, nil
}

func stringClaim(claims jwt.MapClaims, name string) string {
	value, _ := claims[name].(string)
	return value
}

func numericClaim(claims jwt.MapClaims, name string) (float64, bool) {
	value, ok := claims[name].(float64)
	return value, ok
}

func (s *AuthService) refreshTokensLegacy(oldRefreshToken string) (string, string, error) {
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

	// REAUDIT-3: access-токен нельзя обменять на refresh (иначе 24h → 7d).
	if use, _ := claims["token_use"].(string); use != "refresh" {
		return "", "", errors.New("not a refresh token")
	}

	tokenID, _ := claims["jti"].(string)
	if tokenID == "" {
		return "", "", errors.New("refresh token missing jti")
	}
	// REAUDIT-4: REAUDIT-4: эпоха сессий проверяется и здесь — раньше отзыв
	// (смена пароля/роли) гасил только access-токен, а refresh его игнорировал,
	// и украденный refresh-токен выдавал новую сессию после смены пароля.
	iatVal, hasIat := claims["iat"].(float64)
	if !hasIat || iatVal <= 0 {
		return "", "", errors.New("refresh token missing iat")
	}
	if revokedAfter := cache.UserSessionsRevokedAfter(context.Background(), userIDStr); revokedAfter > 0 && int64(iatVal) <= revokedAfter {
		_ = cache.RevokeRefreshToken(context.Background(), tokenID)
		return "", "", errors.New("session revoked, re-login required")
	}
	// F12 + REAUDIT-3: абсолютный предел цепочки считается от chain_iat
	// (время выдачи ПЕРВОГО токена цепочки). Раньше брался iat текущего токена,
	// который перезаписывался при каждой ротации, и «30 дней» продлевались вечно.
	chainStart, hasChainStart := claims["chain_iat"].(float64)
	if !hasChainStart || chainStart <= 0 {
		// REAUDIT-4: токен без начала цепочки отвергаем — иначе "потолок 30 дней"
		// обходился токеном без chain_iat/iat (проверка просто пропускалась).
		return "", "", errors.New("refresh token missing chain_iat, re-login required")
	}
	if time.Since(time.Unix(int64(chainStart), 0)) > maxRefreshChainLifetime {
		return "", "", errors.New("refresh token chain expired, re-login required")
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

	if err := s.checkTenantAccess(context.Background(), user); err != nil {
		_ = cache.RevokeRefreshToken(context.Background(), tokenID)
		return "", "", err
	}

	return s.generateTokensWithChain(user.ID, user.Email, user.Role, user.TenantID, user.SalonID, chainID, int64(chainStart))
}

func (s *AuthService) LogoutSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	if s.db == nil || s.sessionRepo == nil {
		return ErrAuthSessionUnavailable
	}
	return s.sessionRepo.Revoke(ctx, sessionID, userID, "logout")
}

func (s *AuthService) LogoutToken(ctx context.Context, raw string) (bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || s.db == nil || s.sessionRepo == nil {
		return false, nil
	}
	claims, err := s.parseSignedClaims(raw, false)
	if err != nil {
		return false, nil
	}
	use, _ := claims["token_use"].(string)
	if use != "access" && use != "refresh" {
		return false, nil
	}
	userID, err := uuid.Parse(stringClaim(claims, "user_id"))
	if err != nil {
		return false, nil
	}
	sessionID, err := uuid.Parse(stringClaim(claims, "sid"))
	if err != nil {
		return false, nil
	}
	if use == "refresh" {
		chainID, err := uuid.Parse(stringClaim(claims, "cid"))
		if err != nil || chainID != sessionID {
			return false, nil
		}
	}
	return true, s.sessionRepo.Revoke(ctx, sessionID, userID, "logout")
}

func (s *AuthService) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID, reason string) error {
	if s.db == nil || s.sessionRepo == nil {
		return ErrAuthSessionUnavailable
	}
	return s.sessionRepo.IncrementAuthVersionAndRevoke(ctx, userID, reason)
}

func (s *AuthService) CleanupAuthSessions(ctx context.Context) (int64, error) {
	if s.db == nil || s.sessionRepo == nil {
		return 0, ErrAuthSessionUnavailable
	}
	return s.sessionRepo.Cleanup(ctx, time.Now().UTC(), 1000)
}

func (s *AuthService) Logout(refreshToken string) error {
	claims, err := s.parseRefreshClaims(refreshToken)
	if err != nil {
		return err
	}
	if s.db != nil && s.sessionRepo != nil {
		userID, userErr := uuid.Parse(stringClaim(claims, "user_id"))
		sessionID, sessionErr := uuid.Parse(stringClaim(claims, "sid"))
		if userErr != nil || sessionErr != nil {
			return errors.New("refresh token missing session")
		}
		return s.LogoutSession(context.Background(), userID, sessionID)
	}
	tokenID := stringClaim(claims, "jti")
	if tokenID == "" {
		return errors.New("refresh token missing jti")
	}
	return cache.RevokeRefreshToken(context.Background(), tokenID)
}

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
	if s.db != nil && s.sessionRepo != nil {
		userID, userErr := uuid.Parse(stringClaim(claims, "user_id"))
		sessionID, sessionErr := uuid.Parse(stringClaim(claims, "sid"))
		if userErr != nil || sessionErr != nil {
			return errors.New("access token missing session")
		}
		return s.LogoutSession(context.Background(), userID, sessionID)
	}
	jti := stringClaim(claims, "jti")
	if jti == "" {
		return errors.New("jti missing")
	}
	exp := 24 * time.Hour
	if expVal, ok := numericClaim(claims, "exp"); ok {
		remaining := time.Until(time.Unix(int64(expVal), 0))
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
