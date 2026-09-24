package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"franchise-saas-backend/internal/middleware"
	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"
	"franchise-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// F3/F7 e2e: настоящая цепочка sqlite → repo → service → handler +
// настоящие RequireRole/CSRF (раньше matrix тестировалась на моках,
// открытые роуты main.go оставались зелёными).
func setupMatrixDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, db.Exec(`CREATE TABLE users (
		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		email TEXT, password_hash TEXT, role TEXT, status TEXT,
		tenant_id TEXT, salon_id TEXT, managed_by TEXT,
		first_name TEXT, last_name TEXT, phone TEXT,
		display_name TEXT, position TEXT, bio TEXT, quote TEXT, avatar_url TEXT,
		user_status TEXT, available_for_questions INTEGER, achievements TEXT,
		contacts_email_visible INTEGER, contacts_phone_visible INTEGER,
		contacts_phone TEXT, contacts_telegram TEXT, contacts_whats_app TEXT,
		contacts_working_hours TEXT,
		deleted_at DATETIME, created_at DATETIME, updated_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`CREATE TABLE plans (
		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		name TEXT, price NUMERIC, max_salons INTEGER, max_users INTEGER,
		created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`).Error)
	return db
}

func matrixRouter(db *gorm.DB, role, userID, tenantID string, userSvc *services.UserService, planSvc services.PlanService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api/v1/")
	g.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Set("userID", userID)
		c.Set("role", role)
		c.Set("tenantID", tenantID)
		c.Next()
	})
	g.Use(middleware.CSRF())
	uh := NewUserHandler(userSvc)
	ug := g.Group("/users")
	ug.Use(middleware.RequireRole("franchiser", "franchiser_manager", "dealer", "super_admin"))
	ug.POST("", uh.CreateEmployee)
	NewPlanHandler(g, planSvc)
	return r
}

func doMatrixReq(t *testing.T, r *gin.Engine, method, path string, body any, headers map[string]string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	req, err := http.NewRequest(method, path, buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, ck := range cookies {
		req.AddCookie(ck)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func csrfPair() []*http.Cookie {
	return []*http.Cookie{{Name: middleware.CSRFCookie, Value: "test-csrf-token"}}
}

func csrfHeaders() map[string]string {
	return map[string]string{middleware.CSRFHeader: "test-csrf-token"}
}

func TestRouteMatrix_PlansDealerForbidden(t *testing.T) {
	db := setupMatrixDB(t)
	planSvc := services.NewPlanService(repository.NewPlanRepository(db))
	userSvc := services.NewUserService(db)
	r := matrixRouter(db, "dealer", uuid.New().String(), uuid.New().String(), userSvc, planSvc)

	w := doMatrixReq(t, r, "POST", "/api/v1/plans",
		map[string]any{"name": "P", "price": 100.0, "max_salons": 1, "max_users": 5},
		csrfHeaders(), csrfPair())
	require.Equal(t, http.StatusForbidden, w.Code, "F3: dealer не создаёт тарифы")
}

func TestRouteMatrix_PlansSuperAdminAllowed(t *testing.T) {
	db := setupMatrixDB(t)
	planSvc := services.NewPlanService(repository.NewPlanRepository(db))
	userSvc := services.NewUserService(db)
	r := matrixRouter(db, "super_admin", uuid.New().String(), "", userSvc, planSvc)

	w := doMatrixReq(t, r, "POST", "/api/v1/plans",
		map[string]any{"name": "P", "price": 100.0, "max_salons": 1, "max_users": 5},
		csrfHeaders(), csrfPair())
	require.Equal(t, http.StatusCreated, w.Code)
}

func TestRouteMatrix_CSRFBlocksCookieMutations(t *testing.T) {
	db := setupMatrixDB(t)
	planSvc := services.NewPlanService(repository.NewPlanRepository(db))
	userSvc := services.NewUserService(db)
	r := matrixRouter(db, "super_admin", uuid.New().String(), "", userSvc, planSvc)

	body := map[string]any{"name": "P", "price": 100.0, "max_salons": 1, "max_users": 5}
	w := doMatrixReq(t, r, "POST", "/api/v1/plans", body, nil, nil)
	require.Equal(t, http.StatusForbidden, w.Code, "F7: мутация с cookie без CSRF — 403")

	w = doMatrixReq(t, r, "POST", "/api/v1/plans", body,
		map[string]string{"Authorization": "Bearer api-token"}, nil)
	require.Equal(t, http.StatusCreated, w.Code, "F7: Bearer exempt — CSRF не требуется")
}

func TestRouteMatrix_UsersSalonManagerForbidden(t *testing.T) {
	db := setupMatrixDB(t)
	planSvc := services.NewPlanService(repository.NewPlanRepository(db))
	userSvc := services.NewUserService(db)
	r := matrixRouter(db, "salon_manager", uuid.New().String(), uuid.New().String(), userSvc, planSvc)

	w := doMatrixReq(t, r, "POST", "/api/v1/users",
		map[string]any{"email": "x@t.com", "password": "long-password-12", "first_name": "N", "role": "salon_manager"},
		csrfHeaders(), csrfPair())
	require.Equal(t, http.StatusForbidden, w.Code, "F3: salon_manager не создаёт сотрудников")
}

func TestRouteMatrix_UsersDealerCreatesInOwnTenant(t *testing.T) {
	db := setupMatrixDB(t)
	tenantA := uuid.New()
	dealerID := uuid.New()
	require.NoError(t, db.Exec(
		`INSERT INTO users (id, email, role, tenant_id, first_name) VALUES (?, ?, ?, ?, ?)`,
		dealerID.String(), "dealer@t.com", string(models.RoleDealer), tenantA.String(), "D").Error)

	planSvc := services.NewPlanService(repository.NewPlanRepository(db))
	userSvc := services.NewUserService(db)
	r := matrixRouter(db, "dealer", dealerID.String(), tenantA.String(), userSvc, planSvc)

	w := doMatrixReq(t, r, "POST", "/api/v1/users",
		map[string]any{"email": "new@t.com", "password": "long-password-12", "first_name": "N", "role": "salon_manager"},
		csrfHeaders(), csrfPair())
	require.Equal(t, http.StatusCreated, w.Code)
}
