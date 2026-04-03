package services

import (
	"context"
	"time"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KPIService struct {
	DB       *gorm.DB // Публичное поле для доступа из хендлеров
	kpiRepo  *repository.KPIRepository
	schedRep *repository.ScheduleRepository
}

func NewKPIService(db *gorm.DB, kpiRepo *repository.KPIRepository, schedRep *repository.ScheduleRepository) *KPIService {
	return &KPIService{DB: db, kpiRepo: kpiRepo, schedRep: schedRep}
}

// SetGoal - обертка для репозитория
func (s *KPIService) SetGoal(ctx context.Context, goal *models.DailyGoal) error {
	return s.kpiRepo.UpsertGoal(ctx, goal)
}

func (s *KPIService) GetDashboardStats(ctx context.Context, userID uuid.UUID, salonID uuid.UUID, isManager bool) (*models.DashboardStatsResponse, error) {
	today := time.Now()
	todayStr := today.Format("2006-01-02")

	resp := &models.DashboardStatsResponse{Date: todayStr}

	// 1. Получаем план
	var goal *models.DailyGoal
	var err error

	if isManager {
		goal, err = s.kpiRepo.GetGoal(ctx, &userID, nil, today)
	} else {
		goal, err = s.kpiRepo.GetGoal(ctx, nil, &salonID, today)
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if goal == nil {
		goal = &models.DailyGoal{}
	}

	// 2. Считаем ФАКТ Продаж
	var salesFact float64
	qSales := s.DB.Model(&models.Lead{}).Where("status = ?", "sale")
	if isManager {
		qSales = qSales.Where("manager_id = ?", userID)
	} else {
		qSales = qSales.Where("salon_id = ?", salonID)
	}
	qSales.Where("DATE(created_at) = ?", todayStr).Select("COALESCE(SUM(budget), 0)").Scan(&salesFact)

	// 3. ФАКТ Лидов
	var leadsFact int64
	qLeads := s.DB.Model(&models.Lead{})
	if isManager {
		qLeads = qLeads.Where("manager_id = ?", userID)
	} else {
		qLeads = qLeads.Where("salon_id = ?", salonID)
	}
	qLeads.Where("DATE(created_at) = ?", todayStr).Count(&leadsFact)

	// 4. Звонки и Встречи
	var callsFact, meetingsFact int64
	actQuery := s.DB.Model(&models.LeadActivity{}).Where("DATE(created_at) = ?", todayStr)
	if isManager {
		actQuery = actQuery.Where("user_id = ?", userID)
	} else {
		actQuery = actQuery.Where("salon_id = ?", salonID)
	}

	actQuery.Where("type = ?", "call").Count(&callsFact)
	actQuery.Where("type = ?", "meeting").Count(&meetingsFact)

	// Формируем проценты
	resp.Sales = calcPercent(goal.SalesPlan, salesFact)
	resp.Leads = calcPercentInt(goal.LeadsPlan, leadsFact)
	resp.Calls = calcPercentInt(goal.CallsPlan, callsFact)
	resp.Meetings = calcPercentInt(goal.MeetingsPlan, meetingsFact)

	return resp, nil
}

// GetTeamAnalytics - аналитика для дилера
func (s *KPIService) GetTeamAnalytics(ctx context.Context, dealerID uuid.UUID, period string) ([]map[string]interface{}, error) {
	now := time.Now()
	var start, end time.Time

	switch period {
	case "week":
		start = now.AddDate(0, 0, -int(now.Weekday())+1).Truncate(24 * time.Hour) // Пн
		end = now
	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = now
	default: // day
		start = now.Truncate(24 * time.Hour)
		end = now
	}

	var managers []models.User
	if err := s.DB.Where("managed_by = ?", dealerID).Find(&managers).Error; err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for _, mgr := range managers {
		if mgr.SalonID == nil {
			continue
		}

		// --- KPI ---
		var totalSalesPlan float64
		var totalLeadsPlan, totalCallsPlan, totalMeetingsPlan int

		s.DB.Model(&models.DailyGoal{}).
			Where("user_id = ? AND target_date BETWEEN ? AND ?", mgr.ID, start, end).
			Select("COALESCE(SUM(sales_plan), 0)").Scan(&totalSalesPlan)
		s.DB.Model(&models.DailyGoal{}).
			Where("user_id = ? AND target_date BETWEEN ? AND ?", mgr.ID, start, end).
			Select("COALESCE(SUM(leads_plan), 0)").Scan(&totalLeadsPlan)
		s.DB.Model(&models.DailyGoal{}).
			Where("user_id = ? AND target_date BETWEEN ? AND ?", mgr.ID, start, end).
			Select("COALESCE(SUM(calls_plan), 0)").Scan(&totalCallsPlan)
		s.DB.Model(&models.DailyGoal{}).
			Where("user_id = ? AND target_date BETWEEN ? AND ?", mgr.ID, start, end).
			Select("COALESCE(SUM(meetings_plan), 0)").Scan(&totalMeetingsPlan)

		// Факты
		var salesFact float64
		s.DB.Model(&models.Lead{}).
			Where("manager_id = ? AND status = ? AND created_at BETWEEN ? AND ?", mgr.ID, "sale", start, end).
			Select("COALESCE(SUM(budget), 0)").Scan(&salesFact)

		var leadsFact int64
		s.DB.Model(&models.Lead{}).
			Where("manager_id = ? AND created_at BETWEEN ? AND ?", mgr.ID, start, end).
			Count(&leadsFact)

		var callsFact, meetingsFact int64
		s.DB.Model(&models.LeadActivity{}).
			Where("user_id = ? AND created_at BETWEEN ? AND ?", mgr.ID, start, end).
			Where("type = ?", "call").Count(&callsFact)
		s.DB.Model(&models.LeadActivity{}).
			Where("user_id = ? AND created_at BETWEEN ? AND ?", mgr.ID, start, end).
			Where("type = ?", "meeting").Count(&meetingsFact)

		// --- Воронка ---
		var lNew, lWork, lWait, lSale int64
		s.DB.Model(&models.Lead{}).Where("manager_id = ? AND status = ?", mgr.ID, "new").Count(&lNew)
		s.DB.Model(&models.Lead{}).Where("manager_id = ? AND status IN ?", mgr.ID, []string{"contact", "meeting"}).Count(&lWork)
		s.DB.Model(&models.Lead{}).Where("manager_id = ? AND status = ?", mgr.ID, "wait").Count(&lWait)
		s.DB.Model(&models.Lead{}).Where("manager_id = ? AND status = ?", mgr.ID, "sale").Count(&lSale)

		results = append(results, map[string]interface{}{
			"id":   mgr.ID,
			"name": mgr.FirstName + " " + mgr.LastName,
			"kpi": map[string]interface{}{
				"sales":    map[string]interface{}{"plan": totalSalesPlan, "fact": salesFact, "percent": calcPercentVal(totalSalesPlan, salesFact)},
				"leads":    map[string]interface{}{"plan": totalLeadsPlan, "fact": leadsFact, "percent": calcPercentVal(float64(totalLeadsPlan), float64(leadsFact))},
				"calls":    map[string]interface{}{"plan": totalCallsPlan, "fact": callsFact, "percent": calcPercentVal(float64(totalCallsPlan), float64(callsFact))},
				"meetings": map[string]interface{}{"plan": totalMeetingsPlan, "fact": meetingsFact, "percent": calcPercentVal(float64(totalMeetingsPlan), float64(meetingsFact))},
			},
			"funnel": map[string]interface{}{
				"entered": leadsFact,
				"in_work": lWork,
				"waiting": lWait,
				"sale":    lSale,
			},
		})
	}

	return results, nil
}

func calcPercent(plan float64, fact float64) models.KPIItem {
	p := 0
	if plan > 0 {
		p = int((fact / plan) * 100)
		if p > 100 {
			p = 100
		}
	}
	return models.KPIItem{Plan: plan, Fact: fact, Percent: p}
}

func calcPercentInt(plan int, fact int64) models.KPIItem {
	return calcPercent(float64(plan), float64(fact))
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
