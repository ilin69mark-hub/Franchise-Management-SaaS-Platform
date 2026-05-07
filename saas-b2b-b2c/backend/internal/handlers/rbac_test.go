package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"franchise-saas-backend/internal/models"
)

type RBACTestHandler struct {
	dealers map[uuid.UUID]models.Dealer
}

func NewRBACTestHandler() *RBACTestHandler {
	return &RBACTestHandler{
		dealers: make(map[uuid.UUID]models.Dealer),
	}
}

func (h *RBACTestHandler) GetDealers(c *gin.Context) {
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "role not found"})
		return
	}

	switch role.(models.Role) {
	case models.RoleFranchisor, models.RoleFranchisorManager:
		c.JSON(http.StatusOK, gin.H{"dealers": []models.Dealer{
			{ID: uuid.New(), Name: "Dealer 1"},
			{ID: uuid.New(), Name: "Dealer 2"},
		}})
	case models.RoleDealer:
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied", "message": "dealers list is for franchiser only"})
	case models.RoleDealerManager:
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied", "message": "insufficient permissions"})
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
	}
}

func (h *RBACTestHandler) GetDealerByID(c *gin.Context) {
	dealerID := c.Param("id")
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")

	roleStr := role.(models.Role)
	dealerUUID, err := uuid.Parse(dealerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dealer id"})
		return
	}

	dealer, exists := h.dealers[dealerUUID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "dealer not found"})
		return
	}

	switch roleStr {
	case models.RoleFranchisor:
		c.JSON(http.StatusOK, gin.H{"dealer": dealer})
	case models.RoleDealer:
		userIDStr := userID.(string)
		if dealer.UserID.String() != userIDStr {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this dealer"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"dealer": dealer})
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}

func setupRBACRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewRBACTestHandler()

	authMiddleware := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			roleHeader := c.GetHeader("X-Test-Role")
			userIDHeader := c.GetHeader("X-Test-User-ID")

			if roleHeader == "" {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			c.Set("role", models.Role(roleHeader))
			c.Set("userID", userIDHeader)
			c.Next()
		}
	}

	r.GET("/api/v1/dealers", authMiddleware(), handler.GetDealers)
	r.GET("/api/v1/dealers/:id", authMiddleware(), handler.GetDealerByID)

	return r
}

func TestRBAC_DealerList_AccessByRole(t *testing.T) {
	router := setupRBACRouter()

	tests := []struct {
		name      string
		role      string
		userID    string
		wantCode  int
		wantError string
	}{
		{
			name:      "franchiser can access dealers list",
			role:      "franchiser",
			userID:    uuid.New().String(),
			wantCode:  http.StatusOK,
			wantError: "",
		},
		{
			name:      "franchiser_manager can access dealers list",
			role:      "franchiser_manager",
			userID:    uuid.New().String(),
			wantCode:  http.StatusOK,
			wantError: "",
		},
		{
			name:      "dealer cannot access dealers list",
			role:      "dealer",
			userID:    uuid.New().String(),
			wantCode:  http.StatusForbidden,
			wantError: "access denied",
		},
		{
			name:      "salon_manager cannot access dealers list",
			role:      "salon_manager",
			userID:    uuid.New().String(),
			wantCode:  http.StatusForbidden,
			wantError: "insufficient",
		},
		{
			name:      "no auth - unauthorized",
			role:      "",
			userID:    "",
			wantCode:  http.StatusUnauthorized,
			wantError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/dealers", nil)
			if tt.role != "" {
				req.Header.Set("X-Test-Role", tt.role)
				req.Header.Set("X-Test-User-ID", tt.userID)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantError != "" {
				assert.Contains(t, w.Body.String(), tt.wantError)
			}
		})
	}
}

func TestRLS_DealerAccessByUser(t *testing.T) {
	handler := NewRBACTestHandler()
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	authMiddleware := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			roleHeader := c.GetHeader("X-Test-Role")
			userIDHeader := c.GetHeader("X-Test-User-ID")
			if roleHeader == "" {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			c.Set("role", models.Role(roleHeader))
			c.Set("userID", userIDHeader)
			c.Next()
		}
	}
	
	r.GET("/api/v1/dealers/:id", authMiddleware(), handler.GetDealerByID)

	router := r

	dealerID1 := uuid.New()
	dealerID2 := uuid.New()
	userID := uuid.New()

	handler.dealers[dealerID1] = models.Dealer{
		ID:     dealerID1,
		UserID: userID,
		Name:  "My Dealer",
	}
	handler.dealers[dealerID2] = models.Dealer{
		ID:     dealerID2,
		UserID: uuid.New(),
		Name:  "Other Dealer",
	}

	tests := []struct {
		name       string
		userID     string
		dealerID   string
		wantCode   int
		wantResult string
	}{
		{
			name:       "dealer accesses own record",
			userID:     userID.String(),
			dealerID:   dealerID1.String(),
			wantCode:   http.StatusOK,
			wantResult: "My Dealer",
		},
		{
			name:       "dealer tries to access other dealer's record",
			userID:     userID.String(),
			dealerID:   dealerID2.String(),
			wantCode:   http.StatusForbidden,
			wantResult: "access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/dealers/"+tt.dealerID, nil)
			req.Header.Set("X-Test-Role", "dealer")
			req.Header.Set("X-Test-User-ID", tt.userID)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantResult != "" {
				assert.Contains(t, w.Body.String(), tt.wantResult)
			}
		})
	}
}

func TestForbiddenAccess(t *testing.T) {
	router := setupRBACRouter()

	req, _ := http.NewRequest("GET", "/api/v1/dealers", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)

	req2, _ := http.NewRequest("GET", "/api/v1/dealers", nil)
	req2.Header.Set("X-Test-Role", "invalid_role")
	req2.Header.Set("X-Test-User-ID", uuid.New().String())
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	require.Equal(t, http.StatusForbidden, w2.Code)
}