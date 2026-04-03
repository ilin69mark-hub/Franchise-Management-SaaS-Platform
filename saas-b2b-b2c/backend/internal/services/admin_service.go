package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminService struct {
	tenantRepo *repository.TenantRepository
	db         *gorm.DB
}

func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{
		tenantRepo: repository.NewTenantRepository(db),
		db:         db,
	}
}

// GetDashboardStats - статистика
func (s *AdminService) GetDashboardStats() (map[string]interface{}, error) {
	var tenantCount int64
	var userCount int64
	var activeTenants int64
	var churnedTenants int64
	var newTenantsMonth int64
	var totalRevenue float64

	s.db.Model(&models.Tenant{}).Count(&tenantCount)
	s.db.Model(&models.User{}).Count(&userCount)

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	s.db.Model(&models.Invoice{}).
		Where("status = ? AND paid_at >= ?", "paid", startOfMonth).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalRevenue)

	s.db.Model(&models.Tenant{}).Where("status = ?", "active").Count(&activeTenants)
	s.db.Model(&models.Tenant{}).Where("status = ?", "churned").Count(&churnedTenants)
	s.db.Model(&models.Tenant{}).Where("created_at >= ?", startOfMonth).Count(&newTenantsMonth)

	var arpu float64
	if activeTenants > 0 {
		arpu = totalRevenue / float64(activeTenants)
	}

	type PlanDist struct {
		Name  string
		Count int
	}
	var planDistribution []PlanDist
	s.db.Table("tenants").
		Select("plans.name as name, count(tenants.id) as count").
		Joins("left join plans on tenants.plan_id = plans.id").
		Where("tenants.status = ?", "active").
		Group("plans.name").
		Scan(&planDistribution)

	return map[string]interface{}{
		"total_tenants":   tenantCount,
		"active_tenants":  activeTenants,
		"churned_tenants": churnedTenants,
		"total_users":     userCount,
		"mrr":             totalRevenue,
		"arpu":            arpu,
		"new_this_month":  newTenantsMonth,
		"plan_stats":      planDistribution,
	}, nil
}

// GetAllTenants - список всех сетей
func (s *AdminService) GetAllTenants(ctx context.Context) ([]models.Tenant, error) {
	return s.tenantRepo.FindAll(ctx)
}

// GetTenantsPaymentStatus - статус оплат сетей
func (s *AdminService) GetTenantsPaymentStatus() ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := s.db.Table("tenants").
		Select("tenants.id, tenants.name, tenants.status, tenants.plan_id, tenants.paid_until, plans.name as plan_name").
		Joins("LEFT JOIN plans ON tenants.plan_id = plans.id").
		Find(&results).Error
	return results, err
}

// CreateTenant - создать сеть
func (s *AdminService) CreateTenant(name string, planID *uuid.UUID) (*models.Tenant, error) {
	tenant := models.Tenant{
		Name:   name,
		PlanID: planID,
		Status: "active",
	}
	if err := s.db.Create(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

// UpdateTenant - обновить сеть
func (s *AdminService) UpdateTenant(id uuid.UUID, name string, planID *uuid.UUID, paidUntil *time.Time) error {
	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if planID != nil {
		updates["plan_id"] = planID
	}
	if paidUntil != nil {
		updates["paid_until"] = paidUntil
	}
	return s.db.Model(&models.Tenant{}).Where("id = ?", id).Updates(updates).Error
}

// BlockTenant - заблокировать
func (s *AdminService) BlockTenant(id uuid.UUID) error {
	return s.db.Model(&models.Tenant{}).Where("id = ?", id).Update("status", "blocked").Error
}

// DeleteTenant - удалить
func (s *AdminService) DeleteTenant(id uuid.UUID) error {
	s.db.Model(&models.Tenant{}).Where("id = ?", id).Update("status", "churned")
	return s.db.Delete(&models.Tenant{}, id).Error
}

// === PLANS ===

func (s *AdminService) GetAllPlans() ([]models.Plan, error) {
	var plans []models.Plan
	err := s.db.Find(&plans).Error
	return plans, err
}

func (s *AdminService) CreatePlan(name string, price float64, maxUsers int) (*models.Plan, error) {
	plan := models.Plan{
		Name:     name,
		Price:    price,
		MaxUsers: maxUsers,
	}
	if err := s.db.Create(&plan).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

// UpdatePlan - обновить тариф (НОВОЕ)
func (s *AdminService) UpdatePlan(id uuid.UUID, name string, price float64, maxUsers int) (*models.Plan, error) {
	var plan models.Plan
	if err := s.db.First(&plan, id).Error; err != nil {
		return nil, err
	}

	plan.Name = name
	plan.Price = price
	plan.MaxUsers = maxUsers

	if err := s.db.Save(&plan).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

// DeletePlan - удалить тариф (НОВОЕ)
func (s *AdminService) DeletePlan(id uuid.UUID) error {
	return s.db.Delete(&models.Plan{}, id).Error
}

// === INVOICES ===

func (s *AdminService) CreateInvoice(tenantID uuid.UUID, amount float64, description string, dueDate time.Time) (*models.Invoice, error) {
	inv := models.Invoice{
		TenantID:    tenantID,
		Amount:      amount,
		Description: description,
		DueDate:     dueDate,
		Status:      "pending",
	}
	if err := s.db.Create(&inv).Error; err != nil {
		return nil, err
	}
	return &inv, nil
}

func (s *AdminService) MarkInvoicePaid(invoiceID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&models.Invoice{}).Where("id = ?", invoiceID).Updates(map[string]interface{}{
		"status":  "paid",
		"paid_at": now,
	}).Error
}

// === ЭТАП 2: АНАЛИТИКА ===

// AnalyticsResponse структура для фронта
type AnalyticsResponse struct {
	DAU           int64                    `json:"dau"`
	MAU           int64                    `json:"mau"`
	StickyFactor  float64                  `json:"sticky_factor"`
	Registrations []map[string]interface{} `json:"registrations"`
	Funnel        FunnelStats              `json:"funnel"`
}

type FunnelStats struct {
	TotalTenants int64   `json:"total_tenants"`
	TrialActive  int64   `json:"trial_active"`
	PaidActive   int64   `json:"paid_active"`
	Conversion   float64 `json:"conversion"`
}

// GetAnalyticsData собирает метрики активности
func (s *AdminService) GetAnalyticsData() (*AnalyticsResponse, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// 1. DAU / MAU (простые запросы к таблице user_logs)
	// Для простоты делаем прямые запросы через db, можно вынести в repo
	var dau, mau int64
	s.db.Model(&models.UserLog{}).
		Select("COUNT(DISTINCT user_id)").
		Where("created_at >= ?", startOfDay).
		Scan(&dau)

	thirtyDaysAgo := now.AddDate(0, 0, -30)
	s.db.Model(&models.UserLog{}).
		Select("COUNT(DISTINCT user_id)").
		Where("created_at >= ?", thirtyDaysAgo).
		Scan(&mau)

	sticky := 0.0
	if mau > 0 {
		sticky = float64(dau) / float64(mau) * 100
	}

	// 2. Регистрации
	var regs []map[string]interface{}
	s.db.Table("users").
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", time.Now().AddDate(0, 0, -7)).
		Group("DATE(created_at)").
		Order("date asc").
		Find(&regs)

	// 3. Воронка
	var total, trial, paid int64
	s.db.Model(&models.Tenant{}).Count(&total)
	s.db.Model(&models.Tenant{}).Where("trial_ends_at > ?", now).Count(&trial)
	s.db.Model(&models.Tenant{}).Where("status = ?", "active").Count(&paid)

	conv := 0.0
	if total > 0 {
		conv = float64(paid) / float64(total) * 100
	}

	return &AnalyticsResponse{
		DAU:           dau,
		MAU:           mau,
		StickyFactor:  sticky,
		Registrations: regs,
		Funnel: FunnelStats{
			TotalTenants: total,
			TrialActive:  trial,
			PaidActive:   paid,
			Conversion:   conv,
		},
	}, nil
}

// === ЭТАП 3: Продукт и KPI ===

// GetProductAnalytics возвращает метрики продукта
func (s *AdminService) GetProductAnalytics() (map[string]interface{}, error) {
	// 1. KPI по задачам (Checklists)
	var totalTasks, completedTasks int64
	s.db.Model(&models.Checklist{}).Count(&totalTasks)
	s.db.Model(&models.Checklist{}).Where("status = ?", "completed").Count(&completedTasks)

	kpiPercent := 0.0
	if totalTasks > 0 {
		kpiPercent = float64(completedTasks) / float64(totalTasks) * 100
	}

	// 2. Trial Воронка (Используем поля из миграции этапа 2)
	var activeTrials int64
	now := time.Now()
	s.db.Model(&models.Tenant{}).Where("trial_ends_at > ? AND status = ?", now, "active").Count(&activeTrials)

	// 3. Retention (Простая метрика: % активных от общего числа созданных)
	var totalTenants, activeTenants int64
	s.db.Model(&models.Tenant{}).Count(&totalTenants)
	s.db.Model(&models.Tenant{}).Where("status = ?", "active").Count(&activeTenants)

	retention := 0.0
	if totalTenants > 0 {
		retention = float64(activeTenants) / float64(totalTenants) * 100
	}

	return map[string]interface{}{
		"kpi": map[string]interface{}{
			"total_tasks":     totalTasks,
			"completed_tasks": completedTasks,
			"completion_rate": kpiPercent,
		},
		"trials": map[string]interface{}{
			"active_count": activeTrials,
		},
		"product": map[string]interface{}{
			"retention_rate": retention,
		},
	}, nil
}

// === ЭТАП 4: Риски и Health Score ===

type RiskTenant struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	HealthScore  int       `json:"health_score"`
	RiskReason   string    `json:"risk_reason"`
	LastActivity string    `json:"last_activity"`
}

// GetRisksAnalytics возвращает список сетей в зоне риска
func (s *AdminService) GetRisksAnalytics() ([]RiskTenant, error) {
	var tenants []models.Tenant
	// Берем все сети, включая неактивные (soft deleted не берем)
	if err := s.db.Where("status != ?", "churned").Find(&tenants).Error; err != nil {
		return nil, err
	}

	var risks []RiskTenant
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	for _, t := range tenants {
		score := 100
		reasons := []string{}

		// 1. Проверка Финансов (40% веса)
		if t.PaidUntil == nil || t.PaidUntil.Before(now) {
			score -= 40
			reasons = append(reasons, "Просрочена оплата")
		}

		// 2. Проверка Активности (40% веса)
		// Смотрим, были ли логи от юзеров этой сети за последние 30 дней
		var lastLog models.UserLog
		s.db.Where("tenant_id = ?", t.ID).Order("created_at desc").First(&lastLog)

		lastActStr := "Никогда"
		if lastLog.ID != uuid.Nil {
			lastActStr = lastLog.CreatedAt.Format("2006-01-02")
			if lastLog.CreatedAt.Before(thirtyDaysAgo) {
				score -= 40
				reasons = append(reasons, "Нет активности 30 дней")
			}
		} else {
			score -= 40
			reasons = append(reasons, "Нет логов активности")
		}

		// 3. Проверка KPI (20% веса)
		// Если есть задачи, но % выполнения низкий
		var total, completed int64
		s.db.Model(&models.Checklist{}).Where("tenant_id = ?", t.ID).Count(&total)
		if total > 0 {
			s.db.Model(&models.Checklist{}).Where("tenant_id = ? AND status = ?", t.ID, "completed").Count(&completed)
			rate := float64(completed) / float64(total)
			if rate < 0.5 { // Меньше 50% выполнения
				score -= 20
				reasons = append(reasons, "Низкое выполнение KPI")
			}
		}

		// Если есть проблемы, добавляем в список
		if len(reasons) > 0 || t.Status == "blocked" {
			risks = append(risks, RiskTenant{
				ID:           t.ID,
				Name:         t.Name,
				Status:       t.Status,
				HealthScore:  score,
				RiskReason:   strings.Join(reasons, "; "),
				LastActivity: lastActStr,
			})
		}
	}

	// Сортируем: самые "больные" впереди
	// (Простая сортировка пузырьком для малого кол-ва сетей)
	for i := 0; i < len(risks); i++ {
		for j := i + 1; j < len(risks); j++ {
			if risks[i].HealthScore > risks[j].HealthScore {
				risks[i], risks[j] = risks[j], risks[i]
			}
		}
	}

	return risks, nil
}

// === ЭТАП 5: LTV, CAC и Unit Economics ===

// GetUnitEconomics - расчет LTV, CAC и Ratio
func (s *AdminService) GetUnitEconomics() (map[string]interface{}, error) {
	// 1. LTV (Lifetime Value)
	// Считаем средний доход с одной сети за всё время
	// LTV = Общая выручка / Кол-во сетей (когда-либо созданных)
	var totalRevenue float64
	var totalTenantsEver int64

	s.db.Model(&models.Invoice{}).Where("status = ?", "paid").Select("COALESCE(SUM(amount), 0)").Scan(&totalRevenue)
	s.db.Model(&models.Tenant{}).Unscoped().Count(&totalTenantsEver) // Unscoped чтобы считать и удаленных

	ltv := 0.0
	if totalTenantsEver > 0 {
		ltv = totalRevenue / float64(totalTenantsEver)
	}

	// 2. CAC (Customer Acquisition Cost)
	// Расходы на маркетинг / Кол-во НОВЫХ сетей за этот месяц
	// Получаем расходы из system_settings
	var marketingSpendStr string
	s.db.Table("system_settings").Where("key = ?", "marketing_spend_current_month").Select("value").Scan(&marketingSpendStr)

	marketingSpend := 0.0
	if marketingSpendStr != "" {
		fmt.Sscanf(marketingSpendStr, "%f", &marketingSpend)
	}

	var newTenantsThisMonth int64
	startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	s.db.Model(&models.Tenant{}).Where("created_at >= ?", startOfMonth).Count(&newTenantsThisMonth)

	cac := 0.0
	if newTenantsThisMonth > 0 {
		cac = marketingSpend / float64(newTenantsThisMonth)
	}

	// 3. LTV/CAC Ratio (Здоровье бизнеса)
	// > 3 - отлично, < 1 - плохо
	ratio := 0.0
	if cac > 0 {
		ratio = ltv / cac
	}

	return map[string]interface{}{
		"ltv":               ltv,
		"cac":               cac,
		"ltv_cac_ratio":     ratio,
		"marketing_spend":   marketingSpend,
		"new_tenants_month": newTenantsThisMonth,
	}, nil
}

// UpdateMarketingSpend - обновить расходы на маркетинг
func (s *AdminService) UpdateMarketingSpend(amount float64) error {
	amountStr := fmt.Sprintf("%.2f", amount)
	return s.db.Exec(`
        INSERT INTO system_settings (key, value, updated_at) 
        VALUES ('marketing_spend_current_month', ?, NOW()) 
        ON CONFLICT (key) DO UPDATE SET value = ?, updated_at = NOW()
    `, amountStr, amountStr).Error
}
