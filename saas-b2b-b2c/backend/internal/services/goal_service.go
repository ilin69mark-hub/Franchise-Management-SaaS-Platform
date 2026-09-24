package services

import (
	"context"
	"errors"
	"time"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/google/uuid"
)

type GoalService interface {
	CreateGoal(ctx context.Context, dto CreateGoalDTO, assignerID, tenantID string) (*models.Goal, error)
	UpdateGoal(ctx context.Context, id string, dto UpdateGoalDTO, requesterID, tenantID, requesterRole string) (*models.Goal, error)
	GetMyGoal(ctx context.Context, assigneeID string, date time.Time) (*models.Goal, error)
	GetVisibleGoals(ctx context.Context, userID, role, tenantID string) ([]models.Goal, error)
	DeleteGoal(ctx context.Context, id, requesterID, tenantID, requesterRole string) error
}

// errForbidden — владелец/tenant не совпал (маппится в 403, а не 500).
var errGoalForbidden = errors.New("forbidden: goal not in your scope")

// sameTenant — оба tenant заданы и равны; если у цели tenant нет — требуем assigner.
func sameGoalTenant(goalTenant *uuid.UUID, callerTenant string) bool {
	if goalTenant == nil {
		return false
	}
	if callerTenant == "" {
		return false
	}
	tid, err := uuid.Parse(callerTenant)
	if err != nil {
		return false
	}
	return *goalTenant == tid
}

/* DTO – данные, получаемые от фронтенда */
type CreateGoalDTO struct {
	AssigneeID   string  `json:"assignee_id"`
	Role         string  `json:"role"`
	SalesPlan    float64 `json:"sales_plan"`
	LeadsPlan    int     `json:"leads_plan"`
	CallsPlan    int     `json:"calls_plan"`
	MeetingsPlan int     `json:"meetings_plan"`
	Period       string  `json:"period"`      // "day", "week", "month"
	StartDate    string  `json:"start_date"`  // YYYY-MM-DD
	EndDate      string  `json:"end_date"`    // YYYY-MM-DD
	TargetDate   string  `json:"target_date"` // deprecated
}

type UpdateGoalDTO struct {
	SalesPlan    float64 `json:"sales_plan"`
	LeadsPlan    int     `json:"leads_plan"`
	CallsPlan    int     `json:"calls_plan"`
	MeetingsPlan int     `json:"meetings_plan"`
	Period       string  `json:"period"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
}

/* Реализация */
type goalService struct{ repo repository.GoalRepository }

func NewGoalService(r repository.GoalRepository) GoalService { return &goalService{repo: r} }

/* ---------- Проверка прав: кто может назначать план кому ---------- */
func canAssign(assignerRole, assigneeRole string) bool {
	allowed := map[string][]string{
		string(models.RoleSuperAdmin):        {string(models.RoleFranchisor), string(models.RoleFranchisorManager), string(models.RoleDealer), string(models.RoleDealerManager)},
		string(models.RoleFranchisor):        {string(models.RoleFranchisorManager), string(models.RoleDealer), string(models.RoleDealerManager)},
		string(models.RoleFranchisorManager): {string(models.RoleDealer), string(models.RoleDealerManager)},
		string(models.RoleDealer):            {string(models.RoleDealerManager)},
		string(models.RoleDealerManager):     {},
	}
	for _, r := range allowed[assignerRole] {
		if r == assigneeRole {
			return true
		}
	}
	return false
}

/* ---------- CreateGoal ---------- */
func (s *goalService) CreateGoal(ctx context.Context, dto CreateGoalDTO, assignerID, tenantID string) (*models.Goal, error) {
	assignerRole := ""
	switch v := ctx.Value("role").(type) {
	case string:
		assignerRole = v
	case models.Role:
		assignerRole = string(v)
	}
	if !canAssign(assignerRole, dto.Role) {
		return nil, errors.New("you are not allowed to assign a goal to this role")
	}
	// Валидация сумм — защита от накрутки KPI отрицательными значениями
	if dto.SalesPlan < 0 || dto.SalesPlan > 1e12 {
		return nil, errors.New("sales_plan out of range (0..1e12)")
	}
	if dto.LeadsPlan < 0 || dto.LeadsPlan > 100000 {
		return nil, errors.New("leads_plan out of range (0..100000)")
	}
	if dto.CallsPlan < 0 || dto.CallsPlan > 100000 {
		return nil, errors.New("calls_plan out of range")
	}
	if dto.MeetingsPlan < 0 || dto.MeetingsPlan > 100000 {
		return nil, errors.New("meetings_plan out of range")
	}
	if dto.SalesPlan == 0 && dto.LeadsPlan == 0 && dto.CallsPlan == 0 && dto.MeetingsPlan == 0 {
		return nil, errors.New("at least one plan must be >0")
	}

	assigneeUUID, err := uuid.Parse(dto.AssigneeID)
	if err != nil {
		return nil, err
	}

	// RE-AUDIT: assignee обязан существовать и быть в сети назначающего
	// (раньше цели писались на произвольный UUID чужой сети — фантомные планы).
	if assignerRole != string(models.RoleSuperAdmin) {
		assigneeTenant, terr := s.repo.GetUserTenant(ctx, dto.AssigneeID)
		if terr != nil || assigneeTenant == nil {
			return nil, errors.New("assignee not found")
		}
		if tenantID == "" || assigneeTenant.String() != tenantID {
			return nil, errors.New("forbidden: assignee not in your network")
		}
	}

	period := models.PeriodDay
	if dto.Period == "week" || dto.Period == "month" || dto.Period == "year" || dto.Period == "custom" {
		period = models.GoalPeriod(dto.Period)
	} else if dto.Period != "" {
		return nil, errors.New("invalid period")
	}

	var startDate, endDate, targetDate time.Time

	if dto.StartDate != "" {
		var err error
		startDate, err = time.Parse("2006-01-02", dto.StartDate)
		if err != nil {
			return nil, errors.New("invalid start_date")
		}
	}
	if dto.EndDate != "" {
		var err error
		endDate, err = time.Parse("2006-01-02", dto.EndDate)
		if err != nil {
			return nil, errors.New("invalid end_date")
		}
	}
	if startDate.IsZero() && !endDate.IsZero() {
		startDate = endDate
	}
	if !startDate.IsZero() && endDate.IsZero() {
		endDate = startDate
	}
	targetDate = endDate

	goal := &models.Goal{
		AssignerID:   uuid.MustParse(assignerID),
		AssigneeID:   assigneeUUID,
		Role:         dto.Role,
		SalesPlan:    dto.SalesPlan,
		LeadsPlan:    dto.LeadsPlan,
		CallsPlan:    dto.CallsPlan,
		MeetingsPlan: dto.MeetingsPlan,
		Period:       period,
		StartDate:    startDate,
		EndDate:      endDate,
		TargetDate:   targetDate,
	}
	if tenantID != "" {
		tid, _ := uuid.Parse(tenantID)
		goal.TenantID = &tid
	}
	if err := s.repo.Create(ctx, goal); err != nil {
		return nil, err
	}
	return goal, nil
}

/* ---------- UpdateGoal ---------- */
// Правило: super_admin — всё; остальные — только своя цель (assignee/assigner)
// или тот же tenant + canAssign(requesterRole → goal.Role). Иначе 403.
func (s *goalService) UpdateGoal(ctx context.Context, id string, dto UpdateGoalDTO, requesterID, tenantID, requesterRole string) (*models.Goal, error) {
	goal, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("goal not found")
	}
	if requesterRole != string(models.RoleSuperAdmin) {
		isOwner := goal.AssigneeID.String() == requesterID || goal.AssignerID.String() == requesterID
		inTenant := sameGoalTenant(goal.TenantID, tenantID)
		if !isOwner && !(inTenant && canAssign(requesterRole, goal.Role)) {
			return nil, errGoalForbidden
		}
		if goal.TenantID != nil && tenantID != "" && !inTenant && !isOwner {
			return nil, errGoalForbidden
		}
	}
	// Валидация — отрицательные планы запрещены
	if dto.SalesPlan < 0 || dto.SalesPlan > 1e12 {
		return nil, errors.New("sales_plan out of range")
	}
	if dto.LeadsPlan < 0 || dto.LeadsPlan > 100000 {
		return nil, errors.New("leads_plan out of range")
	}
	if dto.CallsPlan < 0 || dto.CallsPlan > 100000 {
		return nil, errors.New("calls_plan out of range")
	}
	if dto.MeetingsPlan < 0 || dto.MeetingsPlan > 100000 {
		return nil, errors.New("meetings_plan out of range")
	}

	if dto.SalesPlan > 0 {
		goal.SalesPlan = dto.SalesPlan
	}
	if dto.LeadsPlan > 0 {
		goal.LeadsPlan = dto.LeadsPlan
	}
	if dto.CallsPlan > 0 {
		goal.CallsPlan = dto.CallsPlan
	}
	if dto.MeetingsPlan > 0 {
		goal.MeetingsPlan = dto.MeetingsPlan
	}
	if dto.Period != "" {
		if dto.Period != "day" && dto.Period != "week" && dto.Period != "month" && dto.Period != "year" && dto.Period != "custom" {
			return nil, errors.New("invalid period")
		}
		goal.Period = models.GoalPeriod(dto.Period)
	}
	if dto.StartDate != "" {
		var err error
		goal.StartDate, err = time.Parse("2006-01-02", dto.StartDate)
		if err != nil {
			return nil, errors.New("invalid start_date")
		}
	}
	if dto.EndDate != "" {
		var err error
		goal.EndDate, err = time.Parse("2006-01-02", dto.EndDate)
		if err != nil {
			return nil, errors.New("invalid end_date")
		}
	}

	if err := s.repo.Update(ctx, goal); err != nil {
		return nil, err
	}
	return goal, nil
}

/* ---------- GetMyGoal – план текущего пользователя на конкретную дату ---------- */
func (s *goalService) GetMyGoal(ctx context.Context, assigneeID string, date time.Time) (*models.Goal, error) {
	return s.repo.GetByAssigneeAndDate(ctx, assigneeID, date)
}

/* ---------- GetVisibleGoals – список целей, которые видит пользователь ---------- */
func (s *goalService) GetVisibleGoals(ctx context.Context, userID, role, tenantID string) ([]models.Goal, error) {
	return s.repo.ListVisibleForUser(ctx, userID, role, tenantID)
}

/* ---------- DeleteGoal ---------- */
// Правило: super_admin — всё; остальные — только назначивший (assigner)
// или тот же tenant + canAssign(requesterRole → goal.Role). Assignee сам
// свою цель удалить не может (защита от скрытия недовыполнения).
func (s *goalService) DeleteGoal(ctx context.Context, id, requesterID, tenantID, requesterRole string) error {
	goal, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.New("goal not found")
	}
	if requesterRole != string(models.RoleSuperAdmin) {
		isAssigner := goal.AssignerID.String() == requesterID
		inTenant := sameGoalTenant(goal.TenantID, tenantID)
		if !isAssigner && !(inTenant && canAssign(requesterRole, goal.Role)) {
			return errGoalForbidden
		}
		if goal.TenantID != nil && tenantID != "" && !inTenant && !isAssigner {
			return errGoalForbidden
		}
	}
	return s.repo.Delete(ctx, id)
}
