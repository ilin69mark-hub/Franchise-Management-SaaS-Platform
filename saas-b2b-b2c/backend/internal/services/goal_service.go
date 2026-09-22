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
	UpdateGoal(ctx context.Context, id string, dto UpdateGoalDTO, assignerID, tenantID string) (*models.Goal, error)
	GetMyGoal(ctx context.Context, assigneeID string, date time.Time) (*models.Goal, error)
	GetVisibleGoals(ctx context.Context, userID, role, tenantID string) ([]models.Goal, error)
	DeleteGoal(ctx context.Context, id string) error
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

	assigneeUUID, err := uuid.Parse(dto.AssigneeID)
	if err != nil {
		return nil, err
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
func (s *goalService) UpdateGoal(ctx context.Context, id string, dto UpdateGoalDTO, assignerID, tenantID string) (*models.Goal, error) {
	goal, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("goal not found")
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
func (s *goalService) DeleteGoal(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
