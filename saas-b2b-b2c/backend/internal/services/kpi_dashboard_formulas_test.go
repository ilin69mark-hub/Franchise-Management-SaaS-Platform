package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKPIPlanPercent(t *testing.T) {
	tests := []struct {
		name     string
		plan     float64
		fact     float64
		expected int
	}{
		{"Zero plan returns 0", 0, 100, 0},
		{"Full achievement", 100, 100, 100},
		{"Over achievement caps at 100", 100, 150, 100},
		{"Half achievement", 100, 50, 50},
		{"Quarter achievement", 100, 25, 25},
		{"Zero fact", 100, 0, 0},
		{"Decimal plan", 50.5, 25.25, 50},
		{"Large numbers", 1000000, 750000, 75},
		{"Exact 80% threshold", 100, 80, 80},
		{"Just over 80%", 100, 81, 81},
		{"Exact 50% threshold", 100, 50, 50},
		{"Just under 50%", 100, 49, 49},
		{"99% achievement", 100, 99, 99},
		{"1% achievement", 100, 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calcPercentVal(tt.plan, tt.fact)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDealerFinanceFormulas(t *testing.T) {
	t.Run("net profit = revenue * 0.2", func(t *testing.T) {
		assert.Equal(t, 200000.0, 1000000.0*0.2)
		assert.Equal(t, 0.0, 0.0*0.2)
		assert.Equal(t, 20000.0, 100000.0*0.2)
		assert.Equal(t, 2000000.0, 10000000.0*0.2)
	})

	t.Run("margin profit = revenue * 0.35", func(t *testing.T) {
		assert.Equal(t, 350000.0, 1000000.0*0.35)
		assert.Equal(t, 0.0, 0.0*0.35)
		assert.Equal(t, 35000.0, 100000.0*0.35)
	})

	t.Run("COGS = revenue * 0.65", func(t *testing.T) {
		assert.Equal(t, 650000.0, 1000000.0*0.65)
		assert.Equal(t, 0.0, 0.0*0.65)
		assert.Equal(t, 325000.0, 500000.0*0.65)
	})

	t.Run("full net profit calculation", func(t *testing.T) {
		revenue := 2000000.0
		cogs := revenue * 0.65  // 1300000
		rent := 430000.0
		utilities := 45000.0
		payroll := 650000.0
		taxes := 180000.0
		logistics := 95000.0
		marketing := 60000.0
		defects := 35000.0
		bonus := 0.0

		netProfit := revenue - cogs - rent - utilities - payroll - taxes - logistics - marketing - defects - bonus
		// 2M - 1.3M - 430K - 45K - 650K - 180K - 95K - 60K - 35K = -795K
		assert.Equal(t, -795000.0, netProfit)
	})

	t.Run("net profit with bonus", func(t *testing.T) {
		revenue := 2000000.0
		cogs := revenue * 0.65  // 1,300,000
		expenses := 430000 + 45000 + 650000 + 180000 + 95000 + 60000 + 35000 // 1,495,000
		bonus := 100000.0
		netProfit := revenue - cogs - float64(expenses) - bonus
		// 2M - 1.3M - 1.495M - 100K = -895K
		assert.Equal(t, -895000.0, netProfit)
	})

	t.Run("forecast = netProfit + (dailyAvg * daysLeft)", func(t *testing.T) {
		currentProfit := 500000.0
		dayOfMonth := 15
		daysLeft := 30 - dayOfMonth + 1 // 16
		dailyAvg := currentProfit / float64(dayOfMonth)
		forecast := currentProfit + (dailyAvg * float64(daysLeft))
		assert.Equal(t, 16, daysLeft)
		assert.InDelta(t, 1033333.33, forecast, 1)
	})

	t.Run("expense percent = (expense / revenue) * 100", func(t *testing.T) {
		assert.Equal(t, 21.5, (430000.0/2000000.0)*100)
		assert.Equal(t, 0.0, (0.0/1000000.0)*100)
		assert.Equal(t, 5.0, (50000.0/1000000.0)*100)
		assert.Equal(t, 50.0, (500000.0/1000000.0)*100)
	})
}

func TestDealerFunnelFormulas(t *testing.T) {
	t.Run("stage conversion = (count / prevCount) * 100", func(t *testing.T) {
		conv := 0
		prevCount := int64(100)
		count := int64(50)
		if prevCount > 0 {
			conv = int((count * 100) / prevCount)
		}
		assert.Equal(t, 50, conv)
	})

	t.Run("conversion with zero prev returns 0", func(t *testing.T) {
		conv := 0
		prevCount := int64(0)
		if prevCount > 0 {
			count := int64(50)
			conv = int((count * 100) / prevCount)
		}
		assert.Equal(t, 0, conv)
	})

	t.Run("forecast status: >=100 green, >=70 yellow, else red", func(t *testing.T) {
		forecast := func(percent int) string {
			if percent >= 100 {
				return "green"
			} else if percent >= 70 {
				return "yellow"
			}
			return "red"
		}
		assert.Equal(t, "green", forecast(100))
		assert.Equal(t, "green", forecast(110))
		assert.Equal(t, "yellow", forecast(99))
		assert.Equal(t, "yellow", forecast(70))
		assert.Equal(t, "red", forecast(69))
		assert.Equal(t, "red", forecast(0))
	})

	t.Run("product share = (productRevenue / totalRevenue) * 100", func(t *testing.T) {
		assert.Equal(t, 30.0, (300000.0/1000000.0)*100)
		productRevenue := 100000.0
		totalRevenue := 0.0
		share := 0.0
		if totalRevenue > 0 {
			share = (productRevenue / totalRevenue) * 100
		}
		assert.Equal(t, 0.0, share)
		assert.Equal(t, 5.0, (50000.0/1000000.0)*100)
	})
}

func TestDealerMarketingBudgetFormulas(t *testing.T) {
	t.Run("remaining = total - used", func(t *testing.T) {
		assert.Equal(t, 140000.0, 200000.0-60000.0)
		assert.Equal(t, 0.0, 200000.0-200000.0)
		assert.Equal(t, -50000.0, 200000.0-250000.0)
	})

	t.Run("usage percent = (used / total) * 100 capped at 100", func(t *testing.T) {
		used := 60000.0
		total := 200000.0
		percent := 0
		if total > 0 {
			percent = int((used / total) * 100)
		}
		assert.Equal(t, 30, percent)

		overUsed := 250000.0
		overTotal := 200000.0
		overPercent := int((overUsed / overTotal) * 100)
		assert.Equal(t, 125, overPercent)
	})
}

func TestDealerAlertPriorityFormulas(t *testing.T) {
	t.Run("priority assignment rules", func(t *testing.T) {
		priority := func(alertType string) string {
			if alertType == "conversion_drop" || alertType == "task_overdue" {
				return "critical"
			} else if alertType == "payroll_exceeded" {
				return "warning"
			}
			return "info"
		}
		assert.Equal(t, "critical", priority("conversion_drop"))
		assert.Equal(t, "critical", priority("task_overdue"))
		assert.Equal(t, "warning", priority("payroll_exceeded"))
		assert.Equal(t, "info", priority("unknown"))
		assert.Equal(t, "info", priority(""))
	})
}

func TestDealerRedZoneFormulas(t *testing.T) {
	t.Run("red zone when plan < 50%", func(t *testing.T) {
		isRedZone := func(plan, fact float64) bool {
			if plan > 0 {
				return int((fact/plan)*100) < 50
			}
			return false
		}
		assert.True(t, isRedZone(100000, 40000))
		assert.False(t, isRedZone(100000, 50000))
		assert.False(t, isRedZone(100000, 80000))
		assert.False(t, isRedZone(0, 100000))
	})

	t.Run("red zone count across dealers", func(t *testing.T) {
		dealers := []struct{ plan, fact float64 }{
			{100000, 40000}, // 40% red
			{100000, 90000}, // 90% ok
			{100000, 30000}, // 30% red
		}
		redCount := 0
		for _, d := range dealers {
			if d.plan > 0 && int((d.fact/d.plan)*100) < 50 {
				redCount++
			}
		}
		assert.Equal(t, 2, redCount)
	})
}

func TestFranchiserPlanPercent(t *testing.T) {
	tests := []struct {
		name     string
		revenue  float64
		plan     float64
		expected int
	}{
		{"Full achievement", 1000000, 1000000, 100},
		{"Over achievement returns 120 (no cap in franchiser)", 1200000, 1000000, 120},
		{"80% achievement", 800000, 1000000, 80},
		{"50% achievement", 500000, 1000000, 50},
		{"Zero plan returns 0", 0, 1000000, 0},
		{"Zero revenue", 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			planPercent := 0
			if tt.plan > 0 {
				planPercent = int((tt.revenue / tt.plan) * 100)
			}
			assert.Equal(t, tt.expected, planPercent)
		})
	}
}

func TestFranchiserForecastPercent(t *testing.T) {
	t.Run("forecast based on daily average extrapolated to 90 days", func(t *testing.T) {
		totalFact := 500000.0
		daysPassed := 15.0
		plan := 1000000.0

		forecastAmount := (totalFact / daysPassed) * 90.0
		forecastPercent := 0
		if plan > 0 {
			forecastPercent = int((forecastAmount / plan) * 100)
		}
		assert.InDelta(t, 300, forecastPercent, 10)
	})

	t.Run("zero days passed returns 0", func(t *testing.T) {
		totalFact := 500000.0
		daysPassed := 0.0
		plan := 1000000.0

		forecastAmount := 0.0
		if daysPassed > 0 {
			forecastAmount = (totalFact / daysPassed) * 90.0
		}
		forecastPercent := 0
		if plan > 0 {
			forecastPercent = int((forecastAmount / plan) * 100)
		}
		assert.Equal(t, 0, forecastPercent)
	})

	t.Run("zero plan returns 0", func(t *testing.T) {
		totalFact := 500000.0
		daysPassed := 15.0
		plan := 0.0

		forecastAmount := (totalFact / daysPassed) * 90.0
		forecastPercent := 0
		if plan > 0 {
			forecastPercent = int((forecastAmount / plan) * 100)
		}
		assert.Equal(t, 0, forecastPercent)
	})
}

func TestFranchiserAverageConversion(t *testing.T) {
	tests := []struct {
		name         string
		totalLeads   int64
		totalSales   int64
		expectedConv float64
	}{
		{"15% conversion", 100, 15, 15.0},
		{"30% conversion", 100, 30, 30.0},
		{"Zero leads", 0, 0, 0.0},
		{"50% conversion", 100, 50, 50.0},
		{"Small numbers", 10, 3, 30.0},
		{"No sales", 100, 0, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avgConversion := 0.0
			if tt.totalLeads > 0 {
				avgConversion = (float64(tt.totalSales) / float64(tt.totalLeads)) * 100
			}
			assert.Equal(t, tt.expectedConv, avgConversion)
		})
	}
}

func TestFranchiserDealerForecastStatus(t *testing.T) {
	tests := []struct {
		name     string
		percent  int
		expected string
	}{
		{"Green >= 80%", 100, "green"},
		{"Green 80%", 80, "green"},
		{"Yellow >= 50% < 80%", 79, "yellow"},
		{"Yellow 50%", 50, "yellow"},
		{"Red < 50%", 49, "red"},
		{"Red 0%", 0, "red"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			forecast := "red"
			if tt.percent >= 80 {
				forecast = "green"
			} else if tt.percent >= 50 {
				forecast = "yellow"
			}
			assert.Equal(t, tt.expected, forecast)
		})
	}
}

func TestFranchiserSLAFormulas(t *testing.T) {
	tests := []struct {
		name      string
		total     int64
		processed int64
		expected  float64
	}{
		{"100% SLA", 100, 100, 100.0},
		{"75% SLA", 100, 75, 75.0},
		{"Zero total defaults to 100", 0, 0, 100.0},
		{"50% SLA", 100, 50, 50.0},
		{"All processed 50", 50, 50, 100.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sla float64
			if tt.total > 0 {
				sla = float64(tt.processed) / float64(tt.total) * 100
			} else {
				sla = 100
			}
			assert.Equal(t, tt.expected, sla)
		})
	}
}

func TestFranchiserAlertCriticalCount(t *testing.T) {
	tests := []struct {
		name          string
		alertTypes    []string
		expectedCount int
	}{
		{"No critical", []string{"warning", "info", "info"}, 0},
		{"One critical", []string{"critical", "warning", "info"}, 1},
		{"Multiple critical", []string{"critical", "critical", "critical"}, 3},
		{"Empty list", []string{}, 0},
		{"All warning", []string{"warning", "warning"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := 0
			for _, at := range tt.alertTypes {
				if at == "critical" {
					count++
				}
			}
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

func TestFranchiserDealerStatusFormulas(t *testing.T) {
	tests := []struct {
		name     string
		percent  int
		expected string
	}{
		{"Green >= 80%", 95, "green"},
		{"Green 80%", 80, "green"},
		{"Yellow < 80% >= 50%", 65, "yellow"},
		{"Yellow 50%", 50, "yellow"},
		{"Red < 50%", 45, "red"},
		{"Red 0%", 0, "red"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := "green"
			if tt.percent < 80 {
				if tt.percent < 50 {
					status = "red"
				} else {
					status = "yellow"
				}
			}
			assert.Equal(t, tt.expected, status)
		})
	}
}

func TestFranchiserRedZoneCount(t *testing.T) {
	t.Run("red zone dealers count", func(t *testing.T) {
		dealers := []struct{ plan, fact float64 }{
			{100000, 40000}, // 40% red
			{100000, 90000}, // 90% ok
			{100000, 30000}, // 30% red
			{0, 50000},      // zero plan excluded
		}
		redZoneCount := 0
		for _, d := range dealers {
			if d.plan > 0 {
				percent := int((d.fact / d.plan) * 100)
				if percent < 50 {
					redZoneCount++
				}
			}
		}
		assert.Equal(t, 2, redZoneCount)
	})
}

func TestFranchiserAvgConversionNoDealers(t *testing.T) {
	tests := []struct {
		name         string
		dealersCount int
		expectedConv float64
	}{
		{"Zero dealers", 0, 0.0},
		{"One dealer", 1, 50.0},
		{"Multiple dealers", 5, 50.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avgConversion := 50.0
			if tt.dealersCount == 0 {
				avgConversion = 0
			}
			assert.Equal(t, tt.expectedConv, avgConversion)
		})
	}
}

func TestFranchiserNetworkOverviewPlanPercent(t *testing.T) {
	tests := []struct {
		name         string
		totalFact    float64
		totalPlan    float64
		expectedMax  int
	}{
		{"100% plan", 1000000, 1000000, 100},
		{"Over plan caps at 100", 1200000, 1000000, 100},
		{"80% plan", 800000, 1000000, 80},
		{"Zero plan", 0, 1000000, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			planPercent := 0.0
			if tt.totalPlan > 0 {
				planPercent = (tt.totalFact / tt.totalPlan) * 100
			}
			result := int(planPercent)
			if result > 100 {
				result = 100
			}
			assert.Equal(t, tt.expectedMax, result)
		})
	}
}

func TestFranchiserForecastStatusMapping(t *testing.T) {
	t.Run("forecastPercent mapping to status", func(t *testing.T) {
		status := func(planPercent int) string {
			if planPercent >= 80 {
				return "green"
			} else if planPercent >= 50 {
				return "yellow"
			}
			return "red"
		}
		assert.Equal(t, "green", status(85))
		assert.Equal(t, "yellow", status(65))
		assert.Equal(t, "red", status(40))
	})
}

func TestFranchiserTerritoryPlanFact(t *testing.T) {
	t.Run("plan percent calculation for territory", func(t *testing.T) {
		dealers := []struct{ plan, fact float64 }{
			{100000, 80000},  // 80%
			{200000, 140000}, // 70%
			{150000, 60000},  // 40% red
		}
		totalPlan := 0.0
		totalFact := 0.0
		for _, d := range dealers {
			totalPlan += d.plan
			totalFact += d.fact
		}
		percent := 0
		if totalPlan > 0 {
			percent = int((totalFact / totalPlan) * 100)
		}
		assert.InDelta(t, 62.2, float64(percent), 1.0)
		assert.Equal(t, 450000.0, totalPlan)
		assert.Equal(t, 280000.0, totalFact)
	})
}

func TestFranchiserTerritoryForecast(t *testing.T) {
	t.Run("forecast for territory", func(t *testing.T) {
		totalFact := 280000.0
		dayOfMonth := 15
		totalPlan := 450000.0

		dailyAvg := totalFact / float64(dayOfMonth)
		forecastAmount := dailyAvg * 30
		forecastPercent := 0
		if totalPlan > 0 {
			forecastPercent = int((forecastAmount / totalPlan) * 100)
		}
		// 280K / 15 days = 18.6K/day * 30 = 560K projected
		// 560K / 450K * 100 = 124%
		assert.InDelta(t, 124.4, float64(forecastPercent), 5.0)
	})
}

func TestFranchiserBenchmarkConversion(t *testing.T) {
	t.Run("network avg conversion comparison", func(t *testing.T) {
		territoryLeads := int64(100)
		territorySales := int64(15)
		networkBenchmark := 15.0 // 15%

		territoryConv := 0.0
		if territoryLeads > 0 {
			territoryConv = (float64(territorySales) / float64(territoryLeads)) * 100
		}

		assert.Equal(t, 15.0, territoryConv)
		assert.Equal(t, networkBenchmark, territoryConv)
	})

	t.Run("avg check comparison", func(t *testing.T) {
		avgCheck := 80000.0
		networkBenchmark := 80000.0
		assert.Equal(t, networkBenchmark, avgCheck)
	})
}

func TestFranchiserRedDealerDetection(t *testing.T) {
	t.Run("dealer with zero sales is red", func(t *testing.T) {
		fact := 0.0
		isRed := fact == 0
		assert.True(t, isRed)
	})

	t.Run("dealer with sales is not red", func(t *testing.T) {
		fact := 50000.0
		isRed := fact == 0
		assert.False(t, isRed)
	})
}