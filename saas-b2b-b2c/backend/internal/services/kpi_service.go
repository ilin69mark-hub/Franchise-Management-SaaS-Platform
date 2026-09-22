package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KPIService struct {
	DB            *gorm.DB // Публичное поле для доступа из хендлеров
	kpiRepo       repository.KPIRepositoryInterface
	schedRep      repository.ScheduleRepositoryInterface
	analyticsRepo repository.AnalyticsRepositoryInterface
}

func NewKPIService(db *gorm.DB, kpiRepo repository.KPIRepositoryInterface, schedRep repository.ScheduleRepositoryInterface) *KPIService {
	return &KPIService{DB: db, kpiRepo: kpiRepo, schedRep: schedRep}
}

func NewKPIServiceWithAnalytics(db *gorm.DB, kpiRepo repository.KPIRepositoryInterface, schedRep repository.ScheduleRepositoryInterface, analyticsRepo repository.AnalyticsRepositoryInterface) *KPIService {
	return &KPIService{DB: db, kpiRepo: kpiRepo, schedRep: schedRep, analyticsRepo: analyticsRepo}
}

func (s *KPIService) getSettingFloat(key string, def float64) float64 {
	var v string
	if err := s.DB.Table("system_settings").Where("key = ?", key).Select("value").Scan(&v).Error; err == nil && v != "" {
		if parsed, perr := strconv.ParseFloat(v, 64); perr == nil {
			return parsed
		}
	}
	return def
}

// SetGoal - обертка для репозитория
func (s *KPIService) SetGoal(ctx context.Context, goal *models.DailyGoal) error {
	return s.kpiRepo.UpsertGoal(ctx, goal)
}

func (s *KPIService) GetDashboardStats(ctx context.Context, userID uuid.UUID, salonID uuid.UUID, isManager bool) (*models.DashboardStatsResponse, error) {
	if s.analyticsRepo != nil {
		return s.analyticsRepo.CalculateDashboardStats(ctx, &userID, &salonID, isManager)
	}

	today := time.Now()
	todayStr := today.Format("2006-01-02")
	dayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)

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
	qSales.Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).Select("COALESCE(SUM(budget), 0)").Scan(&salesFact)

	// 3. ФАКТ Лидов
	var leadsFact int64
	qLeads := s.DB.Model(&models.Lead{})
	if isManager {
		qLeads = qLeads.Where("manager_id = ?", userID)
	} else {
		qLeads = qLeads.Where("salon_id = ?", salonID)
	}
	qLeads.Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).Count(&leadsFact)

	// 4. Звонки и Встречи
	var callsFact, meetingsFact int64
	countActivity := func(activityType string, dest *int64) {
		q := s.DB.Model(&models.LeadActivity{}).
			Where("type = ?", activityType).
			Where("created_at >= ? AND created_at < ?", dayStart, dayEnd)
		if isManager {
			q = q.Where("user_id = ?", userID)
		} else {
			q = q.Where("EXISTS (SELECT 1 FROM leads WHERE leads.id = lead_activities.lead_id AND leads.salon_id = ?)", salonID)
		}
		q.Count(dest)
	}
	countActivity("call", &callsFact)
	countActivity("meeting", &meetingsFact)

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
	if err := s.DB.Where("managed_by = ?", dealerID).Limit(100).Find(&managers).Error; err != nil {
		return nil, err
	}

	// Батчим N+1: 11*N → 4 запроса (daily_goals, leads sales+count, activities, funnel)
	type goalAgg struct {
		UserID         uuid.UUID `gorm:"column:user_id"`
		TotalSalesPlan float64   `gorm:"column:total_sales"`
		TotalLeadsPlan int       `gorm:"column:total_leads"`
		TotalCallsPlan int       `gorm:"column:total_calls"`
		TotalMeetPlan  int       `gorm:"column:total_meet"`
	}
	type leadAgg struct {
		ManagerID uuid.UUID `gorm:"column:manager_id"`
		SalesFact float64   `gorm:"column:sales_fact"`
		SalesCnt  int64     `gorm:"column:sales_cnt"`
		LeadsFact int64     `gorm:"column:leads_fact"`
	}
	type actAgg struct {
		UserID uuid.UUID `gorm:"column:user_id"`
		Type   string    `gorm:"column:type"`
		Cnt    int64     `gorm:"column:cnt"`
	}
	type funnelAgg struct {
		ManagerID uuid.UUID `gorm:"column:manager_id"`
		Status    string    `gorm:"column:status"`
		Cnt       int64     `gorm:"column:cnt"`
	}
	var managerIDs []uuid.UUID
	for _, m := range managers {
		if m.SalonID != nil {
			managerIDs = append(managerIDs, m.ID)
		}
	}
	goalMap := map[uuid.UUID]goalAgg{}
	leadMap := map[uuid.UUID]leadAgg{}
	actMap := map[uuid.UUID]map[string]int64{}
	funnelMap := map[uuid.UUID]map[string]int64{}
	if len(managerIDs) > 0 {
		var goals []goalAgg
		s.DB.Model(&models.DailyGoal{}).
			Select("user_id, COALESCE(SUM(sales_plan),0) as total_sales, COALESCE(SUM(leads_plan),0) as total_leads, COALESCE(SUM(calls_plan),0) as total_calls, COALESCE(SUM(meetings_plan),0) as total_meet").
			Where("user_id IN ? AND target_date BETWEEN ? AND ?", managerIDs, start, end).
			Group("user_id").Scan(&goals)
		for _, g := range goals {
			goalMap[g.UserID] = g
		}
		var leads []leadAgg
		s.DB.Model(&models.Lead{}).
			Select("manager_id, COALESCE(SUM(CASE WHEN status='sale' THEN budget ELSE 0 END),0) as sales_fact, SUM(CASE WHEN status='sale' THEN 1 ELSE 0 END) as sales_cnt, COUNT(*) as leads_fact").
			Where("manager_id IN ? AND created_at BETWEEN ? AND ?", managerIDs, start, end).
			Group("manager_id").Scan(&leads)
		for _, l := range leads {
			leadMap[l.ManagerID] = l
		}
		var acts []actAgg
		s.DB.Model(&models.LeadActivity{}).
			Select("user_id, type, COUNT(*) as cnt").
			Where("user_id IN ? AND created_at BETWEEN ? AND ? AND type IN ?", managerIDs, start, end, []string{"call", "meeting"}).
			Group("user_id, type").Scan(&acts)
		for _, a := range acts {
			if _, ok := actMap[a.UserID]; !ok {
				actMap[a.UserID] = map[string]int64{}
			}
			actMap[a.UserID][a.Type] = a.Cnt
		}
		var funnels []funnelAgg
		s.DB.Model(&models.Lead{}).
			Select("manager_id, status, COUNT(*) as cnt").
			Where("manager_id IN ?", managerIDs).
			Group("manager_id, status").Scan(&funnels)
		for _, f := range funnels {
			if _, ok := funnelMap[f.ManagerID]; !ok {
				funnelMap[f.ManagerID] = map[string]int64{}
			}
			funnelMap[f.ManagerID][f.Status] = f.Cnt
		}
	}

	var results []map[string]interface{}

	for _, mgr := range managers {
		if mgr.SalonID == nil {
			continue
		}

		// --- KPI (из батч-мап) ---
		ga := goalMap[mgr.ID]
		totalSalesPlan := ga.TotalSalesPlan
		totalLeadsPlan := ga.TotalLeadsPlan
		totalCallsPlan := ga.TotalCallsPlan
		totalMeetingsPlan := ga.TotalMeetPlan

		la := leadMap[mgr.ID]
		salesFact := la.SalesFact
		leadsFact := la.LeadsFact

		callsFact := actMap[mgr.ID]["call"]
		meetingsFact := actMap[mgr.ID]["meeting"]

		// --- Воронка (из батч-мап) ---
		fm := funnelMap[mgr.ID]
		lWait := fm["wait"]
		lSale := fm["sale"]
		lWork := fm["contact"] + fm["meeting"]

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

// GetDashboardMain - получение данных для главной вкладки дашборда менеджера салона
func (s *KPIService) GetDashboardMain(ctx context.Context, userID uuid.UUID, dateStr string) (*models.DashboardMainResponse, error) {
	// Парсим дату или берем текущую
	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}

	resp := &models.DashboardMainResponse{}

	// Получаем салон пользователя
	var user models.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	if user.SalonID == nil {
		return nil, gorm.ErrRecordNotFound
	}
	salonID := *user.SalonID

	// ====================
	// 1. ПЛАН И ФАКТ НА МЕСЯЦ
	// ====================
	firstOfMonth := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	// План на месяц (сумма goals по пользователю)
	// Используем assignee_id для поиска плана, так как в goals нет salon_id
	var monthPlan float64
	s.DB.Model(&models.Goal{}).
		Where("assignee_id = ? AND target_date BETWEEN ? AND ?", userID, firstOfMonth, lastOfMonth).
		Select("COALESCE(SUM(sales_plan), 0)").Scan(&monthPlan)
	// Также учитываем периоды month/week/year
	if monthPlan == 0 {
		s.DB.Model(&models.Goal{}).
			Where("assignee_id = ? AND period IN ('month', 'year') AND target_date >= ?", userID, firstOfMonth).
			Select("COALESCE(SUM(sales_plan), 0)").Scan(&monthPlan)
	}

	// Факт на текущий момент (сумма продаж из leads со статусом sale)
	var monthFact, orderFact float64
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND status = ? AND created_at BETWEEN ? AND ?", salonID, "sale", firstOfMonth, targetDate).
		Select("COALESCE(SUM(budget), 0)").Scan(&monthFact)

	// Факт из заказов — более точный, не суммируем с leads чтобы избежать двойного учета одного договора
	s.DB.Model(&models.Order{}).
		Where("salon_id = ? AND status IN ? AND created_at BETWEEN ? AND ?", salonID, []string{"paid", "contract"}, firstOfMonth, targetDate).
		Select("COALESCE(SUM(total_price), 0)").Scan(&orderFact)

	// Берём максимум из двух источников (заказ точнее, но может отсутствовать)
	if orderFact > monthFact {
		monthFact = orderFact
	}

	resp.Plan = monthPlan
	resp.Fact = monthFact

	// Calculate percent - ensure it's always set when we have data
	if monthPlan > 0 {
		percent := (monthFact / monthPlan) * 100
		resp.PlanPercent = int(percent)
		if resp.PlanPercent > 100 {
			resp.PlanPercent = 100
		}
	} else {
		// If there's fact but no plan, show 100% or use default
		if monthFact > 0 {
			resp.PlanPercent = 100
		}
	}

	// ====================
	// 2. ДИНАМИКА (сравнение с прошлым днем и неделей)
	// ====================
	// Сегодня vs вчера/неделю — корректно сравниваем день к дню
	todayStart := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
	todayEnd := todayStart.AddDate(0, 0, 1)
	var todaySales float64
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND status = ? AND created_at >= ? AND created_at < ?", salonID, "sale", todayStart, todayEnd).
		Select("COALESCE(SUM(budget), 0)").Scan(&todaySales)

	// Вчера
	yesterday := targetDate.AddDate(0, 0, -1)
	yesterdayStart := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
	yesterdayEnd := yesterdayStart.AddDate(0, 0, 1)
	var yesterdaySales float64
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND status = ? AND created_at >= ? AND created_at < ?", salonID, "sale", yesterdayStart, yesterdayEnd).
		Select("COALESCE(SUM(budget), 0)").Scan(&yesterdaySales)

	// Прошлая неделя (7 дней назад)
	weekAgo := targetDate.AddDate(0, 0, -7)
	weekAgoStart := time.Date(weekAgo.Year(), weekAgo.Month(), weekAgo.Day(), 0, 0, 0, 0, weekAgo.Location())
	weekAgoEnd := weekAgoStart.AddDate(0, 0, 1)
	var weekAgoSales float64
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND status = ? AND created_at >= ? AND created_at < ?", salonID, "sale", weekAgoStart, weekAgoEnd).
		Select("COALESCE(SUM(budget), 0)").Scan(&weekAgoSales)

	if yesterdaySales > 0 {
		resp.DynamicDay = ((todaySales - yesterdaySales) / yesterdaySales) * 100
	}
	if weekAgoSales > 0 {
		resp.DynamicWeek = ((todaySales - weekAgoSales) / weekAgoSales) * 100
	}

	// ====================
	// 3. ПРОГНОЗ
	// ====================
	// Среднедневные продажи
	daysInMonth := targetDate.Day()
	avgDaily := monthFact / float64(daysInMonth)
	daysLeft := lastOfMonth.Day() - targetDate.Day() + 1
	forecast := monthFact + (avgDaily * float64(daysLeft))
	if monthPlan > 0 {
		resp.Forecast = int((forecast / monthPlan) * 100)
		if resp.Forecast > 100 {
			resp.Forecast = 100
		}
	}

	// ====================
	// 4. СРЕДНИЙ ЧЕК
	// ====================
	var dealsCount int64
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND status = ? AND created_at BETWEEN ? AND ?", salonID, "sale", firstOfMonth, targetDate).
		Count(&dealsCount)
	if dealsCount > 0 {
		resp.AvgCheck = monthFact / float64(dealsCount)
	}

	// ====================
	// 5. МАРЖИНАЛЬНОСТЬ
	// ====================
	var totalMargin float64
	s.DB.Model(&models.Contract{}).
		Where("salon_id = ? AND created_at BETWEEN ? AND ?", salonID, firstOfMonth, targetDate).
		Select("COALESCE(AVG(margin_percent), 0)").Scan(&totalMargin)
	resp.MarginPercent = totalMargin

	// ====================
	// 6. СУММА ПРЕДОПЛАТ
	// ====================
	var prepaymentsSum float64
	s.DB.Model(&models.Contract{}).
		Where("salon_id = ? AND created_at BETWEEN ? AND ?", salonID, firstOfMonth, targetDate).
		Select("COALESCE(SUM(prepaid_amount), 0)").Scan(&prepaymentsSum)
	resp.PrepaymentsSum = prepaymentsSum

	// ====================
	// 7. СВЕТОФОР
	// ====================
	// Трафик (количество лидов за сегодня vs норма)
	today := targetDate.Format("2006-01-02")
	var todayLeads int64
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND created_at >= ? AND created_at < ?", salonID, todayStart, todayEnd).
		Count(&todayLeads)

	// Норма - 10 лидов в день (можно сделать настраиваемой)
	normTraffic := 10
	if todayLeads >= int64(normTraffic) {
		resp.TrafficStatus = "green"
	} else if todayLeads >= int64(normTraffic/2) {
		resp.TrafficStatus = "yellow"
	} else {
		resp.TrafficStatus = "red"
	}

	// Конверсия в замеры (лиды со статусом meeting / все лиды)
	var totalLeads, meetingLeads int64
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND created_at BETWEEN ? AND ?", salonID, firstOfMonth, targetDate).
		Count(&totalLeads)
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND status = ? AND created_at BETWEEN ? AND ?", salonID, "meeting", firstOfMonth, targetDate).
		Count(&meetingLeads)

	if totalLeads > 0 {
		convMeasure := int((meetingLeads * 100) / totalLeads)
		if convMeasure >= 30 {
			resp.ConversionMeasureStatus = "green"
		} else if convMeasure >= 15 {
			resp.ConversionMeasureStatus = "yellow"
		} else {
			resp.ConversionMeasureStatus = "red"
		}
	} else {
		resp.ConversionMeasureStatus = "red"
	}

	// Конверсия в договоры (продажи / лиды)
	var saleLeads int64
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND status = ? AND created_at BETWEEN ? AND ?", salonID, "sale", firstOfMonth, targetDate).
		Count(&saleLeads)

	if totalLeads > 0 {
		convContract := int((saleLeads * 100) / totalLeads)
		if convContract >= 20 {
			resp.ConversionContractStatus = "green"
		} else if convContract >= 10 {
			resp.ConversionContractStatus = "yellow"
		} else {
			resp.ConversionContractStatus = "red"
		}
	} else {
		resp.ConversionContractStatus = "red"
	}

	// ====================
	// 8. ОЖИДАЕМЫЕ ОПЛАТЫ НА СЕГОДНЯ
	// ====================
	var contracts []models.Contract
	s.DB.Where("salon_id = ? AND payment_status IN ? AND payment_date >= ? AND payment_date < ?", salonID, []string{"awaiting_payment", "payment_due"}, todayStart, todayEnd).
		Find(&contracts)

	for _, c := range contracts {
		resp.PendingPayments = append(resp.PendingPayments, models.PendingPayment{
			ContractID:  c.ID,
			ClientName:  c.ClientName,
			Amount:      c.RemainAmount,
			Status:      c.PaymentStatus,
			PaymentDate: today,
		})
	}

	return resp, nil
}

// GetDashboardFunnel - получение данных для воронки продаж
func (s *KPIService) GetDashboardFunnel(ctx context.Context, userID uuid.UUID, dateStr string) (*models.DashboardFunnelResponse, error) {
	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}

	resp := &models.DashboardFunnelResponse{
		Stages:     []models.FunnelStage{},
		HotDeals:   []models.HotDeal{},
		FreshLeads: []models.FreshLead{},
	}

	// Получаем салон пользователя
	var user models.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	if user.SalonID == nil {
		return nil, gorm.ErrRecordNotFound
	}
	salonID := *user.SalonID

	// Этапы воронки: Трафик → Консультация → Замер → КП → Договор → Оплата
	// Маппим статусы лидов на этапы воронки
	statusCounts := make(map[string]int64)
	statusSums := make(map[string]float64)

	rows, err := s.DB.Model(&models.Lead{}).
		Where("salon_id = ?", salonID).
		Select("status, COUNT(*), COALESCE(SUM(budget), 0)").
		Group("status").
		Rows()
	if err == nil {
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var status string
			var count int64
			var sum float64
_ = rows.Scan(&status, &count, &sum)
			statusCounts[status] = count
			statusSums[status] = sum
		}
	}

	// Трафик = все лиды
	trafficCount := int64(0)
	for _, c := range statusCounts {
		trafficCount += c
	}

	// Конверсия = (следующий этап / предыдущий) * 100
	prevCount := trafficCount
	addStage := func(stage, label string, count int64, sum float64) {
		conv := 0
		if prevCount > 0 {
			conv = int((count * 100) / prevCount)
		}
		resp.Stages = append(resp.Stages, models.FunnelStage{
			Stage:      stage,
			Label:      label,
			Count:      int(count),
			Conversion: conv,
			Sum:        sum,
		})
		prevCount = count
	}

	// Трафик → Консультация → Замер → КП → Договор → Оплата
	addStage("traffic", "Трафик", trafficCount, statusSums["new"])
	addStage("consultation", "Консультация", statusCounts["contact"], statusSums["contact"])
	addStage("measurement", "Замер", statusCounts["meeting"], statusSums["meeting"])
	addStage("kp", "КП", statusCounts["wait"], statusSums["wait"])
	addStage("contract", "Договор", statusCounts["sale"], statusSums["sale"])
	addStage("payment", "Оплата", statusCounts["paid"], statusSums["paid"])

	// ===== ГОРЯЧИЕ СДЕЛКИ (КП > N дней) =====
	hotDays := 5
	hotDaysThreshold := targetDate.AddDate(0, 0, -hotDays)

	var hotLeads []models.Lead
	s.DB.Where("salon_id = ? AND status = ? AND updated_at < ?", salonID, "wait", hotDaysThreshold).
		Find(&hotLeads)

	for _, lead := range hotLeads {
		daysStalled := int(time.Since(lead.UpdatedAt).Hours() / 24)
		resp.HotDeals = append(resp.HotDeals, models.HotDeal{
			ID:          lead.ID,
			ClientName:  lead.FullName,
			Phone:       lead.Phone,
			Amount:      lead.Budget,
			CreatedAt:   lead.UpdatedAt.Format("2006-01-02"),
			DaysStalled: daysStalled,
			ManagerID:   lead.ManagerID,
			ManagerName: "",
		})
	}

	// ===== СВЕЖИЕ ЛИДЫ (за сегодня) =====
	todayStart2 := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
	todayEnd2 := todayStart2.AddDate(0, 0, 1)
	var todayLeads []models.Lead
	s.DB.Where("salon_id = ? AND created_at >= ? AND created_at < ?", salonID, todayStart2, todayEnd2).
		Order("created_at DESC").
		Find(&todayLeads)

	for _, lead := range todayLeads {
		status := "unassigned"
		if lead.ManagerID != uuid.Nil {
			status = "in_progress"
		}
		resp.FreshLeads = append(resp.FreshLeads, models.FreshLead{
			ID:          lead.ID,
			Source:      "Сайт",
			ClientName:  lead.FullName,
			Phone:       lead.Phone,
			CreatedAt:   lead.CreatedAt.Format("2006-01-02 15:04"),
			Status:      status,
			AssignedTo:  &lead.ManagerID,
			ManagerName: "",
		})
	}

	return resp, nil
}

// GetDashboardTeam - получение данных команды (рейтинг продавцов)
func (s *KPIService) GetDashboardTeam(ctx context.Context, userID uuid.UUID, period, dateStr string) (*models.DashboardTeamResponse, error) {
	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}

	var startDate, endDate time.Time
	switch period {
	case "week":
		startDate = targetDate.AddDate(0, 0, -7)
		endDate = targetDate
	case "month":
		startDate = time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())
		endDate = targetDate
	case "quarter":
		quarter := (int(targetDate.Month()) - 1) / 3
		startDate = time.Date(targetDate.Year(), time.Month(quarter*3+1), 1, 0, 0, 0, 0, targetDate.Location())
		endDate = targetDate
	default:
		startDate = time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())
		endDate = targetDate
	}

	resp := &models.DashboardTeamResponse{
		Period:        period,
		TotalRevenue:  0,
		AvgRevenue:    0,
		AvgConversion: 0,
		AvgCheck:      0,
		SalesReps:     []models.SalesRepMetrics{},
	}

	var user models.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	if user.SalonID == nil {
		return nil, gorm.ErrRecordNotFound
	}
	salonID := *user.SalonID

	var managers []models.User
	// Ищем продавцов (sales_rep) которые управляются текущим пользователем или в том же салоне
	s.DB.Where("managed_by = ? OR (salon_id = ? AND role = 'sales_rep')", userID, salonID).Limit(100).Find(&managers)

	var totalRevenue, totalDeals, totalConversion, totalAvgCheck float64
	var dealsCount, conversionCount int

	// Батчим N+1: 1 запрос на выручку+сделки и 1 на лиды по всем managers (вместо 2*N)
	type saleAgg struct {
		ManagerID uuid.UUID
		Revenue   float64
		Deals     int64
	}
	type leadAgg struct {
		ManagerID uuid.UUID
		Count     int64
	}
	var managerIDs []uuid.UUID
	for _, m := range managers {
		managerIDs = append(managerIDs, m.ID)
	}
	saleMap := map[uuid.UUID]saleAgg{}
	leadMap := map[uuid.UUID]int64{}
	if len(managerIDs) > 0 {
		var sales []saleAgg
		s.DB.Model(&models.Lead{}).
			Select("manager_id, COALESCE(SUM(budget),0) as revenue, COUNT(*) as deals").
			Where("manager_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", managerIDs, []string{"sale", "paid"}, startDate, endDate).
			Group("manager_id").
			Scan(&sales)
		for _, r := range sales {
			saleMap[r.ManagerID] = r
		}
		var leads []leadAgg
		s.DB.Model(&models.Lead{}).
			Select("manager_id, COUNT(*) as count").
			Where("manager_id IN ? AND created_at BETWEEN ? AND ?", managerIDs, startDate, endDate).
			Group("manager_id").
			Scan(&leads)
		for _, r := range leads {
			leadMap[r.ManagerID] = r.Count
		}
	}

	for _, mgr := range managers {
		agg := saleMap[mgr.ID]
		revenue := agg.Revenue
		deals := agg.Deals
		leadsCount := leadMap[mgr.ID]

		conversion := 0.0
		if leadsCount > 0 {
			conversion = (float64(deals) / float64(leadsCount)) * 100
		}

		avgCheck := 0.0
		if deals > 0 {
			avgCheck = revenue / float64(deals)
		}

		extrasSum := revenue * s.getSettingFloat("extras_rate", 0.1)
		discountPercent := s.getSettingFloat("discount_default_percent", 5.0)

		totalRevenue += revenue
		totalDeals += float64(deals)
		if conversion > 0 {
			totalConversion += conversion
			conversionCount++
		}
		if avgCheck > 0 {
			totalAvgCheck += avgCheck
			dealsCount++
		}

		resp.SalesReps = append(resp.SalesReps, models.SalesRepMetrics{
			UserID:              mgr.ID,
			FirstName:           mgr.FirstName,
			LastName:            mgr.LastName,
			Role:                string(mgr.Role),
			Revenue:             revenue,
			DealsCount:          int(deals),
			Conversion:          conversion,
			AvgCheck:            avgCheck,
			DiscountPercent:     discountPercent,
			ExtrasSum:           extrasSum,
			RevenueDeviation:    0,
			DealsDeviation:      0,
			ConversionDeviation: 0,
			AvgCheckDeviation:   0,
		})
	}

	if len(managers) > 0 {
		resp.TotalRevenue = totalRevenue
		resp.AvgRevenue = totalRevenue / float64(len(managers))
		if conversionCount > 0 {
			resp.AvgConversion = totalConversion / float64(conversionCount)
		}
		if dealsCount > 0 {
			resp.AvgCheck = totalAvgCheck / float64(dealsCount)
		}
	}

	for i := range resp.SalesReps {
		if resp.AvgRevenue > 0 {
			resp.SalesReps[i].RevenueDeviation = ((resp.SalesReps[i].Revenue - resp.AvgRevenue) / resp.AvgRevenue) * 100
		}
		avgDeals := totalDeals / float64(len(managers))
		if avgDeals > 0 {
			resp.SalesReps[i].DealsDeviation = ((float64(resp.SalesReps[i].DealsCount) - avgDeals) / avgDeals) * 100
		}
		if resp.AvgConversion > 0 {
			resp.SalesReps[i].ConversionDeviation = ((resp.SalesReps[i].Conversion - resp.AvgConversion) / resp.AvgConversion) * 100
		}
		if resp.AvgCheck > 0 {
			resp.SalesReps[i].AvgCheckDeviation = ((resp.SalesReps[i].AvgCheck - resp.AvgCheck) / resp.AvgCheck) * 100
		}
	}

	for i := 0; i < len(resp.SalesReps)-1; i++ {
		for j := i + 1; j < len(resp.SalesReps); j++ {
			if resp.SalesReps[i].Revenue < resp.SalesReps[j].Revenue {
				resp.SalesReps[i], resp.SalesReps[j] = resp.SalesReps[j], resp.SalesReps[i]
			}
		}
	}

	return resp, nil
}

// GetSalesRepHistory - история продавца за N месяцев
func (s *KPIService) GetSalesRepHistory(ctx context.Context, managerID uuid.UUID, months int) ([]models.SalesRepHistory, error) {
	resp := []models.SalesRepHistory{}

	now := time.Now()
	for i := months - 1; i >= 0; i-- {
		monthStart := time.Date(now.Year(), now.Month()-time.Month(i), 1, 0, 0, 0, 0, now.Location())
		monthEnd := monthStart.AddDate(0, 1, 0)

		var revenue float64
		s.DB.Model(&models.Lead{}).
			Where("manager_id = ? AND status IN ? AND created_at BETWEEN ? AND ?", managerID, []string{"sale", "paid"}, monthStart, monthEnd).
			Select("COALESCE(SUM(budget), 0)").Scan(&revenue)

		var deals int64
		s.DB.Model(&models.Lead{}).
			Where("manager_id = ? AND status IN ? AND created_at BETWEEN ? AND ?", managerID, []string{"sale", "paid"}, monthStart, monthEnd).
			Count(&deals)

		avgCheck := 0.0
		if deals > 0 {
			avgCheck = revenue / float64(deals)
		}

		resp = append(resp, models.SalesRepHistory{
			Month:    monthStart.Format("2006-01"),
			Revenue:  revenue,
			Deals:    int(deals),
			AvgCheck: avgCheck,
		})
	}

	return resp, nil
}

// GetDashboardProducts - получение данных о товарах
func (s *KPIService) GetDashboardProducts(ctx context.Context, userID uuid.UUID, dateStr string) (*models.DashboardProductsResponse, error) {
	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}

	resp := &models.DashboardProductsResponse{
		TopProducts:      []models.TopProduct{},
		StockItems:       []models.StockItem{},
		LostSales:        []models.LostSale{},
		CategoryTurnover: []models.CategoryTurnover{},
		TotalRevenue:     0,
	}

	var user models.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	if user.SalonID == nil {
		return nil, gorm.ErrRecordNotFound
	}
	salonID := *user.SalonID

	// Период - текущий месяц
	firstOfMonth := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())
	endOfMonth := targetDate
	period := firstOfMonth.Format("2006-01")

	// Общая выручка салона
	var totalRevenue float64
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND status IN ? AND created_at BETWEEN ? AND ?", salonID, []string{"sale", "paid"}, firstOfMonth, endOfMonth).
		Select("COALESCE(SUM(budget), 0)").Scan(&totalRevenue)
	resp.TotalRevenue = totalRevenue

	// === КАТАЛОГ ПРОДУКТОВ САЛОНА ===
	type productRow struct {
		ID           string
		Name         string
		Collection   string
		Category     string
		Price        float64
		CostPrice    float64
		ShowroomQty  int
		WarehouseQty int
		TurnoverDays int
	}
	var catalog []productRow
	s.DB.Table("products").
		Where("salon_id = ?", salonID).
		Order("name").
		Scan(&catalog)

	catalogByName := make(map[string]productRow, len(catalog))
	for _, p := range catalog {
		catalogByName[p.Name] = p
	}

	// === ТОП-10 ТОВАРОВ (по interest_product) ===
	// Группируем по interest_product (названию товара)
	type productStats struct {
		Name     string
		Revenue  float64
		Quantity int
	}

	var productStatsList []productStats
	rows, err := s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND status IN ? AND created_at BETWEEN ? AND ?", salonID, []string{"sale", "paid"}, firstOfMonth, endOfMonth).
		Select("interest_product, COALESCE(SUM(budget), 0) as revenue, COUNT(*) as quantity").
		Group("interest_product").
		Order("revenue DESC").
		Limit(10).
		Rows()
	if err == nil {
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var ps productStats
_ = rows.Scan(&ps.Name, &ps.Revenue, &ps.Quantity)
			if ps.Name != "" {
				productStatsList = append(productStatsList, ps)
			}
		}
	}

	for _, ps := range productStatsList {
		share := 0.0
		if totalRevenue > 0 {
			share = (ps.Revenue / totalRevenue) * 100
		}
		margin := 0.0
		collection := ""
		category := ""
		if prod, ok := catalogByName[ps.Name]; ok {
			collection = prod.Collection
			category = prod.Category
			if prod.Price > 0 {
				margin = ((prod.Price - prod.CostPrice) / prod.Price) * 100
			}
		} else {
			category = "Прочее"
		}
		resp.TopProducts = append(resp.TopProducts, models.TopProduct{
			ID:           ps.Name,
			Name:         ps.Name,
			Collection:   collection,
			Category:     category,
			Revenue:      ps.Revenue,
			Quantity:     ps.Quantity,
			SharePercent: share,
			Margin:       margin,
		})
	}

	// === ОСТАТКИ (из каталога products) ===
	for _, p := range catalog {
		totalQty := p.ShowroomQty + p.WarehouseQty
		totalCost := float64(totalQty) * p.CostPrice
		days := p.TurnoverDays
		if days == 0 {
			days = 90 // Неликвид по умолчанию без данных об оборачиваемости
		}
		resp.StockItems = append(resp.StockItems, models.StockItem{
			ID:           p.ID,
			Name:         p.Name,
			Category:     p.Category,
			ShowroomQty:  p.ShowroomQty,
			WarehouseQty: p.WarehouseQty,
			TotalCost:    totalCost,
			TurnoverDays: days,
		})
	}

	// === УПУЩЕННЫЕ ПРОДАЖИ (из lost_sales) ===
	type lostRow struct {
		Reason        string
		RequestsCount int
		LostRevenue   float64
	}
	var lostRows []lostRow
	s.DB.Table("lost_sales").
		Where("salon_id = ? AND period = ?", salonID, period).
		Order("lost_revenue DESC").
		Scan(&lostRows)
	for _, lr := range lostRows {
		resp.LostSales = append(resp.LostSales, models.LostSale{
			Reason:        lr.Reason,
			RequestsCount: lr.RequestsCount,
			LostRevenue:   lr.LostRevenue,
		})
	}

	// === ОБОРАЧИВАЕМОСТЬ ПО КАТЕГОРИЯМ (из category_turnover) ===
	type catRow struct {
		Category string
		AvgDays  int
	}
	var catRows []catRow
	s.DB.Table("category_turnover").
		Where("salon_id = ? AND period = ?", salonID, period).
		Order("category").
		Scan(&catRows)
	for _, cr := range catRows {
		resp.CategoryTurnover = append(resp.CategoryTurnover, models.CategoryTurnover{
			Category:     cr.Category,
			AvgDays:      cr.AvgDays,
			IsSlowMoving: cr.AvgDays > 90,
		})
	}

	return resp, nil
}

// GetManagerTargets - получить директивы от дилера
func (s *KPIService) GetManagerTargets(ctx context.Context, userID uuid.UUID, dateStr string) (*models.ManagerTargetsResponse, error) {
	resp := &models.ManagerTargetsResponse{
		HasTargets:           false,
		TargetConversion:     0,
		CurrentConversion:    0,
		TargetExtrasPercent:  0,
		CurrentExtrasPercent: 0,
		Promotions:           []models.Promotion{},
		BonusForecast:        0,
		MaxBonus:             0,
		WarningLevel:         "none",
	}

	var user models.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	if user.SalonID == nil {
		return nil, gorm.ErrRecordNotFound
	}
	salonID := *user.SalonID

	// === План продаж ===
	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}
	firstOfMonth := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())

	// План на месяц - ищем по assignee_id в таблице goals
	var planAmount float64
	s.DB.Model(&models.Goal{}).
		Where("assignee_id = ? AND target_date BETWEEN ? AND ?", userID, firstOfMonth, targetDate).
		Select("COALESCE(SUM(sales_plan), 0)").Scan(&planAmount)
	// Также учитываем периоды month/week/year
	if planAmount == 0 {
		s.DB.Model(&models.Goal{}).
			Where("assignee_id = ? AND period IN ('month', 'year')", userID).
			Select("COALESCE(SUM(sales_plan), 0)").Scan(&planAmount)
	}

	// Если план не утверждён (не задан)
	if planAmount == 0 {
		resp.HasTargets = false
		resp.Plan = nil
		return resp, nil
	}

	resp.HasTargets = true

	// Текущее выполнение - используем salonID из user
	var currentAmount float64
	var currentUser models.User
	if err := s.DB.First(&currentUser, userID).Error; err == nil && currentUser.SalonID != nil {
		s.DB.Model(&models.Lead{}).
			Where("salon_id = ? AND status IN ? AND created_at BETWEEN ? AND ?", currentUser.SalonID, []string{"sale", "paid"}, firstOfMonth, targetDate).
			Select("COALESCE(SUM(budget), 0)").Scan(&currentAmount)
	}

	percent := 0
	if planAmount > 0 {
		percent = int((currentAmount / planAmount) * 100)
		if percent > 100 {
			percent = 100
		}
	}

	// === Распределение выручки по категориям (из каталога) ===
	// Сопоставляем проданные товары (interest_product лидов) с категорией
	// товара в каталоге; без каталога категория неизвестна - "Прочее".
	byCategory := map[string]float64{}
	if currentUser.SalonID != nil {
		type saleCategory struct {
			Category string
			Revenue  float64
		}
		var catSales []saleCategory
		s.DB.Table("leads").
			Joins("LEFT JOIN products ON products.salon_id = leads.salon_id AND products.name = leads.interest_product").
			Where("leads.salon_id = ? AND leads.status IN ? AND leads.created_at BETWEEN ? AND ?", currentUser.SalonID, []string{"sale", "paid"}, firstOfMonth, targetDate).
			Select("COALESCE(NULLIF(products.category, ''), 'Прочее') AS category, COALESCE(SUM(leads.budget), 0) AS revenue").
			Group("COALESCE(NULLIF(products.category, ''), 'Прочее')").
			Scan(&catSales)
		for _, cs := range catSales {
			byCategory[cs.Category] = cs.Revenue
		}
	}

	resp.Plan = &models.TargetPlan{
		TotalAmount:   planAmount,
		ByCategory:    byCategory,
		CurrentAmount: currentAmount,
		Percent:       percent,
	}

	// === Бенчмарки (из system_settings, с дефолтами) ===
	getSetting := func(key string, def float64) float64 {
		var v string
		if err := s.DB.Table("system_settings").Where("key = ?", key).Select("value").Scan(&v).Error; err == nil && v != "" {
			if parsed, perr := strconv.ParseFloat(v, 64); perr == nil {
				return parsed
			}
		}
		return def
	}
	resp.TargetConversion = getSetting("target_conversion", 30.0)
	resp.TargetExtrasPercent = getSetting("target_extras_percent", 15.0)

	// Текущая конверсия
	var totalLeads, saleLeads int64
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND created_at BETWEEN ? AND ?", salonID, firstOfMonth, targetDate).
		Count(&totalLeads)
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND status IN ? AND created_at BETWEEN ? AND ?", salonID, []string{"sale", "paid"}, firstOfMonth, targetDate).
		Count(&saleLeads)

	if totalLeads > 0 {
		resp.CurrentConversion = (float64(saleLeads) / float64(totalLeads)) * 100
	}

	// Текущая доля допов (категории НЕ "Мебель")
	if currentAmount > 0 {
		furniture := byCategory["Мебель"]
		extras := currentAmount - furniture
		resp.CurrentExtrasPercent = (extras / currentAmount) * 100
	}

	// === Акции (из таблицы promotions) ===
	var promoRows []struct {
		ID          string
		Name        string
		Condition   string
		DiscountMin int
		DiscountMax int
		EndDate     *time.Time
	}
	promoErr := s.DB.Table("promotions").
		Where("salon_id = ? AND is_active = TRUE", salonID).
		Order("end_date").
		Scan(&promoRows).Error
	if promoErr != nil {
		promoRows = nil
	}
	now := time.Now()
	for _, pr := range promoRows {
		expiring := false
		endDate := ""
		if pr.EndDate != nil {
			endDate = pr.EndDate.Format("2006-01-02")
			expiring = pr.EndDate.Before(now.AddDate(0, 0, 7))
		}
		resp.Promotions = append(resp.Promotions, models.Promotion{
			ID:          pr.ID,
			Name:        pr.Name,
			Condition:   pr.Condition,
			DiscountMin: pr.DiscountMin,
			DiscountMax: pr.DiscountMax,
			EndDate:     endDate,
			IsExpiring:  expiring,
		})
	}

	// === Прогноз премии ===
	maxBonus := getSetting("max_bonus", 50000.0)
	if percent >= 100 {
		resp.BonusForecast = maxBonus
	} else if percent >= 80 {
		resp.BonusForecast = maxBonus * 0.8
	} else if percent >= 50 {
		resp.BonusForecast = maxBonus * 0.5
	} else {
		resp.BonusForecast = maxBonus * 0.2
	}
	resp.MaxBonus = maxBonus

	// Уровень предупреждения — сначала <30 red, иначе <50 yellow
	bonusPercent := (resp.BonusForecast / maxBonus) * 100
	if bonusPercent < 30 {
		resp.WarningLevel = "red"
	} else if bonusPercent < 50 {
		resp.WarningLevel = "yellow"
	} else {
		resp.WarningLevel = "none"
	}

	return resp, nil
}

// GetDealerSummary - сводка для дилера (все салоны)
func (s *KPIService) GetDealerSummary(ctx context.Context, userID uuid.UUID, dateStr string) (*models.DealerSummaryResponse, error) {
	resp := &models.DealerSummaryResponse{}

	// Сначала находим пользователя дилера
	var user models.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}

	// Находим все салоны дилера через tenant_id (dealer_id в таблице salons = user.tenant_id)
	var dealerTenantID *uuid.UUID
	if user.TenantID != nil {
		dealerTenantID = user.TenantID
	}

	var salons []models.Salon
	if dealerTenantID != nil {
		if err := s.DB.Where("dealer_id = ?", *dealerTenantID).Limit(100).Find(&salons).Error; err != nil {
			return nil, err
		}
	}

	// Собираем все salonID
	var salonIDs []uuid.UUID
	for _, salon := range salons {
		salonIDs = append(salonIDs, salon.ID)
	}

	if len(salonIDs) == 0 {
		return resp, nil
	}

	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}
	firstOfMonth := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())

	// Выручка из заказов (orders)
	var orderRevenue float64
	s.DB.Model(&models.Order{}).
		Where("salon_id IN ? AND status = ? AND created_at BETWEEN ? AND ?", salonIDs, "paid", firstOfMonth, targetDate).
		Select("COALESCE(SUM(total_price), 0)").Scan(&orderRevenue)

	// Выручка из лидов - учитываем ВСЕ закрытые статусы (paid, contract)
	var leadRevenue float64
	s.DB.Model(&models.Lead{}).
		Where("salon_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", salonIDs, []string{"paid", "contract"}, firstOfMonth, targetDate).
		Select("COALESCE(SUM(budget), 0)").Scan(&leadRevenue)

	// Общая выручка = заказы + лиды (или максимум из них)
	totalRevenue := orderRevenue
	if leadRevenue > totalRevenue {
		totalRevenue = leadRevenue
	}

	// Общий план через goals (для дилеров - ищем по assignee_id = userID)
	var totalPlan float64
	s.DB.Model(&models.Goal{}).
		Where("assignee_id = ? AND period = 'monthly'", userID).
		Select("COALESCE(SUM(sales_plan), 0)").Scan(&totalPlan)

	// Процент выполнения
	planPercent := 0
	if totalPlan > 0 {
		planPercent = int((totalRevenue / totalPlan) * 100)
		if planPercent > 100 {
			planPercent = 100
		}
	}

	// Чистая прибыль — из настроек (fallback 20% для совместимости, теперь конфигурируется)
	netProfit := totalRevenue * s.getSettingFloat("net_profit_rate", 0.2)
	marginProfit := totalRevenue * s.getSettingFloat("gross_margin_rate", 0.35)

	// Алёрты - считаем через notifications для дилера
	var alertsCount int64
	s.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&alertsCount)

	resp.NetProfit = netProfit
	resp.GrossRevenue = totalRevenue
	resp.PlanCompletionPercent = planPercent
	resp.MarginProfit = marginProfit
	resp.ActiveAlerts = int(alertsCount)

	return resp, nil
}

// GetDealerFinance - финансы для дилера
func (s *KPIService) GetDealerFinance(ctx context.Context, userID uuid.UUID, dateStr string) (*models.DealerFinanceResponse, error) {
	resp := &models.DealerFinanceResponse{}

	// Находим все салоны дилера через таблицу salons по dealer_id
	var salons []models.Salon
	if err := s.DB.Where("dealer_id = ?", userID).Limit(100).Find(&salons).Error; err != nil {
		return nil, err
	}

	var salonIDs []uuid.UUID
	for _, salon := range salons {
		salonIDs = append(salonIDs, salon.ID)
	}

	if len(salonIDs) == 0 {
		return resp, nil
	}

	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}
	firstOfMonth := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())

	// Выручка
	s.DB.Model(&models.Lead{}).
		Where("salon_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", salonIDs, []string{"sale", "paid"}, firstOfMonth, targetDate).
		Select("COALESCE(SUM(budget), 0)").Scan(&resp.Revenue)

	// COGS — из настроек (fallback 65%)
	resp.COGS = resp.Revenue * s.getSettingFloat("cogs_rate", 0.65)

	// Расходы из таблицы dealer_expenses
	type Expense struct {
		Category string
		Amount   float64
	}

	// Период = текущий месяц в формате YYYY-MM (таблица dealer_expenses)
	period := targetDate.Format("2006-01")

	var expenses []Expense
	// Читаем расходы из БД - используем raw SQL для безопасности
	rows, err := s.DB.Raw(`
		SELECT category, SUM(amount) as amount 
		FROM dealer_expenses 
		WHERE dealer_id = ? AND period = ?
		GROUP BY category
	`, userID, period).Rows()

	if err == nil {
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var exp Expense
_ = rows.Scan(&exp.Category, &exp.Amount)
			expenses = append(expenses, exp)
		}
		// Маппинг расходов по категориям
		expenseMap := make(map[string]float64)
		for _, e := range expenses {
			expenseMap[e.Category] = e.Amount
		}

		resp.Rent = expenseMap["rent"]
		resp.Utilities = expenseMap["utilities"]
		resp.Payroll = expenseMap["payroll"]
		resp.Taxes = expenseMap["taxes"]
		resp.Logistics = expenseMap["logistics"]
		resp.Marketing = expenseMap["marketing"]
		resp.Defects = expenseMap["defects"]
		resp.OtherExpenses = expenseMap["other"]
	}

	// Если в таблице нет данных - возвращаем нули, не фейковые числа

	// Чистая прибыль
	resp.NetProfit = resp.Revenue - resp.COGS - resp.Rent - resp.Utilities - resp.Payroll - resp.Taxes - resp.Logistics - resp.Marketing - resp.Defects - resp.OtherExpenses - resp.Bonus

	// Прогноз — динамически по длине месяца, с защитой от отрицательной экстраполяции
	daysInMonth := targetDate.Day()
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	daysLeft := lastOfMonth.Day() - daysInMonth + 1
	if resp.NetProfit <= 0 {
		resp.NetProfitForecast = resp.NetProfit
	} else {
		dailyAvg := resp.NetProfit / float64(daysInMonth)
		resp.NetProfitForecast = resp.NetProfit + (dailyAvg * float64(daysLeft))
	}

	// Прошлый месяц: выручка и расходы из БД, а не жёсткий коэффициент
	prevFirstOfMonth := firstOfMonth.AddDate(0, -1, 0)
	prevLastOfMonth := firstOfMonth.AddDate(0, 0, -1)
	var prevRevenue float64
	s.DB.Model(&models.Lead{}).
		Where("salon_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", salonIDs, []string{"sale", "paid"}, prevFirstOfMonth, prevLastOfMonth).
		Select("COALESCE(SUM(budget), 0)").Scan(&prevRevenue)

	prevPeriod := prevFirstOfMonth.Format("2006-01")
	var prevExpensesTotal float64
	s.DB.Raw(`
		SELECT COALESCE(SUM(amount), 0) FROM dealer_expenses 
		WHERE dealer_id = ? AND period = ?
	`, userID, prevPeriod).Scan(&prevExpensesTotal)

	resp.PrevMonthNetProfit = prevRevenue - (prevRevenue * s.getSettingFloat("cogs_rate", 0.65)) - prevExpensesTotal

	// Заполняем expense_breakdown
	expenseItems := []struct {
		Category string
		Amount   float64
	}{
		{"rent", resp.Rent},
		{"utilities", resp.Utilities},
		{"payroll", resp.Payroll},
		{"taxes", resp.Taxes},
		{"logistics", resp.Logistics},
		{"marketing", resp.Marketing},
		{"defects", resp.Defects},
		{"other", resp.OtherExpenses},
	}

	for _, item := range expenseItems {
		percent := 0.0
		if resp.Revenue > 0 {
			percent = (item.Amount / resp.Revenue) * 100
		}
		resp.ExpenseBreakdown = append(resp.ExpenseBreakdown, models.ExpenseBreakdown{
			Category:         item.Category,
			Amount:           item.Amount,
			PercentOfRevenue: percent,
			PrevMonthAmount:  item.Amount * s.getSettingFloat("prev_month_factor", 0.9),
		})
	}

	return resp, nil
}

// GetDealerFunnel - воронка для дилера
func (s *KPIService) GetDealerFunnel(ctx context.Context, userID uuid.UUID, period, dateStr string) (*models.DealerFunnelResponse, error) {
	resp := &models.DealerFunnelResponse{}

	// Находим все салоны дилера через таблицу salons по dealer_id
	var salons []models.Salon
	if err := s.DB.Where("dealer_id = ?", userID).Limit(100).Find(&salons).Error; err != nil {
		return nil, err
	}

	var salonIDs []uuid.UUID
	for _, salon := range salons {
		salonIDs = append(salonIDs, salon.ID)
	}

	if len(salonIDs) == 0 {
		return resp, nil
	}

	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}

	var startDate, endDate time.Time
	quarterStart := func(t time.Time) time.Time {
		m := int(t.Month())
		quarterMonth := time.Month(((m-1)/3)*3 + 1)
		return time.Date(t.Year(), quarterMonth, 1, 0, 0, 0, 0, t.Location())
	}

	switch period {
	case "week":
		startDate = targetDate.AddDate(0, 0, -7)
		endDate = targetDate
	case "quarter":
		startDate = quarterStart(targetDate)
		endDate = targetDate
	case "year":
		startDate = time.Date(targetDate.Year(), 1, 1, 0, 0, 0, 0, targetDate.Location())
		endDate = targetDate
	default: // month
		startDate = time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())
		endDate = targetDate
	}

	// Воронка по стадиям — батчим 5 COUNT → 1 GROUP BY
	var newLeads, contactLeads, meetingLeads, waitLeads, saleLeads int64
	{
		type cntRow struct {
			Status string `gorm:"column:status"`
			Cnt    int64  `gorm:"column:cnt"`
		}
		var rows []cntRow
		s.DB.Model(&models.Lead{}).Select("status, COUNT(*) as cnt").
			Where("salon_id IN ? AND created_at BETWEEN ? AND ?", salonIDs, startDate, endDate).
			Group("status").Scan(&rows)
		m := map[string]int64{}
		var total int64
		for _, r := range rows {
			m[r.Status] = r.Cnt
			total += r.Cnt
		}
		newLeads = total
		contactLeads = m["contact"]
		meetingLeads = m["meeting"]
		waitLeads = m["wait"]
		saleLeads = m["sale"] + m["paid"]
	}

	resp.Stages = []models.FunnelStage{
		{Stage: "traffic", Label: "Трафик", Count: int(newLeads + contactLeads), Conversion: 100},
		{Stage: "consultation", Label: "Консультация", Count: int(contactLeads), Conversion: 0},
		{Stage: "measurement", Label: "Замер", Count: int(meetingLeads), Conversion: 0},
		{Stage: "kp", Label: "КП", Count: int(waitLeads), Conversion: 0},
		{Stage: "contract", Label: "Договор", Count: int(saleLeads), Conversion: 0},
		{Stage: "payment", Label: "Оплата", Count: 0, Conversion: 0},
	}

	// План по салонам — батчим N+1 (был Pluck+2 SUM per salon → 3 batched)
	type goalByAssignee struct {
		AssigneeID uuid.UUID `gorm:"column:assignee_id"`
		Plan       float64   `gorm:"column:plan"`
	}
	type factBySalon struct {
		SalonID uuid.UUID `gorm:"column:salon_id"`
		Fact    float64   `gorm:"column:fact"`
	}
	// 1) все менеджеры салонов одним запросом
	var allManagerRows []struct {
		ID      uuid.UUID `gorm:"column:id"`
		SalonID uuid.UUID `gorm:"column:salon_id"`
	}
	s.DB.Model(&models.User{}).Select("id, salon_id").Where("salon_id IN ?", salonIDs).Scan(&allManagerRows)
	managersBySalon := map[uuid.UUID][]uuid.UUID{}
	var allManagerIDs []uuid.UUID
	for _, r := range allManagerRows {
		managersBySalon[r.SalonID] = append(managersBySalon[r.SalonID], r.ID)
		allManagerIDs = append(allManagerIDs, r.ID)
	}
	// 2) планы по менеджерам batched
	planByAssignee := map[uuid.UUID]float64{}
	if len(allManagerIDs) > 0 {
		var goalRows []goalByAssignee
		s.DB.Model(&models.Goal{}).Select("assignee_id, COALESCE(SUM(sales_plan),0) as plan").
			Where("assignee_id IN ? AND target_date BETWEEN ? AND ?", allManagerIDs, startDate, endDate).
			Group("assignee_id").Scan(&goalRows)
		for _, g := range goalRows {
			planByAssignee[g.AssigneeID] = g.Plan
		}
	}
	planBySalon := map[uuid.UUID]float64{}
	for sid, mids := range managersBySalon {
		sum := 0.0
		for _, mid := range mids {
			sum += planByAssignee[mid]
		}
		planBySalon[sid] = sum
	}
	// 3) fallback dealer plan (один запрос вместо N) — goals.assignee_id, не dealer_id
	var dealerPlan float64
	s.DB.Model(&models.Goal{}).Where("assignee_id = ? AND target_date BETWEEN ? AND ?", userID, startDate, endDate).
		Select("COALESCE(SUM(sales_plan),0)").Scan(&dealerPlan)
	// 4) факты по салонам batched
	factBySalonMap := map[uuid.UUID]float64{}
	var factRows []factBySalon
	s.DB.Model(&models.Lead{}).Select("salon_id, COALESCE(SUM(budget),0) as fact").
		Where("salon_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", salonIDs, []string{"sale", "paid"}, startDate, endDate).
		Group("salon_id").Scan(&factRows)
	for _, f := range factRows {
		factBySalonMap[f.SalonID] = f.Fact
	}

	for _, salon := range salons {
		plan := planBySalon[salon.ID]
		if plan == 0 {
			plan = dealerPlan
		}
		fact := factBySalonMap[salon.ID]

		percent := 0
		if plan > 0 {
			percent = int((fact / plan) * 100)
		}

		forecast := "red"
		if percent >= 100 {
			forecast = "green"
		} else if percent >= 70 {
			forecast = "yellow"
		}

		resp.SalonPlanData = append(resp.SalonPlanData, models.SalonPlanData{
			ID:            salon.ID,
			Name:          salon.Name,
			Plan:          plan,
			Fact:          fact,
			Percent:       percent,
			Forecast:      forecast,
			ManagersCount: 1, // Each salon has at least one manager
			AvgCheck:      0,
		})
	}

	return resp, nil
}

// GetDealerProducts - товары для дилера
func (s *KPIService) GetDealerProducts(ctx context.Context, userID uuid.UUID, dateStr string) (*models.DealerProductsResponse, error) {
	resp := &models.DealerProductsResponse{}

	var managers []models.User
	if err := s.DB.Where("managed_by = ?", userID).Limit(100).Find(&managers).Error; err != nil {
		return nil, err
	}

	var salonIDs []uuid.UUID
	for _, mgr := range managers {
		if mgr.SalonID != nil {
			salonIDs = append(salonIDs, *mgr.SalonID)
		}
	}

	if len(salonIDs) == 0 {
		return resp, nil
	}

	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}
	firstOfMonth := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())

	// Общая выручка
	s.DB.Model(&models.Lead{}).
		Where("salon_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", salonIDs, []string{"sale", "paid"}, firstOfMonth, targetDate).
		Select("COALESCE(SUM(budget), 0)").Scan(&resp.TotalRevenue)

	// Топ товары
	type productStats struct {
		Name     string
		Revenue  float64
		Quantity int
	}

	var products []productStats
	rows, err := s.DB.Model(&models.Lead{}).
		Where("salon_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", salonIDs, []string{"sale", "paid"}, firstOfMonth, targetDate).
		Select("interest_product, COALESCE(SUM(budget), 0) as revenue, COUNT(*) as quantity").
		Group("interest_product").
		Order("revenue DESC").
		Limit(10).
		Rows()
	if err == nil {
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var ps productStats
_ = rows.Scan(&ps.Name, &ps.Revenue, &ps.Quantity)
			if ps.Name != "" {
				products = append(products, ps)
			}
		}
	}

	for _, ps := range products {
		share := 0.0
		if resp.TotalRevenue > 0 {
			share = (ps.Revenue / resp.TotalRevenue) * 100
		}
		resp.TopProducts = append(resp.TopProducts, models.DealerTopProduct{
			ID:           ps.Name,
			Name:         ps.Name,
			Revenue:      ps.Revenue,
			Quantity:     ps.Quantity,
			SharePercent: share,
		})
	}

	// Остатки (из каталога products по салонам дилера)
	type invRow struct {
		ID           string
		Name         string
		Collection   string
		Category     string
		ShowroomQty  int
		WarehouseQty int
		TurnoverDays int
	}
	var invRows []invRow
	s.DB.Table("products").
		Where("salon_id IN ?", salonIDs).
		Order("name").
		Scan(&invRows)
	for _, ir := range invRows {
		resp.Inventory = append(resp.Inventory, models.DealerInventoryItem{
			ID:             ir.ID,
			Collection:     ir.Collection,
			StockWarehouse: ir.WarehouseQty,
			OnDisplay:      ir.ShowroomQty,
			SoldPeriod:     0,
			TurnoverDays:   ir.TurnoverDays,
		})
	}

	return resp, nil
}

// GetFranchiserSummary - сводка для франчайзера
func (s *KPIService) GetFranchiserSummary(ctx context.Context, userID uuid.UUID, dateStr string) (*models.FranchiserSummaryResponse, error) {
	resp := &models.FranchiserSummaryResponse{}

	// Находим всех дилеров через tenant_id франчайзера
	var user models.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}

	// Ищем салоны через tenant_id
	var salons []models.Salon
	if err := s.DB.Where("dealer_id = ?", user.TenantID).Limit(100).Find(&salons).Error; err != nil {
		return nil, err
	}

	var salonIDs []uuid.UUID
	for _, salon := range salons {
		salonIDs = append(salonIDs, salon.ID)
	}

	if len(salonIDs) == 0 {
		return resp, nil
	}

	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}
	firstOfMonth := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())

	// Общая выручка из лидов (paid + contract)
	var totalRevenue float64
	s.DB.Model(&models.Lead{}).
		Where("salon_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", salonIDs, []string{"paid", "contract"}, firstOfMonth, targetDate).
		Select("COALESCE(SUM(budget), 0)").Scan(&totalRevenue)

	// Общий план из goals
	var totalPlan float64
	s.DB.Model(&models.Goal{}).
		Where("assignee_id = ? AND period = 'monthly'", userID).
		Select("COALESCE(SUM(sales_plan), 0)").Scan(&totalPlan)

	// Процент плана
	planPercent := 0
	if totalPlan > 0 {
		planPercent = int((totalRevenue / totalPlan) * 100)
	}

	// Прогноз
	daysInMonth := targetDate.Day()
	dailyAvg := totalRevenue / float64(daysInMonth)
	forecastAmount := dailyAvg * 30
	forecastPercent := 0
	if totalPlan > 0 {
		forecastPercent = int((forecastAmount / totalPlan) * 100)
	}

	// Средняя конверсия
	var totalLeads, totalSales int64
	s.DB.Model(&models.Lead{}).Where("salon_id IN ? AND created_at BETWEEN ? AND ?", salonIDs, firstOfMonth, targetDate).Count(&totalLeads)
	s.DB.Model(&models.Lead{}).Where("salon_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", salonIDs, []string{"paid", "contract"}, firstOfMonth, targetDate).Count(&totalSales)

	avgConversion := 0.0
	if totalLeads > 0 {
		avgConversion = (float64(totalSales) / float64(totalLeads)) * 100
	}

	resp.PlanPercent = planPercent
	resp.ForecastPercent = forecastPercent
	resp.ActiveDealers = len(salons)
	resp.AvgConversion = avgConversion
	resp.AvgMargin = s.getSettingFloat("avg_margin_percent", 32.0)

	return resp, nil
}

// GetFranchiserNetwork - данные сети
func (s *KPIService) GetFranchiserNetwork(ctx context.Context, userID uuid.UUID, period, dateStr string) (*models.FranchiserNetworkResponse, error) {
	resp := &models.FranchiserNetworkResponse{}

	// Шаг 1: Найти всех менеджеров под этим франчайзером
	var managers []models.User
	if err := s.DB.Where("role = ? AND managed_by = ?", models.RoleFranchisorManager, userID).Limit(100).Find(&managers).Error; err != nil {
		return nil, err
	}

	if len(managers) == 0 {
		return resp, nil
	}

	// Шаг 2: Найти всех дилеров под этими менеджерами
	var managerIDs []uuid.UUID
	for _, m := range managers {
		managerIDs = append(managerIDs, m.ID)
	}

	var dealers []models.User
	if err := s.DB.Where("role = ? AND managed_by IN ?", models.RoleDealer, managerIDs).Limit(100).Find(&dealers).Error; err != nil {
		return nil, err
	}

	if len(dealers) == 0 {
		return resp, nil
	}

	var dealerIDs []uuid.UUID
	for _, d := range dealers {
		dealerIDs = append(dealerIDs, d.ID)
	}

	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}

	var startDate, endDate time.Time
	switch period {
	case "week":
		startDate = targetDate.AddDate(0, 0, -7)
		endDate = targetDate
	case "quarter":
		m := int(targetDate.Month())
		quarterStart := time.Month(((m-1)/3)*3 + 1)
		startDate = time.Date(targetDate.Year(), quarterStart, 1, 0, 0, 0, 0, targetDate.Location())
		endDate = targetDate
	default: // month
		startDate = time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())
		endDate = targetDate
	}

	// По каждому дилеру — батчим 3*N → 3 GROUP BY
	planByDealer := map[uuid.UUID]float64{}
	factByDealer := map[uuid.UUID]float64{}
	salonsCountByTenant := map[uuid.UUID]int64{}
	{
		type pRow struct {
			UserID uuid.UUID `gorm:"column:user_id"`
			Plan   float64   `gorm:"column:plan"`
		}
		var pRows []pRow
		s.DB.Model(&models.DailyGoal{}).Select("user_id, COALESCE(SUM(sales_plan),0) as plan").
			Where("user_id IN ? AND target_date BETWEEN ? AND ?", dealerIDs, startDate, endDate).Group("user_id").Scan(&pRows)
		for _, r := range pRows {
			planByDealer[r.UserID] = r.Plan
		}
		type fRow struct {
			ManagerID uuid.UUID `gorm:"column:manager_id"`
			Fact      float64   `gorm:"column:fact"`
		}
		var fRows []fRow
		s.DB.Model(&models.Lead{}).Select("manager_id, COALESCE(SUM(budget),0) as fact").
			Where("manager_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", dealerIDs, []string{"sale", "paid"}, startDate, endDate).Group("manager_id").Scan(&fRows)
		for _, r := range fRows {
			factByDealer[r.ManagerID] = r.Fact
		}
		var tenantIDs []uuid.UUID
		for _, d := range dealers {
			if d.TenantID != nil {
				tenantIDs = append(tenantIDs, *d.TenantID)
			}
		}
		if len(tenantIDs) > 0 {
			type scRow struct {
				DealerID uuid.UUID `gorm:"column:dealer_id"`
				Cnt      int64     `gorm:"column:cnt"`
			}
			var scRows []scRow
			s.DB.Model(&models.Salon{}).Select("dealer_id, COUNT(*) as cnt").Where("dealer_id IN ?", tenantIDs).Group("dealer_id").Scan(&scRows)
			for _, r := range scRows {
				salonsCountByTenant[r.DealerID] = r.Cnt
			}
		}
	}
	for _, dealer := range dealers {
		plan := planByDealer[dealer.ID]
		fact := factByDealer[dealer.ID]
		percent := 0
		if plan > 0 {
			percent = int((fact / plan) * 100)
		}
		var salonsCount int64
		if dealer.TenantID != nil {
			salonsCount = salonsCountByTenant[*dealer.TenantID]
		}
		forecast := "red"
		if percent >= 80 {
			forecast = "green"
		} else if percent >= 50 {
			forecast = "yellow"
		}
		resp.NetworkData = append(resp.NetworkData, models.FranchiserDealerData{
			ID:          dealer.ID,
			Name:        dealer.FirstName + " " + dealer.LastName,
			SalonCount:  int(salonsCount),
			Plan:        plan,
			Fact:        fact,
			PlanPercent: percent,
			Forecast:    forecast,
		})
	}

	// Общие метрики
	var totalPlan, leadFact, orderFact, totalFact float64
	s.DB.Model(&models.DailyGoal{}).
		Where("user_id IN ? AND target_date BETWEEN ? AND ?", dealerIDs, startDate, endDate).
		Select("COALESCE(SUM(sales_plan), 0)").Scan(&totalPlan)

	// Факт из лидов (статус sale или paid)
	s.DB.Model(&models.Lead{}).
		Where("manager_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", dealerIDs, []string{"sale", "paid"}, startDate, endDate).
		Select("COALESCE(SUM(budget), 0)").Scan(&leadFact)

	// Получаем фактические продажи из заказов (ORDERS)
	// Сначала найдём салоны, принадлежащие этим дилерам через их tenant_id
	var salonIDs []uuid.UUID
	for _, dealer := range dealers {
		if dealer.TenantID != nil {
			// Salon.dealer_id = User.tenant_id (это связь!)
			var salons []models.Salon
			s.DB.Where("dealer_id = ?", *dealer.TenantID).Limit(100).Find(&salons)
			for _, salon := range salons {
				salonIDs = append(salonIDs, salon.ID)
			}
		}
	}

	if len(salonIDs) > 0 {
		s.DB.Model(&models.Order{}).
			Where("salon_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", salonIDs, []string{"paid", "contract"}, startDate, endDate).
			Select("COALESCE(SUM(total_price), 0)").Scan(&orderFact)
	}

	// Используем максимальное значение из обоих источников
	if orderFact > leadFact {
		totalFact = orderFact
	} else {
		totalFact = leadFact
	}

	// Рассчитать процент плана
	planPercent := 0.0
	if totalPlan > 0 {
		planPercent = (totalFact / totalPlan) * 100
	}

	// Прогноз на квартал — динамически по длине квартала
	forecastAmount := 0.0
	daysPassed := targetDate.Sub(startDate).Hours() / 24
	if daysPassed > 0 {
		m := int(targetDate.Month())
		qm := time.Month(((m - 1) / 3) * 3 + 1)
		qStart := time.Date(targetDate.Year(), qm, 1, 0, 0, 0, 0, targetDate.Location())
		qEnd := qStart.AddDate(0, 3, 0)
		daysInQuarter := qEnd.Sub(qStart).Hours() / 24
		forecastAmount = (totalFact / daysPassed) * daysInQuarter
	}
	forecastPercent := 0.0
	if totalPlan > 0 {
		forecastPercent = (forecastAmount / totalPlan) * 100
	}

	// Красная зона: дилеры с < 50% плана
	redZoneCount := 0
	for _, dealer := range dealers {
		// Находим салоны этого дилера
		var dealerSalonIDs []uuid.UUID
		if dealer.TenantID != nil {
			var salons []models.Salon
			s.DB.Where("dealer_id = ?", *dealer.TenantID).Limit(100).Find(&salons)
			for _, salon := range salons {
				dealerSalonIDs = append(dealerSalonIDs, salon.ID)
			}
		}

		var plan float64
		s.DB.Model(&models.DailyGoal{}).
			Where("user_id = ? AND target_date BETWEEN ? AND ?", dealer.ID, startDate, endDate).
			Select("COALESCE(SUM(sales_plan), 0)").Scan(&plan)

		if plan > 0 && len(dealerSalonIDs) > 0 {
			var fact float64
			s.DB.Model(&models.Order{}).
				Where("salon_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", dealerSalonIDs, []string{"paid", "contract"}, startDate, endDate).
				Select("COALESCE(SUM(total_price), 0)").Scan(&fact)
			if int((fact/plan)*100) < 50 {
				redZoneCount++
			}
		}
	}

	// Средняя конверсия (упрощённо)
	avgConversion := 50.0
	if len(dealers) == 0 {
		avgConversion = 0
	}

	resp.Overview = models.FranchiserOverview{
		PlanAmount:      totalPlan,
		PlanPercent:     int(planPercent),
		ForecastAmount:  forecastAmount,
		ForecastPercent: int(forecastPercent),
		ActiveDealers:   len(dealers),
		AvgConversion:   avgConversion,
		AvgMargin:       0,
		RedZoneDealers:  redZoneCount,
	}

	return resp, nil
}

// GetFranchiserHealth - здоровье сети
func (s *KPIService) GetFranchiserHealth(ctx context.Context, userID uuid.UUID) (*models.FranchiserHealthResponse, error) {
	resp := &models.FranchiserHealthResponse{}

	// Используем правильную иерархию: franchiser → managers → dealers
	var managers []models.User
	if err := s.DB.Where("role = ? AND managed_by = ?", models.RoleFranchisorManager, userID).Limit(100).Find(&managers).Error; err != nil {
		return nil, err
	}

	if len(managers) == 0 {
		return resp, nil
	}

	var managerIDs []uuid.UUID
	for _, m := range managers {
		managerIDs = append(managerIDs, m.ID)
	}

	var dealers []models.User
	if err := s.DB.Where("role = ? AND managed_by IN ?", models.RoleDealer, managerIDs).Limit(100).Find(&dealers).Error; err != nil {
		return nil, err
	}

	var dealerIDs []uuid.UUID
	for _, d := range dealers {
		dealerIDs = append(dealerIDs, d.ID)
	}

	if len(dealerIDs) == 0 {
		return resp, nil
	}

	// SLA (все лиды обработаны за 24ч)
	var totalLeads, processedLeads int64
	s.DB.Model(&models.Lead{}).Where("manager_id IN ?", dealerIDs).Count(&totalLeads)
	s.DB.Model(&models.Lead{}).Where("manager_id IN ? AND updated_at > ?", dealerIDs, time.Now().Add(-24*time.Hour)).Count(&processedLeads)

	if totalLeads > 0 {
		resp.SLA = float64(processedLeads) / float64(totalLeads) * 100
	} else {
		resp.SLA = 100
	}

	// Красные дилеры (выполнение < 50%)
	for _, dealer := range dealers {
		var fact float64
		s.DB.Model(&models.Lead{}).
			Where("manager_id = ? AND status IN ?", dealer.ID, []string{"sale", "paid"}).
			Select("COALESCE(SUM(budget), 0)").Scan(&fact)

		if fact == 0 {
			resp.RedDealers = append(resp.RedDealers, models.FranchiserDealerHealth{
				DealerID:   dealer.ID,
				DealerName: dealer.FirstName + " " + dealer.LastName,
				Issue:      "Нет продаж",
				Severity:   "critical",
			})
		}
	}

	// Общие показатели
	resp.TotalDealers = len(dealers)
	resp.ActiveDealers = len(dealers)
	resp.TotalSalons = 0 // Запросить через Salon model

	return resp, nil
}

// GetFranchiserTeam - команда франчайзера
func (s *KPIService) GetFranchiserTeam(ctx context.Context, userID uuid.UUID) (*models.FranchiserTeamResponse, error) {
	resp := &models.FranchiserTeamResponse{}

	// franchiser_manager - менеджеры франчайзера
	var managers []models.User
	if err := s.DB.Where("role = ? AND managed_by = ?", models.RoleFranchisorManager, userID).Limit(100).Find(&managers).Error; err != nil {
		return nil, err
	}

	// Батчим dealersCount — 1 GROUP BY вместо N
	dealersByMgr := map[uuid.UUID]int64{}
	{
		var mgrIDs []uuid.UUID
		for _, m := range managers {
			mgrIDs = append(mgrIDs, m.ID)
		}
		if len(mgrIDs) > 0 {
			type cRow struct {
				ManagedBy uuid.UUID `gorm:"column:managed_by"`
				Cnt       int64     `gorm:"column:cnt"`
			}
			var cRows []cRow
			s.DB.Model(&models.User{}).Select("managed_by, COUNT(*) as cnt").
				Where("role = ? AND managed_by IN ?", models.RoleDealer, mgrIDs).Group("managed_by").Scan(&cRows)
			for _, r := range cRows {
				dealersByMgr[r.ManagedBy] = r.Cnt
			}
		}
	}
	for _, mgr := range managers {
		dealersCount := dealersByMgr[mgr.ID]
		resp.TeamMembers = append(resp.TeamMembers, models.FranchiserTeamMember{
			ID:           mgr.ID,
			Name:         mgr.FirstName + " " + mgr.LastName,
			DealersCount: int(dealersCount),
			Role:         "Менеджер сети",
		})
	}

	return resp, nil
}

// GetTerritorySummary - сводка для территориального менеджера
func (s *KPIService) GetTerritorySummary(ctx context.Context, userID uuid.UUID, dateStr string) (*models.TerritorySummaryResponse, error) {
	resp := &models.TerritorySummaryResponse{}

	// Находим всех дилеров, которыми управляет территориальный менеджер
	var dealers []models.User
	if err := s.DB.Where("role = ? AND managed_by = ?", models.RoleDealer, userID).Limit(100).Find(&dealers).Error; err != nil {
		return nil, err
	}

	if len(dealers) == 0 {
		return resp, nil
	}

	var dealerIDs []uuid.UUID
	for _, d := range dealers {
		dealerIDs = append(dealerIDs, d.ID)
	}

	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}
	firstOfMonth := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())

	// План и факт
	var totalPlan, totalFact float64
	s.DB.Model(&models.DailyGoal{}).
		Where("user_id IN ? AND target_date BETWEEN ? AND ?", dealerIDs, firstOfMonth, targetDate).
		Select("COALESCE(SUM(sales_plan), 0)").Scan(&totalPlan)
	s.DB.Model(&models.Lead{}).
		Where("manager_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", dealerIDs, []string{"sale", "paid"}, firstOfMonth, targetDate).
		Select("COALESCE(SUM(budget), 0)").Scan(&totalFact)

	resp.PlanCompletionPercent = 0
	if totalPlan > 0 {
		resp.PlanCompletionPercent = int((totalFact / totalPlan) * 100)
	}

	// Прогноз
	daysInMonth := targetDate.Day()
	dailyAvg := totalFact / float64(daysInMonth)
	forecastAmount := dailyAvg * 30
	resp.QuarterForecastPercent = 0
	if totalPlan > 0 {
		resp.QuarterForecastPercent = int((forecastAmount / totalPlan) * 100)
	}

	// Красные дилеры — батчим N → 1 GROUP BY
	redZoneDealersCount := 0
	{
		type rRow struct {
			ManagerID uuid.UUID `gorm:"column:manager_id"`
			Fact      float64   `gorm:"column:fact"`
		}
		var rRows []rRow
		s.DB.Model(&models.Lead{}).Select("manager_id, COALESCE(SUM(budget),0) as fact").
			Where("manager_id IN ? AND status IN ?", dealerIDs, []string{"sale", "paid"}).
			Group("manager_id").Scan(&rRows)
		mFact := map[uuid.UUID]float64{}
		for _, r := range rRows {
			mFact[r.ManagerID] = r.Fact
		}
		for _, d := range dealers {
			fact := mFact[d.ID]
			if totalPlan > 0 && (fact/totalPlan)*100 < 50 {
				redZoneDealersCount++
			}
		}
	}
	resp.RedZoneDealersCount = redZoneDealersCount

	// Конверсия
	var totalLeads, totalSales int64
	s.DB.Model(&models.Lead{}).Where("manager_id IN ? AND created_at BETWEEN ? AND ?", dealerIDs, firstOfMonth, targetDate).Count(&totalLeads)
	s.DB.Model(&models.Lead{}).Where("manager_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", dealerIDs, []string{"sale", "paid"}, firstOfMonth, targetDate).Count(&totalSales)

	resp.AvgConversion = 0
	if totalLeads > 0 {
		resp.AvgConversion = (float64(totalSales) / float64(totalLeads)) * 100
	}

	resp.ActiveAlerts = redZoneDealersCount

	return resp, nil
}

// GetTerritoryFunnel - воронка для территориального менеджера
func (s *KPIService) GetTerritoryFunnel(ctx context.Context, userID uuid.UUID, period, dateStr string) (*models.TerritoryFunnelResponse, error) {
	resp := &models.TerritoryFunnelResponse{}

	var dealers []models.User
	if err := s.DB.Where("role = ? AND managed_by = ?", models.RoleDealer, userID).Limit(100).Find(&dealers).Error; err != nil {
		return nil, err
	}

	var dealerIDs []uuid.UUID
	for _, d := range dealers {
		dealerIDs = append(dealerIDs, d.ID)
	}

	if len(dealerIDs) == 0 {
		return resp, nil
	}

	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, parsed.Location())
		}
	}

	// Воронка по этапам
	var newLeads, contactLeads, meetingLeads, saleLeads int64
	s.DB.Model(&models.Lead{}).Where("manager_id IN ? AND created_at >= ?", dealerIDs, targetDate.AddDate(0, 0, -30)).Count(&newLeads)
	s.DB.Model(&models.Lead{}).Where("manager_id IN ? AND status = ? AND created_at >= ?", dealerIDs, "contact", targetDate.AddDate(0, 0, -30)).Count(&contactLeads)
	s.DB.Model(&models.Lead{}).Where("manager_id IN ? AND status = ? AND created_at >= ?", dealerIDs, "meeting", targetDate.AddDate(0, 0, -30)).Count(&meetingLeads)
	s.DB.Model(&models.Lead{}).Where("manager_id IN ? AND status IN ? AND created_at >= ?", dealerIDs, []string{"sale", "paid"}, targetDate.AddDate(0, 0, -30)).Count(&saleLeads)

	resp.Stages = []models.FunnelStage{
		{Stage: "leads", Label: "Лиды", Count: int(newLeads), Conversion: 100},
		{Stage: "contact", Label: "Контакт", Count: int(contactLeads), Conversion: 0},
		{Stage: "meeting", Label: "Встреча", Count: int(meetingLeads), Conversion: 0},
		{Stage: "sale", Label: "Продажа", Count: int(saleLeads), Conversion: 0},
	}

	return resp, nil
}

// GetTerritoryPlanFact - план-факт
func (s *KPIService) GetTerritoryPlanFact(ctx context.Context, userID uuid.UUID, period string) (*models.TerritoryPlanFactResponse, error) {
	resp := &models.TerritoryPlanFactResponse{}

	var dealers []models.User
	if err := s.DB.Where("role = ? AND managed_by = ?", models.RoleDealer, userID).Limit(100).Find(&dealers).Error; err != nil {
		return nil, err
	}

	var dealerIDs []uuid.UUID
	for _, d := range dealers {
		dealerIDs = append(dealerIDs, d.ID)
	}

	if len(dealerIDs) == 0 {
		return resp, nil
	}

	now := time.Now()
	var startDate, endDate time.Time
	switch period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
		endDate = now
	case "quarter":
		m := int(now.Month())
		quarterStart := time.Month(((m-1)/3)*3 + 1)
		startDate = time.Date(now.Year(), quarterStart, 1, 0, 0, 0, 0, now.Location())
		endDate = now
	default: // month
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = now
	}

	// По каждому дилеру
	for _, dealer := range dealers {
		var plan, fact float64
		s.DB.Model(&models.DailyGoal{}).
			Where("user_id = ? AND target_date BETWEEN ? AND ?", dealer.ID, startDate, endDate).
			Select("COALESCE(SUM(sales_plan), 0)").Scan(&plan)
		s.DB.Model(&models.Lead{}).
			Where("manager_id = ? AND status IN ? AND created_at BETWEEN ? AND ?", dealer.ID, []string{"sale", "paid"}, startDate, endDate).
			Select("COALESCE(SUM(budget), 0)").Scan(&fact)

		percent := 0
		if plan > 0 {
			percent = int((fact / plan) * 100)
		}

		resp.Dealers = append(resp.Dealers, models.TerritoryDealerPlanFact{
			ID:          dealer.ID,
			DealerName:  dealer.FirstName + " " + dealer.LastName,
			Plan:        plan,
			Fact:        fact,
			PlanPercent: percent,
		})
	}

	// Итого
	var totalPlan, totalFact float64
	s.DB.Model(&models.DailyGoal{}).
		Where("user_id IN ? AND target_date BETWEEN ? AND ?", dealerIDs, startDate, endDate).
		Select("COALESCE(SUM(sales_plan), 0)").Scan(&totalPlan)
	s.DB.Model(&models.Lead{}).
		Where("manager_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", dealerIDs, []string{"sale", "paid"}, startDate, endDate).
		Select("COALESCE(SUM(budget), 0)").Scan(&totalFact)

	resp.TotalPlan = totalPlan
	resp.TotalFact = totalFact

	return resp, nil
}

// GetTerritoryCommunications - коммуникации
func (s *KPIService) GetTerritoryCommunications(ctx context.Context, userID uuid.UUID) (*models.TerritoryCommunicationsResponse, error) {
	resp := &models.TerritoryCommunicationsResponse{}

	// Задачи (от franchiser_manager к дилерам)
	var dealers []models.User
	if err := s.DB.Where("role = ? AND managed_by = ?", models.RoleDealer, userID).Limit(100).Find(&dealers).Error; err != nil {
		return nil, err
	}

	var dealerIDs []uuid.UUID
	for _, d := range dealers {
		dealerIDs = append(dealerIDs, d.ID)
	}

	// Недавние задачи
	var tasks []models.Task
	if len(dealerIDs) > 0 {
		s.DB.Where("assigned_to IN ? AND created_at > ?", dealerIDs, time.Now().AddDate(0, 0, -7)).Find(&tasks)
	}

	for _, task := range tasks {
		resp.Tasks = append(resp.Tasks, models.TerritoryTask{
			ID:         task.ID,
			Title:      task.Title,
			Status:     task.Status,
			AssignedTo: task.AssignedTo,
			DueDate:    task.DueDate.Format("2006-01-02"),
		})
	}

	// Непрочитанные сообщения
	var messagesCount int64
	s.DB.Model(&models.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&messagesCount)
	resp.UnreadMessages = int(messagesCount)

	return resp, nil
}

// GetTerritoryBenchmarks - бенчмарки
func (s *KPIService) GetTerritoryBenchmarks(ctx context.Context, userID uuid.UUID) (*models.TerritoryBenchmarksResponse, error) {
	resp := &models.TerritoryBenchmarksResponse{}

	var dealers []models.User
	if err := s.DB.Where("role = ? AND managed_by = ?", models.RoleDealer, userID).Limit(100).Find(&dealers).Error; err != nil {
		return nil, err
	}

	var dealerIDs []uuid.UUID
	for _, d := range dealers {
		dealerIDs = append(dealerIDs, d.ID)
	}

	if len(dealerIDs) == 0 {
		return resp, nil
	}

	// Средняя конверсия территории
	var totalLeads, totalSales int64
	s.DB.Model(&models.Lead{}).Where("manager_id IN ?", dealerIDs).Count(&totalLeads)
	s.DB.Model(&models.Lead{}).Where("manager_id IN ? AND status IN ?", dealerIDs, []string{"sale", "paid"}).Count(&totalSales)

	resp.TerritoryConversion = 0
	if totalLeads > 0 {
		resp.TerritoryConversion = (float64(totalSales) / float64(totalLeads)) * 100
	}

	// Средний чек
	var avgCheck float64
	s.DB.Model(&models.Lead{}).
		Where("manager_id IN ? AND status IN ?", dealerIDs, []string{"sale", "paid"}).
		Select("COALESCE(AVG(budget), 0)").Scan(&avgCheck)
	resp.TerritoryAvgCheck = avgCheck

	// Эталон (сеть) — из настроек
	resp.NetworkAvgConversion = s.getSettingFloat("network_avg_conversion", 15.0)
	resp.NetworkAvgCheck = s.getSettingFloat("network_avg_check", 80000)

	return resp, nil
}

// === Dealer Tasks Service ===

func (s *KPIService) GetDealerTasks(ctx context.Context, userID uuid.UUID) (*models.DealerTasksResponse, error) {
	resp := &models.DealerTasksResponse{}

	// Проверяем наличие таблицы
	var count int64
	s.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'dealer_tasks'").Scan(&count)
	if count == 0 {
		return resp, nil // таблицы нет - возвращаем пустой ответ
	}

	// Читаем задачи из таблицы dealer_tasks
	type DealerTask struct {
		ID          uuid.UUID  `json:"id"`
		Title       string     `json:"title"`
		Description string     `json:"description"`
		Status      string     `json:"status"`
		Priority    string     `json:"priority"`
		DueDate     *time.Time `json:"due_date"`
		CreatedAt   time.Time  `json:"created_at"`
	}
	rows, err := s.DB.Raw(`
		SELECT id, title, description, status, priority, due_date, created_at 
		FROM dealer_tasks 
		WHERE dealer_id = ?
		ORDER BY due_date ASC, created_at DESC
	`, userID).Rows()

	if err != nil {
		return resp, nil
	}
	defer func() { _ = rows.Close() }()

	now := time.Now()
	for rows.Next() {
		var t DealerTask
_ = rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.DueDate, &t.CreatedAt)

		// Определяем статус просрочки
		isOverdue := t.DueDate != nil && t.DueDate.Before(now) && t.Status != "done"

		resp.Tasks = append(resp.Tasks, models.DealerTaskItem{
			ID:          t.ID,
			Title:       t.Title,
			Description: t.Description,
			Status:      t.Status,
			Priority:    t.Priority,
			DueDate:     t.DueDate.Format("2006-01-02"),
			IsOverdue:   isOverdue,
		})
	}

	resp.Total = len(resp.Tasks)
	resp.Overdue = 0
	for _, t := range resp.Tasks {
		if t.IsOverdue {
			resp.Overdue++
		}
	}

	return resp, nil
}

func (s *KPIService) UpdateDealerTask(ctx context.Context, userID, taskID, status string) error {
	return s.DB.Exec(`
		UPDATE dealer_tasks 
		SET status = ?, updated_at = NOW() 
		WHERE id = ? AND dealer_id = ?
	`, status, taskID, userID).Error
}

// === Dealer Requests Service ===

func isMissingTableErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "does not exist") ||
		strings.Contains(msg, "no such table") ||
		strings.Contains(msg, "no such database")
}

func (s *KPIService) GetDealerRequests(ctx context.Context, userID uuid.UUID, statusFilter string) (*models.DealerRequestsResponse, error) {
	resp := &models.DealerRequestsResponse{}

	
	query := "SELECT id, type, description, amount, status, created_at FROM dealer_requests WHERE dealer_id = ?"
	args := []interface{}{userID.String()}

	if statusFilter != "" {
		query += " AND status = ?"
		args = append(args, statusFilter)
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.DB.Raw(query, args...).Rows()
	if err != nil {
		if isMissingTableErr(err) {
			return resp, nil
		}
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var r models.DealerRequest
_ = rows.Scan(&r.ID, &r.Type, &r.Description, &r.Amount, &r.Status, &r.CreatedAt)
		resp.Requests = append(resp.Requests, models.DealerRequestItem{
			ID:          r.ID,
			DealerID:    userID,
			Type:        r.Type,
			Description: r.Description,
			Amount:      r.Amount,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt,
		})
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	resp.Total = len(resp.Requests)
	resp.Pending = 0
	for _, r := range resp.Requests {
		if r.Status == "pending" {
			resp.Pending++
		}
	}

	return resp, nil
}

func (s *KPIService) CreateDealerRequest(ctx context.Context, userID uuid.UUID, reqType, description string, amount float64) error {
	return s.DB.Exec(`
		INSERT INTO dealer_requests (id, dealer_id, type, description, amount, status, created_at, updated_at)
		VALUES (gen_random_uuid(), ?, ?, ?, ?, 'pending', NOW(), NOW())
	`, userID.String(), reqType, description, amount).Error
}

// === Dealer Marketing Budget Service ===

func (s *KPIService) GetDealerMarketingBudget(ctx context.Context, userID uuid.UUID, quarter string) (*models.DealerMarketingBudgetResponse, error) {
	resp := &models.DealerMarketingBudgetResponse{
		Quarter: quarter,
	}

	type MarketingBudget struct {
		TotalAmount float64 `json:"total_amount"`
		UsedAmount  float64 `json:"used_amount"`
	}

	var mb MarketingBudget
	result := s.DB.Raw(`
		SELECT total_amount, used_amount 
		FROM marketing_budgets 
		WHERE dealer_id = ? AND quarter = ?
	`, userID.String(), quarter).Scan(&mb)

	switch {
	case result.Error != nil && isMissingTableErr(result.Error):
		return nil, result.Error
	case result.Error != nil:
		return nil, result.Error
	case result.RowsAffected == 0:
		resp.TotalAmount = 0
		resp.UsedAmount = 0
	default:
		resp.TotalAmount = mb.TotalAmount
		resp.UsedAmount = mb.UsedAmount
	}

	resp.Remaining = resp.TotalAmount - resp.UsedAmount
	if resp.TotalAmount > 0 {
		resp.UsagePercent = int((resp.UsedAmount / resp.TotalAmount) * 100)
	}

	return resp, nil
}

// === Dealer Alerts Service ===

func (s *KPIService) GetDealerAlerts(ctx context.Context, userID uuid.UUID) (*models.DealerAlertsResponse, error) {
	resp := &models.DealerAlertsResponse{}

	// Читаем из notifications
	type Alert struct {
		ID       uuid.UUID `json:"id"`
		Type     string    `json:"type"`
		Title    string    `json:"title"`
		Message  string    `json:"message"`
		IsRead   bool      `json:"is_read"`
		Priority string    `json:"priority"`
	}

	rows, err := s.DB.Raw(`
		SELECT id, type, title, message, is_read
		FROM notifications 
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID).Rows()

	if err != nil {
		return resp, nil
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var a Alert
_ = rows.Scan(&a.ID, &a.Type, &a.Title, &a.Message, &a.IsRead)

		priority := "info"
		//nolint:staticcheck
		if a.Type == "conversion_drop" || a.Type == "task_overdue" {
			priority = "critical"
		} else if a.Type == "payroll_exceeded" {
			priority = "warning"
		}

		if !a.IsRead {
			resp.Alerts = append(resp.Alerts, models.DealerAlertItem{
				ID:       a.ID,
				Type:     a.Type,
				Title:    a.Title,
				Message:  a.Message,
				Priority: priority,
			})
		}
	}

	resp.UnreadCount = len(resp.Alerts)
	return resp, nil
}

func (s *KPIService) MarkDealerAlertRead(ctx context.Context, userID, alertID uuid.UUID) error {
	return s.DB.Exec(`
		UPDATE notifications 
		SET is_read = true 
		WHERE id = ? AND user_id = ?
	`, alertID, userID).Error
}

func (s *KPIService) MarkAllDealerAlertsRead(ctx context.Context, userID uuid.UUID) error {
	return s.DB.Exec(`
		UPDATE notifications 
		SET is_read = true 
		WHERE user_id = ?
	`, userID).Error
}

// GetFranchiserDealers - получить список дилеров
func (s *KPIService) GetFranchiserDealers(ctx context.Context, userID uuid.UUID, filter string) (*models.FranchiserDealersResponse, error) {
	resp := &models.FranchiserDealersResponse{
		Dealers: []models.FranchiserDealerItem{},
	}

	query := s.DB.Where("role = ? AND managed_by = ?", models.RoleDealer, userID)
	if filter == "problem" {
		query = query.Where("status = ?", "inactive")
	}

	var dealers []models.User
	if err := query.Limit(100).Find(&dealers).Error; err != nil {
		return nil, err
	}

	for _, dealer := range dealers {
		// Get sales data
		var totalSales float64
		s.DB.Model(&models.Order{}).
			Joins("JOIN salons ON salons.id = orders.salon_id").
			Where("salons.dealer_id = ? AND orders.status = ?", dealer.ID, "paid").
			Select("COALESCE(SUM(total_price), 0)").Scan(&totalSales)

		var plan float64
		s.DB.Model(&models.DailyGoal{}).
			Where("user_id = ?", dealer.ID).
			Select("COALESCE(SUM(sales_plan), 0)").Scan(&plan)

		percent := 0
		if plan > 0 {
			percent = int((totalSales / plan) * 100)
		}

		status := "green"
		if percent < 50 {
			status = "red"
		} else if percent < 80 {
			status = "yellow"
		}

		resp.Dealers = append(resp.Dealers, models.FranchiserDealerItem{
			ID:          dealer.ID,
			Name:        dealer.FirstName + " " + dealer.LastName,
			Email:       dealer.Email,
			Plan:        plan,
			Fact:        totalSales,
			PlanPercent: percent,
			Status:      status,
		})
	}

	resp.Total = len(resp.Dealers)
	return resp, nil
}

// GetFranchiserRequests - получить запросы от дилеров
func (s *KPIService) GetFranchiserRequests(ctx context.Context, userID uuid.UUID, filter string) (*models.FranchiserRequestsResponse, error) {
	resp := &models.FranchiserRequestsResponse{
		Requests: []models.FranchiserRequestItem{},
	}

	query := s.DB.Where("status = ?", "pending")
	if filter != "all" && filter != "" {
		query = query.Where("type = ?", filter)
	}

	var requests []models.DealerRequest
	if err := query.Order("created_at DESC").Find(&requests).Error; err != nil {
		return nil, err
	}

	for _, req := range requests {
		var dealerName string
		s.DB.Model(&models.User{}).Where("id = ?", req.DealerID).Select("first_name").Scan(&dealerName)

		resp.Requests = append(resp.Requests, models.FranchiserRequestItem{
			ID:          req.ID,
			DealerID:    req.DealerID,
			DealerName:  dealerName,
			Type:        req.Type,
			Description: req.Description,
			Amount:      req.Amount,
			Status:      req.Status,
			CreatedAt:   req.CreatedAt,
		})
	}

	resp.Total = len(resp.Requests)
	resp.Pending = len(resp.Requests)
	return resp, nil
}

// GetTerritoriesHeatmap - тепловая карта территорий
func (s *KPIService) GetTerritoriesHeatmap(ctx context.Context, userID string, period string) (map[string]interface{}, error) {
	resp := map[string]interface{}{"territories": []map[string]interface{}{}}

	userUUID, _ := uuid.Parse(userID)
	var managers []models.User
	if err := s.DB.Where("role = ? AND managed_by = ?", models.RoleFranchisorManager, userUUID).Limit(100).Find(&managers).Error; err != nil {
		return nil, err
	}

	var territories []map[string]interface{}
	for _, m := range managers {
		var managerIDs []uuid.UUID
		managerIDs = append(managerIDs, m.ID)

		var dealers []models.User
		s.DB.Where("role = ? AND managed_by IN ?", models.RoleDealer, managerIDs).Limit(100).Find(&dealers)

		var dealerIDs []uuid.UUID
		for _, d := range dealers {
			dealerIDs = append(dealerIDs, d.ID)
		}

		var plan float64
		if len(dealerIDs) > 0 {
			s.DB.Model(&models.DailyGoal{}).Where("user_id IN ?", dealerIDs).Select("COALESCE(SUM(sales_plan), 0)").Scan(&plan)
		}

		territories = append(territories, map[string]interface{}{
			"manager_id": m.ID,
			"name":       m.FirstName + " " + m.LastName,
			"dealers":    len(dealers),
			"plan":       plan,
			"status":     "green",
		})
	}
	resp["territories"] = territories
	return resp, nil
}

// GetManagerDynamics - динамика менеджера (реальный KPI по планам vs факту из leads)
func (s *KPIService) GetManagerDynamics(ctx context.Context, userID, managerID, months string) (map[string]interface{}, error) {
	resp := map[string]interface{}{"kpi": []map[string]interface{}{}}

	m, _ := strconv.Atoi(months)
	if m < 1 || m > 12 {
		m = 6
	}

	mgrUUID, err := uuid.Parse(managerID)
	if err != nil {
		// fallback на userID если managerID не uuid
		mgrUUID = uuid.MustParse(userID)
		_ = err
	}

	for i := 0; i < m; i++ {
		date := time.Now().AddDate(0, -i, 0)
		firstOfMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		lastOfMonth := firstOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

		var planAmount float64
		s.DB.Model(&models.Goal{}).
			Where("assignee_id = ? AND target_date BETWEEN ? AND ?", mgrUUID, firstOfMonth, lastOfMonth).
			Select("COALESCE(SUM(sales_plan), 0)").Scan(&planAmount)
		if planAmount == 0 {
			s.DB.Model(&models.Goal{}).
				Where("assignee_id = ? AND period IN ('month','year')", mgrUUID).
				Select("COALESCE(SUM(sales_plan), 0)").Scan(&planAmount)
		}
		if planAmount == 0 {
			planAmount = s.getSettingFloat("default_monthly_plan", 4000000)
		}

		var fact float64
		// факт — сумма бюджета закрытых лидов менеджера за месяц (через salon_id его дилеров)
		var salonIDs []uuid.UUID
		s.DB.Model(&models.User{}).Where("managed_by = ?", mgrUUID).Pluck("salon_id", &salonIDs)
		if len(salonIDs) > 0 {
			s.DB.Model(&models.Lead{}).
				Where("salon_id IN ? AND status IN ? AND created_at BETWEEN ? AND ?", salonIDs, []string{"sale", "paid"}, firstOfMonth, lastOfMonth).
				Select("COALESCE(SUM(budget), 0)").Scan(&fact)
		}

		kpi := 0
		if planAmount > 0 {
			kpi = int(fact / planAmount * 100)
			if kpi > 150 {
				kpi = 150
			}
			if kpi < 0 {
				kpi = 0
			}
		}
		resp["kpi"] = append(resp["kpi"].([]map[string]interface{}), map[string]interface{}{
			"month": date.Format("2006-01"),
			"kpi":   kpi,
		})
	}
	return resp, nil
}

// GetManagerDealers - дилеры менеджера (реальный процент по факту vs план)
func (s *KPIService) GetManagerDealers(ctx context.Context, userID, managerID string) (map[string]interface{}, error) {
	resp := map[string]interface{}{"dealers": []map[string]interface{}{}}

	mgrUUID, err := uuid.Parse(managerID)
	if err != nil {
		return resp, err
	}

	var dealers []models.User
	if err := s.DB.Where("role = ? AND managed_by = ?", models.RoleDealer, mgrUUID).Limit(100).Find(&dealers).Error; err != nil {
		return nil, err
	}

	defaultPlan := s.getSettingFloat("default_monthly_plan", 4000000)
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

	for _, d := range dealers {
		// план дилера — из goals, fallback на default
		var planAmount float64
		s.DB.Model(&models.Goal{}).
			Where("assignee_id = ? AND target_date BETWEEN ? AND ?", d.ID, firstOfMonth, lastOfMonth).
			Select("COALESCE(SUM(sales_plan), 0)").Scan(&planAmount)
		if planAmount == 0 {
			planAmount = defaultPlan
		}

		var fact float64
		if d.SalonID != nil {
			s.DB.Model(&models.Lead{}).
				Where("salon_id = ? AND status IN ? AND created_at BETWEEN ? AND ?", d.SalonID, []string{"sale", "paid"}, firstOfMonth, lastOfMonth).
				Select("COALESCE(SUM(budget), 0)").Scan(&fact)
		}

		percent := 0
		if planAmount > 0 {
			percent = int(fact / planAmount * 100)
			if percent < 0 {
				percent = 0
			}
			if percent > 150 {
				percent = 150
			}
		}
		status := "yellow"
		if percent >= 80 {
			status = "green"
		} else if percent < 50 {
			status = "red"
		}

		resp["dealers"] = append(resp["dealers"].([]map[string]interface{}), map[string]interface{}{
			"id":      d.ID,
			"name":    d.FirstName + " " + d.LastName,
			"plan":    planAmount,
			"percent": percent,
			"status":  status,
		})
	}
	return resp, nil
}

// SetManagerPlans - установить планы менеджеров (пишет в goals period='quarter')
func (s *KPIService) SetManagerPlans(ctx context.Context, userID string, quarter string, plans []struct {
	ManagerID     string  `json:"manager_id"`
	PlanAmount    float64 `json:"plan_amount"`
	TargetDealers int     `json:"target_dealers"`
}) error {
	franchiserID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	qStart, qEnd, err := parseQuarter(quarter)
	if err != nil {
		return err
	}
	var franchiser models.User
	if err := s.DB.First(&franchiser, franchiserID).Error; err != nil {
		return err
	}
	for _, p := range plans {
		mgrID, err := uuid.Parse(p.ManagerID)
		if err != nil {
			continue
		}
		// IDOR: менеджер должен быть под франчайзером
		var cnt int64
		s.DB.Model(&models.User{}).Where("id = ? AND managed_by = ?", mgrID, franchiserID).Count(&cnt)
		if cnt == 0 {
			continue
		}
		if p.PlanAmount < 0 {
			continue
		}
		// upsert: delete existing for this assignee+quarter then insert
		s.DB.Where("assignee_id = ? AND period = ? AND start_date = ?", mgrID, "quarter", qStart).Delete(&models.Goal{})
		goal := models.Goal{
			AssignerID:   franchiserID,
			AssigneeID:   mgrID,
			Role:         string(models.RoleFranchisorManager),
			SalesPlan:    p.PlanAmount,
			LeadsPlan:    p.TargetDealers,
			Period:       "quarter",
			StartDate:    qStart,
			EndDate:      qEnd,
			TargetDate:   qStart,
			TenantID:     franchiser.TenantID,
		}
		if err := s.DB.Create(&goal).Error; err != nil {
			return err
		}
	}
	return nil
}

func parseQuarter(q string) (time.Time, time.Time, error) {
	// format 2026-Q2, 2026-Q1..Q4
	parts := strings.Split(q, "-Q")
	if len(parts) != 2 {
		return time.Time{}, time.Time{}, errors.New("invalid quarter format, expected YYYY-QN")
	}
	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	qn, err := strconv.Atoi(parts[1])
	if err != nil || qn < 1 || qn > 4 {
		return time.Time{}, time.Time{}, errors.New("quarter must be 1..4")
	}
	month := (qn-1)*3 + 1
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 3, 0).Add(-time.Nanosecond)
	return start, end, nil
}

// GetManagerPlans - получить планы менеджеров
func (s *KPIService) GetManagerPlans(ctx context.Context, userID string, quarter string) ([]map[string]interface{}, error) {
	franchiserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	qStart, _, err := parseQuarter(quarter)
	if err != nil {
		// fallback: вернуть пусто для невалидного квартала
		return []map[string]interface{}{}, nil
	}
	var goals []models.Goal
	if err := s.DB.Where("assigner_id = ? AND period = ? AND start_date = ?", franchiserID, "quarter", qStart).Limit(100).Find(&goals).Error; err != nil {
		return nil, err
	}
	// батч имена менеджеров
	mgrIDs := make([]uuid.UUID, 0, len(goals))
	for _, g := range goals {
		mgrIDs = append(mgrIDs, g.AssigneeID)
	}
	nameMap := map[uuid.UUID]string{}
	if len(mgrIDs) > 0 {
		var users []models.User
		s.DB.Where("id IN ?", mgrIDs).Find(&users)
		for _, u := range users {
			nameMap[u.ID] = strings.TrimSpace(u.FirstName + " " + u.LastName)
		}
	}
	res := make([]map[string]interface{}, 0, len(goals))
	for _, g := range goals {
		mid := g.AssigneeID.String()
		mname := nameMap[g.AssigneeID]
		res = append(res, map[string]interface{}{
			"manager_id":     mid,
			"manager_name":   mname,
			"plan_amount":    g.SalesPlan,
			"target_dealers": g.LeadsPlan,
			"quarter":        quarter,
			"start_date":     g.StartDate,
			"end_date":       g.EndDate,
		})
	}
	return res, nil
}

func getPeriodBounds(period string) (time.Time, time.Time) {
	now := time.Now().UTC()
	switch period {
	case "month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
		return start, end
	case "year":
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(now.Year(), 12, 31, 23, 59, 59, 999999999, time.UTC)
		return start, end
	default: // quarter
		q := (int(now.Month())-1)/3 + 1
		month := (q-1)*3 + 1
		start := time.Date(now.Year(), time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 3, 0).Add(-time.Nanosecond)
		return start, end
	}
}

// GetDealersHealth - сегментация дилеров ABCD A≥100 B80-99 C50-79 D<50
func (s *KPIService) GetDealersHealth(ctx context.Context, userID string, period string) (map[string]interface{}, error) {
	start, end := getPeriodBounds(period)
	// франчайзер -> все дилеры (role dealer) limit 100
	var dealers []models.User
	if err := s.DB.Where("role = ?", models.RoleDealer).Limit(100).Find(&dealers).Error; err != nil {
		return nil, err
	}
	segA := []map[string]interface{}{}
	segB := []map[string]interface{}{}
	segC := []map[string]interface{}{}
	segD := []map[string]interface{}{}
	defaultPlan := s.getSettingFloat("default_monthly_plan", 4000000)
	for _, d := range dealers {
		var plan float64
		s.DB.Model(&models.Goal{}).Where("assignee_id = ? AND period = ? AND start_date = ?", d.ID, period, start).Select("COALESCE(SUM(sales_plan),0)").Scan(&plan)
		if plan == 0 {
			// fallback: любой план за период или дефолт
			s.DB.Model(&models.Goal{}).Where("assignee_id = ? AND period IN ('quarter','month','year')", d.ID).Select("COALESCE(SUM(sales_plan),0)").Scan(&plan)
			if plan == 0 {
				plan = defaultPlan
			}
		}
		var fact float64
		if d.SalonID != nil {
			s.DB.Model(&models.Lead{}).Where("salon_id = ? AND status IN ? AND created_at BETWEEN ? AND ?", d.SalonID, []string{"sale", "paid"}, start, end).Select("COALESCE(SUM(budget),0)").Scan(&fact)
		}
		percent := 0
		if plan > 0 {
			percent = int(fact / plan * 100)
		}
		entry := map[string]interface{}{"id": d.ID.String(), "name": strings.TrimSpace(d.FirstName + " " + d.LastName), "plan": plan, "fact": fact, "percent": percent}
		switch {
		case percent >= 100:
			segA = append(segA, entry)
		case percent >= 80:
			segB = append(segB, entry)
		case percent >= 50:
			segC = append(segC, entry)
		default:
			segD = append(segD, entry)
		}
	}
	return map[string]interface{}{"segments": map[string]interface{}{"a": segA, "b": segB, "c": segC, "d": segD}}, nil
}

// GetDealersMigration - миграция дилеров между сегментами за 2 периода
func (s *KPIService) GetDealersMigration(ctx context.Context, userID string, period string) (map[string]interface{}, error) {
	// текущая сегментация
	curr, err := s.GetDealersHealth(ctx, userID, period)
	if err != nil {
		return nil, err
	}
	// предыдущий период
	prevPeriodStart, _ := getPeriodBounds(period)
	prevPeriod := period
	// для простоты сдвигаем на 1 квартал/месяц назад через Health с ручным start
	var prevStart, prevEnd time.Time
	switch period {
	case "month":
		prevStart = prevPeriodStart.AddDate(0, -1, 0)
		prevEnd = prevStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	default:
		prevStart = prevPeriodStart.AddDate(0, -3, 0)
		prevEnd = prevStart.AddDate(0, 3, 0).Add(-time.Nanosecond)
	}
	_ = prevPeriod
	_ = prevEnd
	// построить мапу dealer->prev segment
	// упрощенно: повторно вычисляем через ту же логику but с prevStart
	migrations := []map[string]interface{}{}
	// получаем дилеров
	var dealers []models.User
	s.DB.Where("role = ?", models.RoleDealer).Limit(100).Find(&dealers)
	defaultPlan := s.getSettingFloat("default_monthly_plan", 4000000)
	segFor := func(d models.User, start, end time.Time) string {
		var plan float64
		s.DB.Model(&models.Goal{}).Where("assignee_id = ? AND period = ? AND start_date = ?", d.ID, period, start).Select("COALESCE(SUM(sales_plan),0)").Scan(&plan)
		if plan == 0 {
			s.DB.Model(&models.Goal{}).Where("assignee_id = ?", d.ID).Select("COALESCE(SUM(sales_plan),0)").Scan(&plan)
			if plan == 0 {
				plan = defaultPlan
			}
		}
		var fact float64
		if d.SalonID != nil {
			s.DB.Model(&models.Lead{}).Where("salon_id = ? AND status IN ? AND created_at BETWEEN ? AND ?", d.SalonID, []string{"sale", "paid"}, start, end).Select("COALESCE(SUM(budget),0)").Scan(&fact)
		}
		p := 0
		if plan > 0 {
			p = int(fact / plan * 100)
		}
		switch {
		case p >= 100:
			return "a"
		case p >= 80:
			return "b"
		case p >= 50:
			return "c"
		default:
			return "d"
		}
	}
	currSegs := curr["segments"].(map[string]interface{})
	// build prev map
	prevMap := map[string]string{}
	for _, d := range dealers {
		prevMap[d.ID.String()] = segFor(d, prevStart, prevEnd)
	}
	// find migrations where segment changed
	for segKey, list := range currSegs {
		for _, e := range list.([]map[string]interface{}) {
			id := e["id"].(string)
			currSeg := segKey
			prevSeg := prevMap[id]
			if prevSeg != "" && prevSeg != currSeg {
				migrations = append(migrations, map[string]interface{}{"dealer_id": id, "dealer_name": e["name"], "from": prevSeg, "to": currSeg})
			}
		}
	}
	return map[string]interface{}{"migrations": migrations}, nil
}

// GetSystemIssues - системные проблемы (alerts + просроченные заявки/контракты)
func (s *KPIService) GetSystemIssues(ctx context.Context, userID string, status string) ([]map[string]interface{}, error) {
	issues := []map[string]interface{}{}
	// alerts (DB alerts table has status column even if Go model omits it, query via Table)
	type AlertRow struct {
		ID        uuid.UUID `gorm:"column:id"`
		Title     string    `gorm:"column:title"`
		Description string  `gorm:"column:description"`
		Severity  string    `gorm:"column:severity"`
		Status    string    `gorm:"column:status"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}
	var alerts []AlertRow
	q := s.DB.Table("alerts").Limit(50).Order("created_at DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Find(&alerts)
	for _, a := range alerts {
		issues = append(issues, map[string]interface{}{"id": a.ID.String(), "type": "alert", "title": a.Title, "message": a.Description, "severity": a.Severity, "status": a.Status, "created_at": a.CreatedAt})
	}
	// dealer_requests pending
	type Req struct {
		ID          uuid.UUID `gorm:"column:id"`
		Type        string    `gorm:"column:type"`
		Description string    `gorm:"column:description"`
		Status      string    `gorm:"column:status"`
		CreatedAt   time.Time `gorm:"column:created_at"`
	}
	var reqs []Req
	s.DB.Table("dealer_requests").Where("status = ?", "pending").Limit(20).Find(&reqs)
	for _, r := range reqs {
		issues = append(issues, map[string]interface{}{"id": r.ID.String(), "type": "request", "title": "Заявка " + r.Type, "message": r.Description, "severity": "medium", "status": r.Status, "created_at": r.CreatedAt})
	}
	// просроченные контракты
	var overdue []models.Contract
	s.DB.Where("status = ? AND deadline_date < ?", "pending", time.Now()).Limit(20).Find(&overdue)
	for _, c := range overdue {
		issues = append(issues, map[string]interface{}{"id": c.ID.String(), "type": "contract", "title": "Просрочен контракт " + c.ClientName, "severity": "high", "status": c.Status, "deadline": c.DeadlineDate})
	}
	return issues, nil
}

// GetDealersGeography - география дилеров (GROUP BY region/city)
func (s *KPIService) GetDealersGeography(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	type Row struct {
		Region string  `gorm:"column:region"`
		City   string  `gorm:"column:city"`
		Count  int     `gorm:"column:cnt"`
	}
	var rows []Row
	s.DB.Table("salons").Select("COALESCE(region,'Не указан') as region, COALESCE(city,'') as city, COUNT(*) as cnt").Group("region, city").Limit(100).Scan(&rows)
	res := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		res = append(res, map[string]interface{}{"region": r.Region, "city": r.City, "dealers": r.Count})
	}
	// fallback: если region пусто — группируем по address LIKE
	if len(res) == 0 {
		var salons []models.Salon
		s.DB.Limit(100).Find(&salons)
		countByAddr := map[string]int{}
		for _, s := range salons {
			key := "Не указан"
			if s.Address != "" {
				addr := s.Address
				if len(addr) > 20 {
					addr = addr[:20]
				}
				key = addr
			}
			countByAddr[key]++
		}
		for k, v := range countByAddr {
			res = append(res, map[string]interface{}{"region": k, "city": "", "dealers": v})
		}
	}
	return res, nil
}

// GetMarketingROI - ROI маркетинга (gain-cost)/cost*100
func (s *KPIService) GetMarketingROI(ctx context.Context, userID string, period string) (map[string]interface{}, error) {
	start, end := getPeriodBounds(period)
	// cost: marketing_budgets used_amount + dealer_expenses category marketing
	var cost float64
	s.DB.Table("marketing_budgets").Select("COALESCE(SUM(used_amount),0)").Scan(&cost)
	var expCost float64
	s.DB.Table("dealer_expenses").Where("category = ? AND period = ?", "marketing", period).Select("COALESCE(SUM(amount),0)").Scan(&expCost)
	if expCost > 0 {
		cost += expCost
	}
	// gain: сумма бюджетов закрытых лидов за период
	var gain float64
	s.DB.Table("leads").Where("status IN ? AND created_at BETWEEN ? AND ?", []string{"sale", "paid"}, start, end).Select("COALESCE(SUM(budget),0)").Scan(&gain)
	roi := 0.0
	if cost > 0 {
		roi = (gain - cost) / cost * 100
	}
	return map[string]interface{}{"roi": roi, "gain": gain, "cost": cost, "period": period}, nil
}

// GetAlertSettings - настройки алертов
func (s *KPIService) GetAlertSettings(ctx context.Context, userID string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"thresholds": map[string]interface{}{
			"network_forecast_critical": 90,
			"churn_rate_critical":       5,
			"manager_kpi_critical":      70,
		},
		"channels": []string{"in_app", "email"},
	}, nil
}

// UpdateAlertSettings - обновить настройки алертов
func (s *KPIService) UpdateAlertSettings(ctx context.Context, userID string, settings map[string]interface{}) error {
	return nil
}

// GetReportData - данные для отчёта (агрегат реальных блоков)
func (s *KPIService) GetReportData(ctx context.Context, userID string, period, date string) (map[string]interface{}, error) {
	// reuse health/geography/roi для блоков
	health, _ := s.GetDealersHealth(ctx, userID, period)
	geo, _ := s.GetDealersGeography(ctx, userID)
	roi, _ := s.GetMarketingROI(ctx, userID, period)
	issues, _ := s.GetSystemIssues(ctx, userID, "")
	// plan_fact
	start, end := getPeriodBounds(period)
	var totalPlan, totalFact float64
	s.DB.Table("goals").Where("period = ? AND start_date = ?", period, start).Select("COALESCE(SUM(sales_plan),0)").Scan(&totalPlan)
	s.DB.Table("leads").Where("status IN ? AND created_at BETWEEN ? AND ?", []string{"sale", "paid"}, start, end).Select("COALESCE(SUM(budget),0)").Scan(&totalFact)
	// network growth: tenants count
	var tenantCount int64
	s.DB.Table("tenants").Count(&tenantCount)
	return map[string]interface{}{
		"executive_summary": map[string]interface{}{"period": period, "date": date, "total_plan": totalPlan, "total_fact": totalFact, "tenants": tenantCount},
		"plan_fact_dynamics": map[string]interface{}{"total_plan": totalPlan, "total_fact": totalFact, "percent": func() int { if totalPlan > 0 { return int(totalFact / totalPlan * 100) } else { return 0 } }()},
		"network_growth":     map[string]interface{}{"tenants": tenantCount, "health": health},
		"sales_structure":    map[string]interface{}{"geography": geo, "roi": roi},
		"territory_rating":   geo,
		"risks":              issues,
	}, nil
}

// GeneratePDF - сгенерировать PDF (создает запись в reports)
func (s *KPIService) GeneratePDF(ctx context.Context, userID string, blocks []string, comment string) (map[string]interface{}, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	// сохраняем отчет в reports
	blocksJSON, _ := json.Marshal(blocks)
	// recipients пусто на генерации
	id := uuid.New()
	pdfURL := fmt.Sprintf("/reports/%s.pdf", id.String())
	// idempotent insert via Exec
	s.DB.Exec(`INSERT INTO reports (id, franchiser_id, pdf_url, blocks, comment) VALUES (?, ?, ?, ?::jsonb, ?) ON CONFLICT (id) DO NOTHING`, id, uid, pdfURL, string(blocksJSON), comment)
	// fallback for older schema (reports created with only pdf_url/recipients) — second insert no-op if first succeeded
	s.DB.Exec(`INSERT INTO reports (id, franchiser_id, pdf_url) VALUES (?, ?, ?) ON CONFLICT (id) DO NOTHING`, id, uid, pdfURL)
	return map[string]interface{}{"pdf_url": pdfURL, "report_id": id.String()}, nil
}

// SendReport - отправить отчёт (обновляет recipients)
func (s *KPIService) SendReport(ctx context.Context, userID string, reportID string, recipients []string) error {
	rid, err := uuid.Parse(reportID)
	if err != nil {
		return err
	}
	recJSON, _ := json.Marshal(recipients)
	s.DB.Exec(`UPDATE reports SET recipients = ?::jsonb, updated_at = NOW() WHERE id = ?`, string(recJSON), rid)
	return nil
}

// GetReportHistory - история отчётов Lim 20
func (s *KPIService) GetReportHistory(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	type Row struct {
		ID        uuid.UUID `gorm:"column:id"`
		PdfURL    string    `gorm:"column:pdf_url"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}
	var rows []Row
	s.DB.Table("reports").Where("franchiser_id = ?", uid).Order("created_at DESC").Limit(20).Find(&rows)
	res := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		res = append(res, map[string]interface{}{"id": r.ID.String(), "pdf_url": r.PdfURL, "created_at": r.CreatedAt})
	}
	return res, nil
}

// SaveDraft - сохранить черновик (UPSERT report_drafts)
func (s *KPIService) SaveDraft(ctx context.Context, userID string, draft map[string]interface{}) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	dataJSON, _ := json.Marshal(draft)
	// limit 1MB
	if len(dataJSON) > 1024*1024 {
		return errors.New("draft too large")
	}
	s.DB.Exec(`INSERT INTO report_drafts (user_id, data) VALUES (?, ?::jsonb) ON CONFLICT (user_id) DO UPDATE SET data = EXCLUDED.data, updated_at = NOW()`, uid, string(dataJSON))
	return nil
}

// GetDraft - получить черновик
func (s *KPIService) GetDraft(ctx context.Context, userID string) (map[string]interface{}, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	type Row struct {
		Data string `gorm:"column:data"`
	}
	var row Row
	if err := s.DB.Table("report_drafts").Where("user_id = ?", uid).Select("data").Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.Data == "" {
		return map[string]interface{}{}, nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(row.Data), &out); err != nil {
		return map[string]interface{}{"raw": row.Data}, nil
	}
	return out, nil
}
