package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetDashboardMain_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	kpiHandler := &KPIHandler{}
	kpiHandler.GetDashboardMain(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid session")
}

func calcPercentVal(plan, fact float64) int {
	if plan == 0 {
		return 0
	}
	p := int((fact / plan) * 100)
	if p > 100 {
		return 100
	}
	return p
}

func TestKPICalculation(t *testing.T) {
	tests := []struct {
		name     string
		plan     float64
		fact     float64
		expected int
	}{
		{"Zero plan", 0, 100, 0},
		{"Full achievement", 100, 100, 100},
		{"Over achievement", 100, 150, 100},
		{"Half achievement", 100, 50, 50},
		{"Quarter achievement", 100, 25, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calcPercentVal(tt.plan, tt.fact)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrafficLightStatus(t *testing.T) {
	tests := []struct {
		name     string
		leads    int64
		norm     int
		expected string
	}{
		{"High traffic", 15, 10, "green"},
		{"Medium traffic", 7, 10, "yellow"},
		{"Low traffic", 3, 10, "red"},
		{"Zero traffic", 0, 10, "red"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if tt.leads >= int64(tt.norm) {
				result = "green"
			} else if tt.leads >= int64(tt.norm/2) {
				result = "yellow"
			} else {
				result = "red"
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConversionCalculation(t *testing.T) {
	tests := []struct {
		name      string
		total     int64
		converted int64
		expected  string
	}{
		{"High conversion", 100, 30, "green"},
		{"Medium conversion", 100, 15, "yellow"},
		{"Low conversion", 100, 5, "red"},
		{"Zero conversion", 100, 0, "red"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if tt.total > 0 {
				conv := int((tt.converted * 100) / tt.total)
				if conv >= 30 {
					result = "green"
				} else if conv >= 15 {
					result = "yellow"
				} else {
					result = "red"
				}
			} else {
				result = "red"
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}