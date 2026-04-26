package repository

import (
	"context"

	"franchise-saas-backend/internal/models"

	"gorm.io/gorm"
)

// ---------- Интерфейс ----------
type PlanRepository interface {
	Create(ctx context.Context, p *models.Plan) error
	GetByID(ctx context.Context, id string) (*models.Plan, error)
	List(ctx context.Context, opts ListOptions) ([]*models.Plan, int64, error)
	Update(ctx context.Context, p *models.Plan) error
	Delete(ctx context.Context, id string) error
}

// ---------- Реализация ----------
type planRepo struct{ db *gorm.DB }

func NewPlanRepository(db *gorm.DB) PlanRepository { return &planRepo{db: db} }

func (r *planRepo) Create(ctx context.Context, p *models.Plan) error {
	return r.db.WithContext(ctx).Create(p).Error
}
func (r *planRepo) GetByID(ctx context.Context, id string) (*models.Plan, error) {
	var p models.Plan
	if err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}
func (r *planRepo) List(ctx context.Context, opts ListOptions) ([]*models.Plan, int64, error) {
	var (
		plans []*models.Plan
		cnt   int64
	)
	q := r.db.WithContext(ctx).Model(&models.Plan{})

	if opts.Search != "" {
		q = q.Where("LOWER(name) LIKE LOWER(?)", "%"+opts.Search+"%")
	}
	if opts.OrderBy != "" {
		dir := "ASC"
		if opts.Desc {
			dir = "DESC"
		}
		q = q.Order(opts.OrderBy + " " + dir)
	}
	if err := q.Count(&cnt).Error; err != nil {
		return nil, 0, err
	}
	if opts.Limit > 0 {
		q = q.Offset(opts.Offset).Limit(opts.Limit)
	}
	if err := q.Find(&plans).Error; err != nil {
		return nil, 0, err
	}
	return plans, cnt, nil
}
func (r *planRepo) Update(ctx context.Context, p *models.Plan) error {
	return r.db.WithContext(ctx).
		Model(&models.Plan{}).
		Where("id = ?", p.ID).
		Updates(map[string]interface{}{
			"name":       p.Name,
			"price":      p.Price,
			"max_salons": p.MaxSalons,
			"max_users":  p.MaxUsers,
		}).Error
}
func (r *planRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Plan{}).Error
}
