package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestProductsCalculation(t *testing.T) {
	tests := []struct {
		name          string
		totalRevenue  float64
		productRevenue float64
		expectedShare float64
	}{
		{"100% share", 1000000, 1000000, 100},
		{"50% share", 1000000, 500000, 50},
		{"25% share", 1000000, 250000, 25},
		{"0% share", 1000000, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			share := 0.0
			if tt.totalRevenue > 0 {
				share = (tt.productRevenue / tt.totalRevenue) * 100
			}
			assert.InDelta(t, tt.expectedShare, share, 0.1)
		})
	}
}

func TestStockTurnover(t *testing.T) {
	tests := []struct {
		name        string
		days        int
		isSlowMoving bool
	}{
		{"Fast moving", 30, false},
		{"Medium", 60, false},
		{"Slow", 90, false},
		{"Non-liquid", 120, true},
		{"Very slow", 180, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isSlowMoving := tt.days > 90
			assert.Equal(t, tt.isSlowMoving, isSlowMoving)
		})
	}
}

func TestLostSalesThreshold(t *testing.T) {
	threshold := 100000.0

	tests := []struct {
		name     string
		revenue  float64
		expected bool
	}{
		{"Critical", 150000, true},
		{"At threshold", 100000, false},
		{"Below threshold", 80000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isCritical := tt.revenue > threshold
			assert.Equal(t, tt.expected, isCritical)
		})
	}
}

func TestProductsSortByRevenue(t *testing.T) {
	products := []struct {
		Name   string
		Revenue float64
	}{
		{"Product A", 500000},
		{"Product B", 300000},
		{"Product C", 700000},
	}

	// Sort by revenue descending
	for i := 0; i < len(products)-1; i++ {
		for j := i + 1; j < len(products); j++ {
			if products[i].Revenue < products[j].Revenue {
				products[i], products[j] = products[j], products[i]
			}
		}
	}

	assert.Equal(t, "Product C", products[0].Name)
	assert.Equal(t, "Product A", products[1].Name)
	assert.Equal(t, "Product B", products[2].Name)
}

func TestProductsSortByQuantity(t *testing.T) {
	products := []struct {
		Name    string
		Quantity int
	}{
		{"Product A", 10},
		{"Product B", 25},
		{"Product C", 5},
	}

	// Sort by quantity descending
	for i := 0; i < len(products)-1; i++ {
		for j := i + 1; j < len(products); j++ {
			if products[i].Quantity < products[j].Quantity {
				products[i], products[j] = products[j], products[i]
			}
		}
	}

	assert.Equal(t, "Product B", products[0].Name)
	assert.Equal(t, "Product A", products[1].Name)
	assert.Equal(t, "Product C", products[2].Name)
}

func TestProductsSortByMargin(t *testing.T) {
	products := []struct {
		Name   string
		Margin float64
	}{
		{"Product A", 20},
		{"Product B", 35},
		{"Product C", 15},
	}

	// Sort by margin descending
	for i := 0; i < len(products)-1; i++ {
		for j := i + 1; j < len(products); j++ {
			if products[i].Margin < products[j].Margin {
				products[i], products[j] = products[j], products[i]
			}
		}
	}

	assert.Equal(t, "Product B", products[0].Name)
	assert.Equal(t, "Product A", products[1].Name)
	assert.Equal(t, "Product C", products[2].Name)
}

func TestDashboardProductsResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	kpiHandler := &KPIHandler{}
	kpiHandler.GetDashboardProducts(c)

	// Should return 401 for unauthenticated user
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid session")
}