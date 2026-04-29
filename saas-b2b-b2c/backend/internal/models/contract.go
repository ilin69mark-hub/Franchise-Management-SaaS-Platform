package models

import (
	"time"

	"github.com/google/uuid"
)

// Contract - договор с клиентом на мебель
type Contract struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SalonID     uuid.UUID  `json:"salon_id" gorm:"type:uuid;not null;index"`
	LeadID      *uuid.UUID `json:"lead_id" gorm:"type:uuid"`
	ManagerID   uuid.UUID  `json:"manager_id" gorm:"type:uuid;not null"`

	ClientName  string    `json:"client_name" gorm:"not null"`
	ClientPhone string    `json:"client_phone"`
	ClientEmail string    `json:"client_email"`

	TotalAmount   float64   `json:"total_amount"`   // Общая сумма договора
	PrepaidAmount float64   `json:"prepaid_amount"` // Сумма предоплаты
	PaidAmount    float64   `json:"paid_amount"`    // Уже оплачено
	RemainAmount  float64   `json:"remain_amount"` // Остаток к оплате
	MarginPercent float64   `json:"margin_percent"` // Маржинальность в %

	Status        string    `json:"status"` // pending, partial, paid, cancelled
	PaymentStatus string    `json:"payment_status"` // awaiting_payment, payment_due, paid

	PaymentDate   *time.Time `json:"payment_date"`   // Дата когда нужно оплатить
	DeadlineDate  *time.Time `json:"deadline_date"`  // Срок сдачи

	Products      string    `json:"products"` // Список товаров
	Description   string    `json:"description"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Contract) TableName() string {
	return "contracts"
}

// === DTOs ===

type CreateContractRequest struct {
	LeadID       *uuid.UUID `json:"lead_id"`
	ClientName   string     `json:"client_name" binding:"required"`
	ClientPhone  string     `json:"client_phone"`
	ClientEmail  string     `json:"client_email"`
	TotalAmount  float64    `json:"total_amount" binding:"required"`
	PrepaidAmount float64   `json:"prepaid_amount"`
	Products    string     `json:"products"`
	Description string     `json:"description"`
	PaymentDate *time.Time `json:"payment_date"`
}

type UpdateContractRequest struct {
	ClientName   string     `json:"client_name"`
	ClientPhone  string     `json:"client_phone"`
	ClientEmail  string     `json:"client_email"`
	TotalAmount  float64    `json:"total_amount"`
	PrepaidAmount float64   `json:"prepaid_amount"`
	PaidAmount   float64    `json:"paid_amount"`
	Status       string     `json:"status"`
	PaymentStatus string    `json:"payment_status"`
	PaymentDate  *time.Time `json:"payment_date"`
}

// DashboardMainResponse - ответ для главной вкладки дашборда менеджера салона
type DashboardMainResponse struct {
	// План и факт
	Plan           float64 `json:"plan"`           // План продаж на месяц
	Fact           float64 `json:"fact"`           // Факт продаж на текущий момент
	PlanPercent    int     `json:"plan_percent"`  // % выполнения плана

	// Динамика
	DynamicDay   float64 `json:"dynamic_day"`   // Динамика по сравнению со вчера (%)
	DynamicWeek  float64 `json:"dynamic_week"`  // Динамика по сравнению с прошлой неделей (%)

	// Прогноз
	Forecast int `json:"forecast"` // Прогноз выполнения до конца месяца (%)

	// Метрики
	AvgCheck       float64 `json:"avg_check"`        // Средний чек за месяц
	MarginPercent  float64 `json:"margin_percent"`   // Маржинальность (%)
	PrepaymentsSum float64 `json:"prepayments_sum"`  // Сумма предоплат за месяц

	// Светофор
	TrafficStatus           string `json:"traffic_status"`            // green/yellow/red
	ConversionMeasureStatus string `json:"conversion_measure_status"` // green/yellow/red
	ConversionContractStatus string `json:"conversion_contract_status"` // green/yellow/red

	// Оплаты
	PendingPayments []PendingPayment `json:"pending_payments"` // Список договоров на сегодня
}

// PendingPayment - запланированная оплата
type PendingPayment struct {
	ContractID   uuid.UUID `json:"contract_id"`
	ClientName   string    `json:"client_name"`
	Amount       float64   `json:"amount"`
	Status       string    `json:"status"`
	PaymentDate  string    `json:"payment_date"`
}

// === Воронка продаж ===

// FunnelStage - этап воронки
type FunnelStage struct {
	Stage       string `json:"stage"`       // traffic, consultation, measurement, kp, contract, payment
	Label       string `json:"label"`       // Отображаемое название
	Count       int    `json:"count"`       // Количество сделок
	Conversion  int    `json:"conversion"`  // Конверсия в % от предыдущего этапа
	Sum         float64 `json:"sum"`       // Сумма сделок
}

// HotDeal - горячая сделка (КП)
type HotDeal struct {
	ID          uuid.UUID `json:"id"`
	ClientName  string    `json:"client_name"`
	Phone       string    `json:"phone"`
	Amount      float64   `json:"amount"`
	CreatedAt   string    `json:"created_at"`   // Дата создания КП
	DaysStalled int       `json:"days_stalled"` // Дней без движения
	ManagerID   uuid.UUID `json:"manager_id"`
	ManagerName string    `json:"manager_name"`
}

// FreshLead - свежий лид
type FreshLead struct {
	ID          uuid.UUID `json:"id"`
	Source      string    `json:"source"`       // Источник
	ClientName  string    `json:"client_name"`
	Phone       string    `json:"phone"`
	CreatedAt   string    `json:"created_at"`   // Время поступления
	Status      string    `json:"status"`       // unassigned, in_progress, processed
	AssignedTo  *uuid.UUID `json:"assigned_to"`  // Ответственный
	ManagerName string    `json:"manager_name"`
}

// DashboardFunnelResponse - ответ для воронки
type DashboardFunnelResponse struct {
	Stages      []FunnelStage `json:"stages"`       // Этапы воронки
	HotDeals    []HotDeal     `json:"hot_deals"`    // Горячие сделки (КП)
	FreshLeads  []FreshLead   `json:"fresh_leads"`  // Свежие лиды
}

// AssignLeadRequest - назначение ответственного
type AssignLeadRequest struct {
	ManagerID string `json:"manager_id" binding:"required"`
}

// === Команда ===

// SalesRepMetrics - метрики продавца
type SalesRepMetrics struct {
	UserID        uuid.UUID `json:"user_id"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Role          string    `json:"role"`

	// Основные метрики
	Revenue       float64   `json:"revenue"`        // Выручка
	DealsCount    int       `json:"deals_count"`    // Количество чеков
	Conversion    float64   `json:"conversion"`     // Конверсия в договор (%)
	AvgCheck      float64   `json:"avg_check"`      // Средний чек
	DiscountPercent float64 `json:"discount_percent"` // % скидки
	ExtrasSum     float64   `json:"extras_sum"`     // Сумма допов (услуги, аксессуары)

	// Отклонения от среднего по салону (%)
	RevenueDeviation    float64 `json:"revenue_deviation"`
	DealsDeviation      float64 `json:"deals_deviation"`
	ConversionDeviation float64 `json:"conversion_deviation"`
	AvgCheckDeviation   float64 `json:"avg_check_deviation"`
}

// SalesRepHistory - история продавца для графиков
type SalesRepHistory struct {
	Month   string  `json:"month"`   // YYYY-MM
	Revenue float64 `json:"revenue"`
	Deals   int     `json:"deals"`
	AvgCheck float64 `json:"avg_check"`
}

// === Товары ===

// TopProduct - топ товар
type TopProduct struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Collection   string  `json:"collection"`
	Category     string  `json:"category"`
	Revenue      float64 `json:"revenue"`
	Quantity     int     `json:"quantity"`
	SharePercent float64 `json:"share_percent"` // Доля в общей выручке
	Margin       float64 `json:"margin"`       // Маржинальность %
}

// StockItem - остаток товара
type StockItem struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	ShowroomQty  int     `json:"showroom_qty"`  // На витрине
	WarehouseQty int     `json:"warehouse_qty"` // На складе
	TotalCost    float64 `json:"total_cost"`    // Общая стоимость запаса
	TurnoverDays int     `json:"turnover_days"` // Оборачиваемость в днях
}

// LostSale - упущенная продажа
type LostSale struct {
	Reason        string  `json:"reason"`         // Причина отказа
	RequestsCount int     `json:"requests_count"`  // Количество запросов
	LostRevenue   float64 `json:"lost_revenue"`    // Потерянная выручка
}

// CategoryTurnover - оборачиваемость по категориям
type CategoryTurnover struct {
	Category    string `json:"category"`
	AvgDays     int    `json:"avg_days"`     // Средняя оборачиваемость в днях
	IsSlowMoving bool  `json:"is_slow_moving"` // > 90 дней
}

// DashboardProductsResponse - ответ для товаров
type DashboardProductsResponse struct {
	TopProducts      []TopProduct      `json:"top_products"`      // Топ-10
	StockItems       []StockItem       `json:"stock_items"`       // Остатки
	LostSales        []LostSale         `json:"lost_sales"`        // Упущенные продажи
	CategoryTurnover []CategoryTurnover `json:"category_turnover"` // Оборачиваемость
	TotalRevenue     float64            `json:"total_revenue"`     // Общая выручка
}

// === Алерты ===

type AlertType string

const (
	AlertTypeOverdueMeasurement  AlertType = "overdue_measurement"  // Просроченный замер > 3 дней
	AlertTypeAbandonedKP         AlertType = "abandoned_kp"         // Брошенное КП > 5 дней
	AlertTypeConversionDrop      AlertType = "conversion_drop"     // Падение конверсии > 20%
	AlertTypeTrafficDrop         AlertType = "traffic_drop"         // Падение трафика > 30%
)

type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "critical" // Красный
	AlertSeverityWarning   AlertSeverity = "warning"   // Желтый
	AlertSeverityInfo      AlertSeverity = "info"      // Синий
)

// Alert - алерт
type Alert struct {
	ID          uuid.UUID     `json:"id"`
	UserID      uuid.UUID     `json:"user_id"`
	Type        AlertType     `json:"type"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Severity    AlertSeverity `json:"severity"`
	Link        string        `json:"link"` // Deep link
	ReadAt      *time.Time    `json:"read_at"`
	CreatedAt   time.Time     `json:"created_at"`
}

// AlertsResponse - ответ для списка алертов
type AlertsResponse struct {
	Alerts     []Alert `json:"alerts"`
	UnreadCount int    `json:"unread_count"`
}

// DashboardTeamResponse - ответ для команды
type DashboardTeamResponse struct {
	Period     string             `json:"period"`       // week/month/quarter
	TotalRevenue float64          `json:"total_revenue"`
	AvgRevenue   float64          `json:"avg_revenue"`  // Среднее по салону
	AvgConversion float64         `json:"avg_conversion"`
	AvgCheck     float64          `json:"avg_check"`
	SalesReps    []SalesRepMetrics `json:"sales_reps"`
}

// === Директивы от дилера ===

// TargetPlan - план продаж
type TargetPlan struct {
	TotalAmount    float64            `json:"total_amount"`    // Общая сумма плана
	ByCategory    map[string]float64  `json:"by_category"`     // По товарным группам
	CurrentAmount float64             `json:"current_amount"`  // Текущее выполнение
	Percent       int                 `json:"percent"`         // % выполнения
}

// Promotion - акция
type Promotion struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Condition     string    `json:"condition"`     // Условие
	DiscountMin   int       `json:"discount_min"`  // Мин. скидка
	DiscountMax   int       `json:"discount_max"`  // Макс. скидка
	EndDate       string    `json:"end_date"`      // Срок действия
	IsExpiring    bool      `json:"is_expiring"`   // Истекает < 7 дней
}

// ManagerTargetsResponse - директивы от дилера
type ManagerTargetsResponse struct {
	HasTargets bool            `json:"has_targets"` // План утверждён
	Plan       *TargetPlan     `json:"plan"`       // План продаж (nil если не утверждён)
	TargetConversion float64   `json:"target_conversion"` // Целевая конверсия
	CurrentConversion float64  `json:"current_conversion"` // Текущая конверсия
	TargetExtrasPercent float64 `json:"target_extras_percent"` // Целевая доля допов
	CurrentExtrasPercent float64 `json:"current_extras_percent"` // Текущая доля допов
	Promotions []Promotion     `json:"promotions"`  // Акции
	BonusForecast float64      `json:"bonus_forecast"` // Прогноз премии
	MaxBonus float64           `json:"max_bonus"`    // Максимальная премия
	WarningLevel string        `json:"warning_level"` // none/yellow/red
}

// === DASHBOARD DEALER ===

// DealerSummaryResponse - сводка для дилера
type DealerSummaryResponse struct {
	NetProfit            float64 `json:"netProfit"`
	GrossRevenue        float64 `json:"grossRevenue"`
	PlanCompletionPercent int   `json:"planCompletionPercent"`
	MarginProfit        float64 `json:"marginProfit"`
	ActiveAlerts        int     `json:"activeAlerts"`
}

// DealerFinanceResponse - финансы для дилера
type DealerFinanceResponse struct {
	Revenue           float64          `json:"revenue"`
	COGS              float64          `json:"cogs"`
	Rent              float64          `json:"rent"`
	Utilities         float64          `json:"utilities"`
	Payroll           float64          `json:"payroll"`
	Taxes              float64          `json:"taxes"`
	Logistics         float64          `json:"logistics"`
	Marketing         float64          `json:"marketing"`
	Defects           float64          `json:"defects"`
	OtherExpenses     float64          `json:"other_expenses"`
	Bonus             float64          `json:"bonus"`
	NetProfit        float64          `json:"net_profit"`
	NetProfitForecast float64          `json:"net_profit_forecast"`
	PrevMonthNetProfit float64        `json:"prev_month_net_profit"`
	ExpenseBreakdown  []ExpenseBreakdown `json:"expense_breakdown"`
}

// ExpenseBreakdown - статья расходов
type ExpenseBreakdown struct {
	Category         string  `json:"category"`
	Amount          float64  `json:"amount"`
	PercentOfRevenue float64 `json:"percent_of_revenue"`
	PrevMonthAmount float64 `json:"prev_month_amount"`
}

// DealerFunnelResponse - воронка для дилера
type DealerFunnelResponse struct {
	Stages        []FunnelStage    `json:"stages"`
	SalonPlanData []SalonPlanData `json:"salon_plan_data"`
}

// SalonPlanData - план салона
type SalonPlanData struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Plan           float64   `json:"plan"`
	Fact           float64   `json:"fact"`
	Percent        int       `json:"percent"`
	Forecast       string    `json:"forecast"`
	ManagersCount  int       `json:"managers_count"`
	AvgCheck        float64   `json:"avg_check"`
}

// DealerProductsResponse - товары для дилера
type DealerProductsResponse struct {
	TotalRevenue   float64             `json:"total_revenue"`
	TopProducts    []DealerTopProduct   `json:"top_products"`
	Inventory     []DealerInventoryItem `json:"inventory"`
}

// DealerTopProduct - топ товар
type DealerTopProduct struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Revenue      float64 `json:"revenue"`
	Quantity     int     `json:"quantity"`
	SharePercent float64 `json:"share_percent"`
}

// DealerInventoryItem - остаток товара
type DealerInventoryItem struct {
	ID             string  `json:"id"`
	Collection    string  `json:"collection"`
	StockWarehouse int    `json:"stock_warehouse"`
	OnDisplay     int     `json:"on_display"`
	SoldPeriod    int     `json:"sold_period"`
	TurnoverDays  int     `json:"turnover_days"`
}

// === DASHBOARD FRANCHISER ===

// FranchiserSummaryResponse - сводка для франчайзера
type FranchiserSummaryResponse struct {
	PlanPercent       int     `json:"planPercent"`
	ForecastPercent  int     `json:"forecastPercent"`
	ActiveDealers    int     `json:"activeDealers"`
	AvgConversion   float64 `json:"avgConversion"`
	AvgMargin      float64 `json:"avgMargin"`
}

// FranchiserNetworkResponse - данные сети
type FranchiserNetworkResponse struct {
	NetworkData []FranchiserDealerData `json:"network_data"`
	Overview   FranchiserOverview     `json:"overview"`
}

// FranchiserDealerData - данные дилера
type FranchiserDealerData struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	SalonCount    int       `json:"salon_count"`
	Plan          float64   `json:"plan"`
	Fact          float64   `json:"fact"`
	PlanPercent   int       `json:"plan_percent"`
	Forecast     string    `json:"forecast"`
}

// FranchiserOverview - обзор сети
type FranchiserOverview struct {
	PlanAmount        float64 `json:"plan_amount"`
	PlanPercent      int     `json:"plan_percent"`
	ForecastAmount   float64 `json:"forecast_amount"`
	ForecastPercent  int     `json:"forecast_percent"`
	ActiveDealers    int     `json:"active_dealers"`
	AvgConversion   float64 `json:"avg_conversion"`
	AvgMargin       float64 `json:"avg_margin"`
	RedZoneDealers   int     `json:"red_zone_dealers"`
}

// FranchiserHealthResponse - здоровье сети
type FranchiserHealthResponse struct {
	SLA          float64             `json:"sla"`
	TotalDealers int                 `json:"total_dealers"`
	ActiveDealers int                `json:"active_dealers"`
	TotalSalons  int                 `json:"total_salons"`
	RedDealers   []FranchiserDealerHealth `json:"red_dealers"`
}

// FranchiserDealerHealth - проблемный дилер
type FranchiserDealerHealth struct {
	DealerID   uuid.UUID `json:"dealer_id"`
	DealerName string   `json:"dealer_name"`
	Issue     string   `json:"issue"`
	Severity   string   `json:"severity"` // critical, warning
}

// FranchiserTeamResponse - команда
type FranchiserTeamResponse struct {
	TeamMembers []FranchiserTeamMember `json:"team_members"`
}

// FranchiserTeamMember - член команды
type FranchiserTeamMember struct {
	ID          uuid.UUID `json:"id"`
	Name        string   `json:"name"`
	DealersCount int    `json:"dealers_count"`
	Role        string   `json:"role"`
}

// === DASHBOARD TERRITORY MANAGER ===

// TerritorySummaryResponse - сводка для территориального менеджера
type TerritorySummaryResponse struct {
	PlanCompletionPercent      int `json:"planCompletionPercent"`
	QuarterForecastPercent    int `json:"quarterForecastPercent"`
	RedZoneDealersCount       int `json:"redZoneDealersCount"`
	AvgConversion          float64 `json:"avgConversion"`
	ActiveAlerts          int     `json:"activeAlerts"`
}

// TerritoryFunnelResponse - воронка
type TerritoryFunnelResponse struct {
	Stages []FunnelStage `json:"stages"`
}

// TerritoryPlanFactResponse - план-факт
type TerritoryPlanFactResponse struct {
	Dealers   []TerritoryDealerPlanFact `json:"dealers"`
	TotalPlan float64                  `json:"total_plan"`
	TotalFact float64                  `json:"total_fact"`
}

// TerritoryDealerPlanFact - план-факт дилера
type TerritoryDealerPlanFact struct {
	ID          uuid.UUID `json:"id"`
	DealerName  string   `json:"dealer_name"`
	Plan        float64  `json:"plan"`
	Fact        float64  `json:"fact"`
	PlanPercent int     `json:"plan_percent"`
}

// TerritoryCommunicationsResponse - коммуникации
type TerritoryCommunicationsResponse struct {
	Tasks          []TerritoryTask `json:"tasks"`
	UnreadMessages int             `json:"unread_messages"`
}

// TerritoryTask - задача
type TerritoryTask struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Status      string    `json:"status"`
	AssignedTo  uuid.UUID `json:"assigned_to"`
	DueDate     string    `json:"due_date"`
}

// TerritoryBenchmarksResponse - бенчмарки
type TerritoryBenchmarksResponse struct {
	TerritoryConversion float64 `json:"territory_conversion"`
	TerritoryAvgCheck    float64 `json:"territory_avg_check"`
	NetworkAvgConversion float64 `json:"network_avg_conversion"`
	NetworkAvgCheck      float64 `json:"network_avg_check"`
}

// === Dealer Additional DTOs ===

// DealerTasksResponse - задачи для дилера
type DealerTasksResponse struct {
	Tasks     []DealerTaskItem `json:"tasks"`
	Total     int             `json:"total"`
	Overdue   int             `json:"overdue"`
}

type DealerTaskItem struct {
	ID          uuid.UUID `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Status     string   `json:"status"`
	Priority   string   `json:"priority"`
	DueDate    string   `json:"due_date"`
	IsOverdue   bool     `json:"is_overdue"`
}

// DealerRequest - запрос от дилера (db model)
type DealerRequest struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	DealerID    uuid.UUID `json:"dealer_id" gorm:"type:uuid"`
	Type        string    `json:"type" gorm:"type:varchar(50)"`
	Description string   `json:"description" gorm:"type:text"`
	Amount      float64   `json:"amount" gorm:"type:decimal(12,2)"`
	Status     string    `json:"status" gorm:"type:varchar(50);default:'pending'"`
	CreatedAt   time.Time `json:"created_at"`
}

func (DealerRequest) TableName() string {
	return "dealer_requests"
}

// DealerRequestsResponse - запросы к бренду
type DealerRequestsResponse struct {
	Requests []DealerRequestItem `json:"requests"`
	Total    int                `json:"total"`
	Pending int                `json:"pending"`
}

type DealerRequestItem struct {
	ID          uuid.UUID `json:"id"`
	DealerID    uuid.UUID `json:"dealer_id"`
	Type        string    `json:"type"`
	Description string   `json:"description"`
	Amount      float64   `json:"amount"`
	Status     string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// DealerMarketingBudgetResponse - маркетинговый бюджет
type DealerMarketingBudgetResponse struct {
	Quarter       string `json:"quarter"`
	TotalAmount   float64 `json:"total_amount"`
	UsedAmount    float64 `json:"used_amount"`
	Remaining     float64 `json:"remaining"`
	UsagePercent  int     `json:"usage_percent"`
}

// DealerAlertsResponse - алерты дилера
type DealerAlertsResponse struct {
	Alerts      []DealerAlertItem `json:"alerts"`
	UnreadCount int              `json:"unread_count"`
}

type DealerAlertItem struct {
	ID       uuid.UUID `json:"id"`
	Type     string    `json:"type"`
	Title    string    `json:"title"`
	Message  string    `json:"message"`
	Priority string    `json:"priority"`
}

// === FRANCHISER DASHBOARD ===

// FranchiserDealersResponse - ответ с списком дилеров
type FranchiserDealersResponse struct {
	Dealers []FranchiserDealerItem `json:"dealers"`
	Total   int                    `json:"total"`
}

// FranchiserDealerItem - элемент списка дилеров
type FranchiserDealerItem struct {
	ID           uuid.UUID `json:"id"`
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	Plan         float64  `json:"plan"`
	Fact         float64  `json:"fact"`
	PlanPercent  int      `json:"plan_percent"`
	Status      string   `json:"status"` // green, yellow, red
}

// FranchiserRequestsResponse - ответ с запросами
type FranchiserRequestsResponse struct {
	Requests []FranchiserRequestItem `json:"requests"`
	Total    int                      `json:"total"`
	Pending  int                      `json:"pending"`
}

// FranchiserRequestItem - элемент запроса
type FranchiserRequestItem struct {
	ID          uuid.UUID `json:"id"`
	DealerID    uuid.UUID `json:"dealer_id"`
	DealerName  string    `json:"dealer_name"`
	Type        string    `json:"type"` // discount, marketing, return, etc
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}