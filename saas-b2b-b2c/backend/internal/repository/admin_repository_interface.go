package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"franchise-saas-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminRepositoryInterface interface {
	GetDashboardStats(ctx context.Context) (map[string]interface{}, error)
	GetSystemStats(ctx context.Context) (map[string]interface{}, error)
	GetTenantsWithPaidUntil(ctx context.Context) ([]map[string]interface{}, error)
	GetTenantsPaymentStatus(ctx context.Context) ([]map[string]interface{}, error)
	GetAllPlans(ctx context.Context) ([]models.Plan, error)
	CreatePlan(ctx context.Context, name string, price float64, maxUsers int) (*models.Plan, error)
	UpdatePlan(ctx context.Context, id uuid.UUID, name string, price float64, maxUsers int) (*models.Plan, error)
	DeletePlan(ctx context.Context, id uuid.UUID) error
	CreateInvoice(ctx context.Context, tenantID uuid.UUID, amount float64, description string, dueDate time.Time) (*models.Invoice, error)
	GetAllInvoices(ctx context.Context) ([]map[string]interface{}, error)
	MarkInvoicePaid(ctx context.Context, invoiceID uuid.UUID) error
	GetAnalyticsData(ctx context.Context) (map[string]interface{}, error)
	GetProductAnalytics(ctx context.Context) (map[string]interface{}, error)
	GetRisksAnalytics(ctx context.Context) ([]map[string]interface{}, error)
	GetUnitEconomics(ctx context.Context) (map[string]interface{}, error)
	UpdateMarketingSpend(ctx context.Context, amount float64) error
	GetTenantByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	UpdateTenant(ctx context.Context, id uuid.UUID, name string, planID *uuid.UUID, paidUntil *time.Time) error
	BlockTenant(ctx context.Context, id uuid.UUID) error
	UnblockTenant(ctx context.Context, id uuid.UUID) error
}

type AdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) GetDashboardStats(ctx context.Context) (map[string]interface{}, error) {
	var tenantCount int64
	var userCount int64
	var activeTenants int64
	var churnedTenants int64
	var newTenantsMonth int64
	var totalRevenue float64

	r.db.Model(&models.Tenant{}).Count(&tenantCount)
	r.db.Model(&models.User{}).Count(&userCount)

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	r.db.Model(&models.Invoice{}).
		Where("status = ? AND paid_at >= ?", "paid", startOfMonth).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalRevenue)

	r.db.Model(&models.Tenant{}).Where("status = ?", "active").Count(&activeTenants)
	r.db.Model(&models.Tenant{}).Where("status = ?", "churned").Count(&churnedTenants)
	r.db.Model(&models.Tenant{}).Where("created_at >= ?", startOfMonth).Count(&newTenantsMonth)

	var arpu float64
	if activeTenants > 0 {
		arpu = totalRevenue / float64(activeTenants)
	}

	type PlanDist struct {
		Name  string
		Count int
	}
	var planDistribution []PlanDist
	r.db.Table("tenants").
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

func (r *AdminRepository) GetSystemStats(ctx context.Context) (map[string]interface{}, error) {
	var tenants []models.Tenant
	if err := r.db.Find(&tenants).Error; err != nil {
		return nil, err
	}

	var totalUsers int64
	r.db.Model(&models.User{}).Count(&totalUsers)

	activeTenants := 0
	for _, t := range tenants {
		if t.Status == "active" {
			activeTenants++
		}
	}

	return map[string]interface{}{
		"total_tenants":  len(tenants),
		"active_tenants": activeTenants,
		"total_users":    totalUsers,
	}, nil
}

func (r *AdminRepository) GetTenantsWithPaidUntil(ctx context.Context) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Table("tenants").
		Select("id, paid_until").
		Where("paid_until IS NOT NULL").
		Find(&results).Error
	return results, err
}

func (r *AdminRepository) GetTenantsPaymentStatus(ctx context.Context) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Table("tenants").
		Select("tenants.id, tenants.name, tenants.status, tenants.plan_id, tenants.paid_until, plans.name as plan_name").
		Joins("LEFT JOIN plans ON tenants.plan_id = plans.id").
		Find(&results).Error
	return results, err
}

func (r *AdminRepository) GetAllPlans(ctx context.Context) ([]models.Plan, error) {
	var plans []models.Plan
	err := r.db.Find(&plans).Error
	return plans, err
}

func (r *AdminRepository) CreatePlan(ctx context.Context, name string, price float64, maxUsers int) (*models.Plan, error) {
	plan := models.Plan{
		Name:     name,
		Price:    price,
		MaxUsers: maxUsers,
	}
	if err := r.db.Create(&plan).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *AdminRepository) UpdatePlan(ctx context.Context, id uuid.UUID, name string, price float64, maxUsers int) (*models.Plan, error) {
	var plan models.Plan
	if err := r.db.First(&plan, id).Error; err != nil {
		return nil, err
	}

	plan.Name = name
	plan.Price = price
	plan.MaxUsers = maxUsers

	if err := r.db.Save(&plan).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *AdminRepository) DeletePlan(ctx context.Context, id uuid.UUID) error {
	return r.db.Delete(&models.Plan{}, id).Error
}

func (r *AdminRepository) CreateInvoice(ctx context.Context, tenantID uuid.UUID, amount float64, description string, dueDate time.Time) (*models.Invoice, error) {
	inv := models.Invoice{
		TenantID:    tenantID,
		Amount:      amount,
		Description: description,
		DueDate:     dueDate,
		Status:      "pending",
	}
	if err := r.db.Create(&inv).Error; err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *AdminRepository) GetAllInvoices(ctx context.Context) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := r.db.Table("invoices").
		Select("invoices.id, invoices.tenant_id, invoices.amount, invoices.status, invoices.due_date, invoices.paid_at, tenants.name as tenant_name").
		Joins("LEFT JOIN tenants ON invoices.tenant_id = tenants.id").
		Order("invoices.created_at DESC").
		Find(&results).Error
	return results, err
}

func (r *AdminRepository) MarkInvoicePaid(ctx context.Context, invoiceID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.Invoice{}).Where("id = ?", invoiceID).Updates(map[string]interface{}{
		"status":  "paid",
		"paid_at": now,
	}).Error
}

func (r *AdminRepository) GetAnalyticsData(ctx context.Context) (map[string]interface{}, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	var dau, mau int64
	r.db.Model(&models.UserLog{}).
		Select("COUNT(DISTINCT user_id)").
		Where("created_at >= ?", startOfDay).
		Scan(&dau)

	thirtyDaysAgo := now.AddDate(0, 0, -30)
	r.db.Model(&models.UserLog{}).
		Select("COUNT(DISTINCT user_id)").
		Where("created_at >= ?", thirtyDaysAgo).
		Scan(&mau)

	sticky := 0.0
	if mau > 0 {
		sticky = float64(dau) / float64(mau) * 100
	}

	var regs []map[string]interface{}
	r.db.Table("users").
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ?", time.Now().AddDate(0, 0, -7)).
		Group("DATE(created_at)").
		Order("date asc").
		Find(&regs)

	var total, trial, paid int64
	r.db.Model(&models.Tenant{}).Count(&total)
	r.db.Model(&models.Tenant{}).Where("trial_ends_at > ?", now).Count(&trial)
	r.db.Model(&models.Tenant{}).Where("status = ?", "active").Count(&paid)

	conv := 0.0
	if total > 0 {
		conv = float64(paid) / float64(total) * 100
	}

	return map[string]interface{}{
		"dau":           dau,
		"mau":           mau,
		"sticky_factor": sticky,
		"registrations": regs,
		"funnel": map[string]interface{}{
			"total_tenants": total,
			"trial_active":  trial,
			"paid_active":  paid,
			"conversion":   conv,
		},
	}, nil
}

func (r *AdminRepository) GetProductAnalytics(ctx context.Context) (map[string]interface{}, error) {
	var totalTasks, completedTasks int64
	r.db.Model(&models.Checklist{}).Count(&totalTasks)
	r.db.Model(&models.Checklist{}).Where("status = ?", "completed").Count(&completedTasks)

	kpiPercent := 0.0
	if totalTasks > 0 {
		kpiPercent = float64(completedTasks) / float64(totalTasks) * 100
	}

	var activeTrials int64
	now := time.Now()
	r.db.Model(&models.Tenant{}).Where("trial_ends_at > ? AND status = ?", now, "active").Count(&activeTrials)

	var totalTenants, activeTenants int64
	r.db.Model(&models.Tenant{}).Count(&totalTenants)
	r.db.Model(&models.Tenant{}).Where("status = ?", "active").Count(&activeTenants)

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

func (r *AdminRepository) GetRisksAnalytics(ctx context.Context) ([]map[string]interface{}, error) {
	var tenants []models.Tenant
	if err := r.db.Where("status != ?", "churned").Find(&tenants).Error; err != nil {
		return nil, err
	}

	var risks []map[string]interface{}
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	for _, t := range tenants {
		score := 100
		reasons := []string{}

		if t.PaidUntil == nil || t.PaidUntil.Before(now) {
			score -= 40
			reasons = append(reasons, "Просрочена оплата")
		}

		var lastLog models.UserLog
		r.db.Where("tenant_id = ?", t.ID).Order("created_at desc").First(&lastLog)

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

		var total, completed int64
		r.db.Model(&models.Checklist{}).Where("tenant_id = ?", t.ID).Count(&total)
		if total > 0 {
			r.db.Model(&models.Checklist{}).Where("tenant_id = ? AND status = ?", t.ID, "completed").Count(&completed)
			rate := float64(completed) / float64(total)
			if rate < 0.5 {
				score -= 20
				reasons = append(reasons, "Низкое выполнение KPI")
			}
		}

		if len(reasons) > 0 || t.Status == "blocked" {
			risks = append(risks, map[string]interface{}{
				"id":            t.ID,
				"name":          t.Name,
				"status":        t.Status,
				"health_score":  score,
				"risk_reason":   strings.Join(reasons, "; "),
				"last_activity": lastActStr,
			})
		}
	}

	return risks, nil
}

func (r *AdminRepository) GetUnitEconomics(ctx context.Context) (map[string]interface{}, error) {
	var totalRevenue float64
	var totalTenantsEver int64

	r.db.Model(&models.Invoice{}).Where("status = ?", "paid").Select("COALESCE(SUM(amount), 0)").Scan(&totalRevenue)
	r.db.Model(&models.Tenant{}).Unscoped().Count(&totalTenantsEver)

	ltv := 0.0
	if totalTenantsEver > 0 {
		ltv = totalRevenue / float64(totalTenantsEver)
	}

	var marketingSpendStr string
	r.db.Table("system_settings").Where("key = ?", "marketing_spend_current_month").Select("value").Scan(&marketingSpendStr)

	marketingSpend := 0.0
	if marketingSpendStr != "" {
		fmt.Sscanf(marketingSpendStr, "%f", &marketingSpend)
	}

	var newTenantsThisMonth int64
	startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
	r.db.Model(&models.Tenant{}).Where("created_at >= ?", startOfMonth).Count(&newTenantsThisMonth)

	cac := 0.0
	if newTenantsThisMonth > 0 {
		cac = marketingSpend / float64(newTenantsThisMonth)
	}

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

func (r *AdminRepository) UpdateMarketingSpend(ctx context.Context, amount float64) error {
	amountStr := fmt.Sprintf("%.2f", amount)
	return r.db.Exec(`
        INSERT INTO system_settings (key, value, updated_at) 
        VALUES ('marketing_spend_current_month', ?, NOW()) 
        ON CONFLICT (key) DO UPDATE SET value = ?, updated_at = NOW()
    `, amountStr, amountStr).Error
}

func (r *AdminRepository) GetTenantByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	var tenant models.Tenant
	err := r.db.First(&tenant, id).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *AdminRepository) UpdateTenant(ctx context.Context, id uuid.UUID, name string, planID *uuid.UUID, paidUntil *time.Time) error {
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
	return r.db.Model(&models.Tenant{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AdminRepository) BlockTenant(ctx context.Context, id uuid.UUID) error {
	return r.db.Model(&models.Tenant{}).Where("id = ?", id).Update("status", "blocked").Error
}

func (r *AdminRepository) UnblockTenant(ctx context.Context, id uuid.UUID) error {
	return r.db.Model(&models.Tenant{}).Where("id = ?", id).Update("status", "active").Error
}

var _ AdminRepositoryInterface = (*AdminRepository)(nil)