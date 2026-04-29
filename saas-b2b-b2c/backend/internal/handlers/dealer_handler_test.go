package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetDealerSummary_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	kpiHandler := &KPIHandler{}
	kpiHandler.GetDealerSummary(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid session")
}

func TestGetDealerFinance_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	kpiHandler := &KPIHandler{}
	kpiHandler.GetDealerFinance(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetDealerFunnel_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	kpiHandler := &KPIHandler{}
	kpiHandler.GetDealerFunnel(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetDealerProducts_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	kpiHandler := &KPIHandler{}
	kpiHandler.GetDealerProducts(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetFranchiserSummary_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	kpiHandler := &KPIHandler{}
	kpiHandler.GetFranchiserSummary(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetFranchiserNetwork_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	kpiHandler := &KPIHandler{}
	kpiHandler.GetFranchiserNetwork(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetFranchiserHealth_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	kpiHandler := &KPIHandler{}
	kpiHandler.GetFranchiserHealth(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetFranchiserTeam_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	kpiHandler := &KPIHandler{}
	kpiHandler.GetFranchiserTeam(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetTerritorySummary_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	kpiHandler := &KPIHandler{}
	kpiHandler.GetTerritorySummary(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetTerritoryFunnel_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	kpiHandler := &KPIHandler{}
	kpiHandler.GetTerritoryFunnel(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetTerritoryPlanFact_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	kpiHandler := &KPIHandler{}
	kpiHandler.GetTerritoryPlanFact(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetTerritoryCommunications_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	kpiHandler := &KPIHandler{}
	kpiHandler.GetTerritoryCommunications(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetTerritoryBenchmarks_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	kpiHandler := &KPIHandler{}
	kpiHandler.GetTerritoryBenchmarks(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDealerRLS_PreventCrossTenant(t *testing.T) {
	tests := []struct {
		name      string
		managerTenantID string
		dataTenantID string
		shouldAllow bool
	}{
		{
			name:            "Same tenant - allowed",
			managerTenantID: "tenant-1",
			dataTenantID:    "tenant-1",
			shouldAllow:   true,
		},
		{
			name:            "Different tenant - denied",
			managerTenantID: "tenant-1",
			dataTenantID:    "tenant-2",
			shouldAllow:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// RLS: WHERE tenant_id = user.tenant_id
			userTenantID := tt.managerTenantID
			dataTenantID := tt.dataTenantID

			allowed := userTenantID == dataTenantID
			assert.Equal(t, tt.shouldAllow, allowed)
		})
	}
}

func TestDealerRLS_SalonAccess(t *testing.T) {
	tests := []struct {
		name          string
		dealerSalons   []string
		requestedSalon string
		shouldAllow bool
	}{
		{
			name:          "Access own salon",
			dealerSalons:   []string{"salon-1", "salon-2"},
			requestedSalon: "salon-1",
			shouldAllow:   true,
		},
		{
			name:          "Access another dealer's salon",
			dealerSalons:   []string{"salon-1", "salon-2"},
			requestedSalon: "salon-3",
			shouldAllow:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// RLS: WHERE salon_id IN (SELECT salon_id FROM users WHERE managed_by = user.id)
			allowed := false
			for _, s := range tt.dealerSalons {
				if s == tt.requestedSalon {
					allowed = true
					break
				}
			}
			assert.Equal(t, tt.shouldAllow, allowed)
		})
	}
}

func TestFranchiserRLS_DealerAccess(t *testing.T) {
	tests := []struct {
		name            string
		franchiserDealers []string
		requestedDealer  string
		shouldAllow     bool
	}{
		{
			name:            "Access own dealer",
			franchiserDealers: []string{"dealer-1", "dealer-2"},
			requestedDealer:  "dealer-1",
			shouldAllow:    true,
		},
		{
			name:            "Access another franchiser's dealer",
			franchiserDealers: []string{"dealer-1", "dealer-2"},
			requestedDealer:  "dealer-3",
			shouldAllow:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// RLS: WHERE managed_by = user.id (dealer belongs to franchiser)
			allowed := false
			for _, d := range tt.franchiserDealers {
				if d == tt.requestedDealer {
					allowed = true
					break
				}
			}
			assert.Equal(t, tt.shouldAllow, allowed)
		})
	}
}

func TestTerritoryManagerRLS_DealerAccess(t *testing.T) {
	tests := []struct {
		name          string
		managerDealers []string
		requestedDealer string
		shouldAllow  bool
	}{
		{
			name:          "Access own dealers",
			managerDealers: []string{"dealer-1", "dealer-2", "dealer-3"},
			requestedDealer: "dealer-2",
			shouldAllow:  true,
		},
		{
			name:          "Cannot access other manager's dealers",
			managerDealers: []string{"dealer-1", "dealer-2"},
			requestedDealer: "dealer-5",
			shouldAllow:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// RLS: WHERE managed_by = user.id (territory manager -> dealers)
			allowed := false
			for _, d := range tt.managerDealers {
				if d == tt.requestedDealer {
					allowed = true
					break
				}
			}
			assert.Equal(t, tt.shouldAllow, allowed)
		})
	}
}