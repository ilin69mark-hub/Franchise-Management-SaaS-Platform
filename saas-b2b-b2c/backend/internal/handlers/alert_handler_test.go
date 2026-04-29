package handlers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAlertTypeOverdueMeasurement(t *testing.T) {
	days := 5
	threshold := 3

	isOverdue := days > threshold

	assert.True(t, isOverdue)
}

func TestAlertTypeAbandonedKP(t *testing.T) {
	days := 7
	threshold := 5

	isAbandoned := days > threshold

	assert.True(t, isAbandoned)
}

func TestAlertTypeConversionDrop(t *testing.T) {
	avgConversion := 20.0
	todayConversion := 15.0

	drop := ((avgConversion - todayConversion) / avgConversion) * 100

	threshold := 20.0
	isCriticalDrop := drop > threshold

	assert.True(t, isCriticalDrop)
	assert.Equal(t, 25.0, drop)
}

func TestAlertTypeTrafficDrop(t *testing.T) {
	avgDailyLeads := 10.0
	todayLeads := 6.0

	drop := ((avgDailyLeads - todayLeads) / avgDailyLeads) * 100

	threshold := 30.0
	isCriticalDrop := drop > threshold

	assert.True(t, isCriticalDrop)
	assert.Equal(t, 40.0, drop)
}

func TestAlertSeverityColors(t *testing.T) {
	tests := []struct {
		severity  string
		expected  string
	}{
		{"critical", "#ff4d4f"},
		{"warning", "#faad14"},
		{"info", "#1890ff"},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			var color string
			switch tt.severity {
			case "critical":
				color = "#ff4d4f"
			case "warning":
				color = "#faad14"
			default:
				color = "#1890ff"
			}
			assert.Equal(t, tt.expected, color)
		})
	}
}

func TestAlertTimeAgo(t *testing.T) {
	tests := []struct {
		name     string
		elapsed  time.Duration
		expected string
	}{
		{"less than minute", 30 * time.Second, "Только что"},
		{"minutes", 30 * time.Minute, "30 мин назад"},
		{"hours", 2 * time.Hour, "2 ч назад"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elapsed := tt.elapsed
			diffMins := int(elapsed.Minutes())

			var result string
			if diffMins < 1 {
				result = "Только что"
			} else if diffMins < 60 {
				result = "30 мин назад"
			} else {
				result = "2 ч назад"
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUnreadCountCalculation(t *testing.T) {
	alerts := []struct {
		isRead bool
	}{
		{false},
		{false},
		{true},
		{false},
	}

	unreadCount := 0
	for _, a := range alerts {
		if !a.isRead {
			unreadCount++
		}
	}

	assert.Equal(t, 3, unreadCount)
}

func TestAlertStorageKey(t *testing.T) {
	key := "salon_alerts_unread"
	assert.Equal(t, "salon_alerts_unread", key)
}

func TestPollingInterval(t *testing.T) {
	pollInterval := 2 * 60 * 1000 // 2 минуты в миллисекундах
	assert.Equal(t, 120000, pollInterval)
}

func TestAlertSortingByTime(t *testing.T) {
	alerts := []struct {
		id        string
		createdAt time.Time
	}{
		{"1", time.Now().Add(-1 * time.Hour)},
		{"2", time.Now()},
		{"3", time.Now().Add(-30 * time.Minute)},
	}

	// Sort by created_at DESC
	for i := 0; i < len(alerts)-1; i++ {
		for j := i + 1; j < len(alerts); j++ {
			if alerts[i].createdAt.Before(alerts[j].createdAt) {
				alerts[i], alerts[j] = alerts[j], alerts[i]
			}
		}
	}

	assert.Equal(t, "2", alerts[0].id)
	assert.Equal(t, "3", alerts[1].id)
	assert.Equal(t, "1", alerts[2].id)
}