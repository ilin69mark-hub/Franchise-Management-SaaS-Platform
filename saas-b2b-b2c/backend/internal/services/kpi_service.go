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

// GetDashboardMain - получение данных для главной вкладки дашборда менеджера салона
func (s *KPIService) GetDashboardMain(ctx context.Context, userID uuid.UUID, dateStr string) (*models.DashboardMainResponse, error) {
	// Парсим дату или берем текущую
	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = parsed
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

	// План на месяц (сумма daily_goals)
	var monthPlan float64
	s.DB.Model(&models.DailyGoal{}).
		Where("salon_id = ? AND target_date BETWEEN ? AND ?", salonID, firstOfMonth, lastOfMonth).
		Select("COALESCE(SUM(sales_plan), 0)").Scan(&monthPlan)

	// Факт на текущий момент (сумма продаж из leads со статусом sale)
	var monthFact float64
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND status = ? AND created_at BETWEEN ? AND ?", salonID, "sale", firstOfMonth, targetDate).
		Select("COALESCE(SUM(budget), 0)").Scan(&monthFact)

	resp.Plan = monthPlan
	resp.Fact = monthFact
	if monthPlan > 0 {
		resp.PlanPercent = int((monthFact / monthPlan) * 100)
		if resp.PlanPercent > 100 {
			resp.PlanPercent = 100
		}
	}

	// ====================
	// 2. ДИНАМИКА (сравнение с прошлым днем и неделей)
	// ====================
	// Вчера
	yesterday := targetDate.AddDate(0, 0, -1)
	var yesterdaySales float64
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND status = ? AND DATE(created_at) = ?", salonID, "sale", yesterday.Format("2006-01-02")).
		Select("COALESCE(SUM(budget), 0)").Scan(&yesterdaySales)

	// Прошлая неделя (7 дней назад)
	weekAgo := targetDate.AddDate(0, 0, -7)
	var weekAgoSales float64
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND status = ? AND DATE(created_at) = ?", salonID, "sale", weekAgo.Format("2006-01-02")).
		Select("COALESCE(SUM(budget), 0)").Scan(&weekAgoSales)

	if yesterdaySales > 0 {
		resp.DynamicDay = ((monthFact - yesterdaySales) / yesterdaySales) * 100
	}
	if weekAgoSales > 0 {
		resp.DynamicWeek = ((monthFact - weekAgoSales) / weekAgoSales) * 100
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
		Where("salon_id = ? AND DATE(created_at) = ?", salonID, today).
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
	s.DB.Where("salon_id = ? AND payment_status IN ? AND DATE(payment_date) = ?", salonID, []string{"awaiting_payment", "payment_due"}, today).
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
			targetDate = parsed
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
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int64
			var sum float64
			rows.Scan(&status, &count, &sum)
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
	var prevCount int64 = trafficCount
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
	addStage("payment", "Оплата", 0, 0)

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
	today := targetDate.Format("2006-01-02")
	var todayLeads []models.Lead
	s.DB.Where("salon_id = ? AND DATE(created_at) = ?", salonID, today).
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
			targetDate = parsed
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
		Period:         period,
		TotalRevenue:   0,
		AvgRevenue:     0,
		AvgConversion:  0,
		AvgCheck:       0,
		SalesReps:      []models.SalesRepMetrics{},
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
	s.DB.Where("salon_id = ? AND role = ?", salonID, models.RoleDealerManager).Find(&managers)

	if len(managers) == 0 {
		return resp, nil
	}

	var totalRevenue, totalDeals, totalConversion, totalAvgCheck float64
	var dealsCount, conversionCount int

	for _, mgr := range managers {
		var revenue float64
		s.DB.Model(&models.Lead{}).
			Where("manager_id = ? AND status IN ? AND created_at BETWEEN ? AND ?", mgr.ID, []string{"sale", "paid"}, startDate, endDate).
			Select("COALESCE(SUM(budget), 0)").Scan(&revenue)

		var deals int64
		s.DB.Model(&models.Lead{}).
			Where("manager_id = ? AND status IN ? AND created_at BETWEEN ? AND ?", mgr.ID, []string{"sale", "paid"}, startDate, endDate).
			Count(&deals)

		var leadsCount int64
		s.DB.Model(&models.Lead{}).
			Where("manager_id = ? AND created_at BETWEEN ? AND ?", mgr.ID, startDate, endDate).
			Count(&leadsCount)

		conversion := 0.0
		if leadsCount > 0 {
			conversion = (float64(deals) / float64(leadsCount)) * 100
		}

		avgCheck := 0.0
		if deals > 0 {
			avgCheck = revenue / float64(deals)
		}

		extrasSum := revenue * 0.1
		discountPercent := 5.0

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
			UserID:           mgr.ID,
			FirstName:        mgr.FirstName,
			LastName:         mgr.LastName,
			Role:             string(mgr.Role),
			Revenue:          revenue,
			DealsCount:       int(deals),
			Conversion:       conversion,
			AvgCheck:         avgCheck,
			DiscountPercent:  discountPercent,
			ExtrasSum:        extrasSum,
			RevenueDeviation: 0,
			DealsDeviation:   0,
			ConversionDeviation: 0,
			AvgCheckDeviation: 0,
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
			targetDate = parsed
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

	// Общая выручка салона
	var totalRevenue float64
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND status IN ? AND created_at BETWEEN ? AND ?", salonID, []string{"sale", "paid"}, firstOfMonth, endOfMonth).
		Select("COALESCE(SUM(budget), 0)").Scan(&totalRevenue)
	resp.TotalRevenue = totalRevenue

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
		defer rows.Close()
		for rows.Next() {
			var ps productStats
			rows.Scan(&ps.Name, &ps.Revenue, &ps.Quantity)
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
		resp.TopProducts = append(resp.TopProducts, models.TopProduct{
			ID:           ps.Name,
			Name:         ps.Name,
			Collection:   "Основная",
			Category:     "Мебель",
			Revenue:      ps.Revenue,
			Quantity:     ps.Quantity,
			SharePercent: share,
			Margin:       25.0, // Заглушка
		})
	}

	// === ОСТАТКИ (заглушка - нужна таблица products/stock) ===
	// Симулируем остатки для демонстрации
	sampleProducts := []string{"Диван", "Кресло", "Кровать", "Шкаф", "Стол"}
	for i, name := range sampleProducts {
		showroom := 1 + i
		warehouse := 5 + i*2
		cost := float64((showroom + warehouse) * 50000)
		turnover := 30 + i*10
		if turnover > 90 {
			turnover = 120 // Неликвид
		}
		resp.StockItems = append(resp.StockItems, models.StockItem{
			ID:           name,
			Name:         name,
			Category:     "Мебель",
			ShowroomQty:  showroom,
			WarehouseQty: warehouse,
			TotalCost:    cost,
			TurnoverDays: turnover,
		})
	}

	// === УПУЩЕННЫЕ ПРОДАЖИ ===
	lostReasons := map[string]float64{
		"Нет в наличии":           150000,
		"Долгий срок производства": 80000,
		"Не устроила цена":        200000,
		"Не подошёл дизайн":       120000,
	}
	for reason, revenue := range lostReasons {
		count := int(revenue / 50000) // Примерное количество
		resp.LostSales = append(resp.LostSales, models.LostSale{
			Reason:        reason,
			RequestsCount: count,
			LostRevenue:   revenue,
		})
	}

	// === ОБОРАЧИВАЕМОСТЬ ПО КАТЕГОРИЯМ ===
	categories := []string{"Диваны", "Кресла", "Кровати", "Шкафы", "Столы"}
	for _, cat := range categories {
		days := 30 + (len(cat) * 5) // Заглушка
		slowMoving := days > 90
		resp.CategoryTurnover = append(resp.CategoryTurnover, models.CategoryTurnover{
			Category:    cat,
			AvgDays:     days,
			IsSlowMoving: slowMoving,
		})
	}

	return resp, nil
}

// GetManagerTargets - получить директивы от дилера
func (s *KPIService) GetManagerTargets(ctx context.Context, userID uuid.UUID, dateStr string) (*models.ManagerTargetsResponse, error) {
	resp := &models.ManagerTargetsResponse{
		HasTargets:          false,
		TargetConversion:     0,
		CurrentConversion:    0,
		TargetExtrasPercent:  0,
		CurrentExtrasPercent: 0,
		Promotions:          []models.Promotion{},
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
	firstOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())

	var planAmount float64
	s.DB.Model(&models.DailyGoal{}).
		Where("salon_id = ? AND target_date BETWEEN ? AND ?", salonID, firstOfMonth, time.Now()).
		Select("COALESCE(SUM(sales_plan), 0)").Scan(&planAmount)

	// Если план не утверждён (не задан)
	if planAmount == 0 {
		resp.HasTargets = false
		resp.Plan = nil
		return resp, nil
	}

	resp.HasTargets = true

	// Текущее выполнение
	var currentAmount float64
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND status IN ? AND created_at BETWEEN ? AND ?", salonID, []string{"sale", "paid"}, firstOfMonth, time.Now()).
		Select("COALESCE(SUM(budget), 0)").Scan(&currentAmount)

	percent := 0
	if planAmount > 0 {
		percent = int((currentAmount / planAmount) * 100)
		if percent > 100 {
			percent = 100
		}
	}

	byCategory := map[string]float64{
		"Мебель": currentAmount * 0.7,
		"Допы":   currentAmount * 0.2,
		"Услуги": currentAmount * 0.1,
	}

	resp.Plan = &models.TargetPlan{
		TotalAmount:    planAmount,
		ByCategory:    byCategory,
		CurrentAmount:  currentAmount,
		Percent:        percent,
	}

	// === Бенчмарки ===
	resp.TargetConversion = 30.0 // 30% - цель
	resp.TargetExtrasPercent = 15.0 // 15% - цель

	// Текущая конверсия
	var totalLeads, saleLeads int64
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND created_at BETWEEN ? AND ?", salonID, firstOfMonth, time.Now()).
		Count(&totalLeads)
	s.DB.Model(&models.Lead{}).
		Where("salon_id = ? AND status IN ? AND created_at BETWEEN ? AND ?", salonID, []string{"sale", "paid"}, firstOfMonth, time.Now()).
		Count(&saleLeads)

	if totalLeads > 0 {
		resp.CurrentConversion = (float64(saleLeads) / float64(totalLeads)) * 100
	}
	resp.CurrentExtrasPercent = 12.0 // Заглушка

	// === Акции ===
	now := time.Now()
	promotions := []models.Promotion{
		{
			ID:            "1",
			Name:          "Летняя распродажа",
			Condition:     "При покупке дивана - кресло в подарок",
			DiscountMin:   10,
			DiscountMax:   25,
			EndDate:       now.AddDate(0, 0, 5).Format("2006-01-02"),
			IsExpiring:    true,
		},
		{
			ID:            "2",
			Name:          "Комплект со скидкой",
			Condition:     "Мебель + услуги дизайнера",
			DiscountMin:   15,
			DiscountMax:   30,
			EndDate:       now.AddDate(0, 0, 14).Format("2006-01-02"),
			IsExpiring:    false,
		},
		{
			ID:            "3",
			Name:          "Акция выходного дня",
			Condition:     "Скидка 20% в субботу и воскресенье",
			DiscountMin:   20,
			DiscountMax:   20,
			EndDate:       now.AddDate(0, 0, 3).Format("2006-01-02"),
			IsExpiring:    true,
		},
	}
	resp.Promotions = promotions

	// === Прогноз премии ===
	maxBonus := 50000.0 // Максимальная премия
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

	// Уровень предупреждения
	bonusPercent := (resp.BonusForecast / maxBonus) * 100
	if bonusPercent < 50 {
		resp.WarningLevel = "yellow"
	} else if bonusPercent < 30 {
		resp.WarningLevel = "red"
	} else {
		resp.WarningLevel = "none"
	}

	return resp, nil
}
