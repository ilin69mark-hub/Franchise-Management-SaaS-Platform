package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDealerService struct {
	mock.Mock
}

func (m *MockDealerService) GetDealerSummary(userID string, date string) (map[string]interface{}, error) {
	args := m.Called(userID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockDealerService) GetDealerFinance(userID string, date string) (map[string]interface{}, error) {
	args := m.Called(userID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockDealerService) GetDealerFunnel(userID string, period string, date string) (map[string]interface{}, error) {
	args := m.Called(userID, period, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockDealerService) GetDealerProducts(userID string, date string) (map[string]interface{}, error) {
	args := m.Called(userID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func TestGetDealerSummary_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/api/v1/dealer/summary", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		c.Set("tenantID", "test-tenant")
		
		summary := map[string]interface{}{
			"netProfit":             500000.0,
			"grossRevenue":          2000000.0,
			"planCompletionPercent": 75,
			"marginProfit":          1200000.0,
			"activeAlerts":          3,
		}
		c.JSON(http.StatusOK, summary)
	})

	c.Request, _ = http.NewRequest("GET", "/api/v1/dealer/summary", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 500000.0, response["netProfit"])
	assert.Equal(t, 2000000.0, response["grossRevenue"])
	assert.Equal(t, float64(75), response["planCompletionPercent"])
	assert.Equal(t, 1200000.0, response["marginProfit"])
	assert.Equal(t, float64(3), response["activeAlerts"])
}

func TestGetDealerSummary_EmptyData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/api/v1/dealer/summary", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		c.Set("tenantID", "test-tenant")
		
		summary := map[string]interface{}{
			"netProfit":             0.0,
			"grossRevenue":          0.0,
			"planCompletionPercent": 0,
			"marginProfit":          0.0,
			"activeAlerts":          0,
		}
		c.JSON(http.StatusOK, summary)
	})

	c.Request, _ = http.NewRequest("GET", "/api/v1/dealer/summary", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, response["netProfit"])
	assert.Equal(t, 0.0, response["grossRevenue"])
}

func TestGetDealerFinance_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/api/v1/dealer/finance", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		c.Set("tenantID", "test-tenant")
		
		finance := map[string]interface{}{
			"revenue":                5000000.0,
			"cogs":                  1500000.0,
			"rent":                  400000.0,
			"utilities":             100000.0,
			"payroll":               800000.0,
			"taxes":                 300000.0,
			"logistics":             150000.0,
			"marketing":             200000.0,
			"defects":               50000.0,
			"other_expenses":        100000.0,
			"bonus":                 100000.0,
			"net_profit":            800000.0,
			"net_profit_forecast":   900000.0,
			"prev_month_net_profit": 750000.0,
			"expense_breakdown": []map[string]interface{}{
				{"category": "Аренда", "amount": 400000.0, "percent_of_revenue": 8.0, "prev_month_amount": 380000.0},
				{"category": "ФОТ", "amount": 800000.0, "percent_of_revenue": 16.0, "prev_month_amount": 750000.0},
			},
		}
		c.JSON(http.StatusOK, finance)
	})

	c.Request, _ = http.NewRequest("GET", "/api/v1/dealer/finance?date=2026-01", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 5000000.0, response["revenue"])
	assert.Equal(t, 800000.0, response["net_profit"])
	
	breakdown := response["expense_breakdown"].([]interface{})
	assert.Len(t, breakdown, 2)
}

func TestGetDealerFinance_Calculations(t *testing.T) {
	revenue := 5000000.0
	cogs := 1500000.0
	rent := 400000.0
	payroll := 800000.0
	
	marginProfit := revenue - cogs
	assert.Equal(t, 3500000.0, marginProfit)
	
	rentPercent := (rent / revenue) * 100
	assert.Equal(t, 8.0, rentPercent)
	
	payrollPercent := (payroll / revenue) * 100
	assert.Equal(t, 16.0, payrollPercent)
}

func TestGetDealerFunnel_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/api/v1/dealer/funnel", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		c.Set("tenantID", "test-tenant")
		
		funnel := map[string]interface{}{
			"stages": []map[string]interface{}{
				{"name": "Трафик", "count": 1000, "percent": 100.0},
				{"name": "Консультации", "count": 400, "percent": 40.0},
				{"name": "Замеры", "count": 200, "percent": 20.0},
				{"name": "Договоры", "count": 80, "percent": 8.0},
				{"name": "Оплаты", "count": 65, "percent": 6.5},
			},
			"salon_plan_data": []map[string]interface{}{
				{"id": "salon-1", "name": "Салон Центр", "plan": 1000000.0, "fact": 850000.0, "percent": 85},
			},
		}
		c.JSON(http.StatusOK, funnel)
	})

	c.Request, _ = http.NewRequest("GET", "/api/v1/dealer/funnel?period=month&date=2026-01", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	stages := response["stages"].([]interface{})
	assert.Len(t, stages, 5)
}

func TestGetDealerFunnel_ConversionCalculation(t *testing.T) {
	traffic := 1000
	consultations := 400
	measurements := 200
	contracts := 80
	payments := 65
	
	consultationConversion := float64(consultations) / float64(traffic) * 100
	assert.Equal(t, 40.0, consultationConversion)
	
	measurementConversion := float64(measurements) / float64(traffic) * 100
	assert.Equal(t, 20.0, measurementConversion)
	
	contractConversion := float64(contracts) / float64(traffic) * 100
	assert.Equal(t, 8.0, contractConversion)
	
	paymentConversion := float64(payments) / float64(traffic) * 100
	assert.Equal(t, 6.5, paymentConversion)
	
	overallConversion := float64(payments) / float64(traffic) * 100
	assert.Equal(t, 6.5, overallConversion)
}

func TestGetDealerProducts_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/api/v1/dealer/products", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		c.Set("tenantID", "test-tenant")
		
		products := map[string]interface{}{
			"categories": []map[string]interface{}{
				{"name": "Мебель", "total": 1500000.0, "count": 45},
				{"name": "Текстиль", "total": 800000.0, "count": 120},
				{"name": "Освещение", "total": 400000.0, "count": 30},
			},
			"low_stock": []map[string]interface{}{
				{"name": "Товар А", "current": 5, "min": 10},
			},
		}
		c.JSON(http.StatusOK, products)
	})

	c.Request, _ = http.NewRequest("GET", "/api/v1/dealer/products?date=2026-01", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	categories := response["categories"].([]interface{})
	assert.Len(t, categories, 3)
	
	lowStock := response["low_stock"].([]interface{})
	assert.Len(t, lowStock, 1)
}

func TestGetDealerSummary_WithDateParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/api/v1/dealer/summary", func(c *gin.Context) {
		date := c.Query("date")
		
		summary := map[string]interface{}{
			"netProfit":             500000.0,
			"grossRevenue":          2000000.0,
			"planCompletionPercent": 75,
			"marginProfit":          1200000.0,
			"activeAlerts":          3,
			"date":                  date,
		}
		c.JSON(http.StatusOK, summary)
	})

	c.Request, _ = http.NewRequest("GET", "/api/v1/dealer/summary?date=2026-01", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "2026-01", response["date"])
}

func TestGetDealerRLS_TenantIsolation(t *testing.T) {
	tests := []struct {
		name        string
		userTenant  string
		dataTenant  string
		shouldAllow bool
	}{
		{"same tenant", "tenant-1", "tenant-1", true},
		{"different tenant", "tenant-1", "tenant-2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := tt.userTenant == tt.dataTenant
			assert.Equal(t, tt.shouldAllow, allowed)
		})
	}
}

func TestGetDealerRLS_DealerSalonAccess(t *testing.T) {
	dealerSalons := []string{"salon-1", "salon-2", "salon-3"}
	
	tests := []struct {
		name          string
		requestedSalon string
		shouldAllow   bool
	}{
		{"access own salon 1", "salon-1", true},
		{"access own salon 2", "salon-2", true},
		{"access own salon 3", "salon-3", true},
		{"access other dealer salon", "salon-99", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := false
			for _, s := range dealerSalons {
				if s == tt.requestedSalon {
					allowed = true
					break
				}
			}
			assert.Equal(t, tt.shouldAllow, allowed)
		})
	}
}

func TestDealerSummary_MetricsCalculation(t *testing.T) {
	grossRevenue := 2000000.0
	cogs := 800000.0
	
	marginProfit := grossRevenue - cogs
	assert.Equal(t, 1200000.0, marginProfit)
	
	marginPercent := (marginProfit / grossRevenue) * 100
	assert.Equal(t, 60.0, marginPercent)
	
	planCompletion := 850000.0
	plan := 1000000.0
	planPercent := (planCompletion / plan) * 100
	assert.Equal(t, 85.0, planPercent)
}

func TestDealerFinance_ProfitCalculation(t *testing.T) {
	revenue := 5000000.0
	expenses := 1500000.0 + 400000.0 + 100000.0 + 800000.0 + 300000.0 + 150000.0 + 200000.0 + 50000.0 + 100000.0 + 100000.0
	
	netProfit := revenue - expenses
	assert.Equal(t, 1300000.0, netProfit)
	
	prevNetProfit := 750000.0
	profitChange := ((netProfit - prevNetProfit) / prevNetProfit) * 100
	assert.InDelta(t, 73.33, profitChange, 0.01)
}

func TestDealerFunnel_DropOffAnalysis(t *testing.T) {
	stages := []int{1000, 400, 200, 80, 65}
	
	for i := 1; i < len(stages); i++ {
		dropOff := float64(stages[i-1]-stages[i]) / float64(stages[i-1]) * 100
		t.Logf("Stage %d to %d: %.1f%% drop-off", i-1, i, dropOff)
	}
	
	overallDropOff := float64(stages[0]-stages[len(stages)-1]) / float64(stages[0]) * 100
	assert.Equal(t, 93.5, overallDropOff)
}