package services

import (
	"context"
	"errors"
	"log"
	"time"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo repository.UserRepositoryInterface
	db       *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(db),
		db:       db,
	}
}

func NewUserServiceWithInterface(userRepo repository.UserRepositoryInterface, db *gorm.DB) *UserService {
	return &UserService{
		userRepo: userRepo,
		db:       db,
	}
}

// === МЕТОДЫ ПРОФИЛЯ ===

func (s *UserService) GetProfile(userID uuid.UUID) (*models.User, error) {
	return s.userRepo.GetUserByID(context.Background(), userID)
}

func (s *UserService) UpdateProfile(userID uuid.UUID, req models.UserUpdateRequest) (*models.User, error) {
	updateData := map[string]interface{}{
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"phone":      req.Phone,
		"updated_at": time.Now(),
	}
	err := s.userRepo.UpdateUserFields(context.Background(), userID, updateData)
	if err != nil {
		return nil, err
	}
	return s.userRepo.GetUserByID(context.Background(), userID)
}

func (s *UserService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetUserByID(context.Background(), userID)
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
	if err != nil {
		return errors.New("invalid old password")
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	updateData := map[string]interface{}{
		"password_hash": string(newHash),
		"updated_at":    time.Now(),
	}
	return s.userRepo.UpdateUserFields(context.Background(), userID, updateData)
}

// === МЕТОДЫ HR (СОТРУДНИКИ) ===

func (s *UserService) GetEmployees(tenantID uuid.UUID) ([]models.User, error) {
	return s.userRepo.FindUsersByTenantID(context.Background(), tenantID)
}

// CreateEmployee
func (s *UserService) CreateEmployee(req models.CreateEmployeeRequest, tenantID uuid.UUID, creatorID uuid.UUID, creatorRole string) (*models.User, error) {
	// 1. Проверка прав
	if creatorRole != string(models.RoleSuperAdmin) {
		if req.Role == models.RoleSuperAdmin || req.Role == models.RoleFranchisor {
			return nil, errors.New("permission denied: cannot create user with this role")
		}
	}

	// 2. Валидация TenantID для Франчайзера
	// Франчайзер ОБЯЗАН быть привязан к сети.
	if req.Role == models.RoleFranchisor && tenantID == uuid.Nil {
		return nil, errors.New("tenant_id is required for franchiser role")
	}

	// 3. Подготовка данных
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	managerID := req.ManagedBy
	if managerID == nil {
		managerID = &creatorID
	}

	// 4. Создание пользователя
	user := models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
		TenantID:     &tenantID, // Если uuid.Nil, GORM запишет NULL (для SuperAdmin)
		ManagedBy:    managerID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
	}

	if err := s.userRepo.CreateUser(context.Background(), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) UpdateEmployee(userID, tenantID uuid.UUID, req models.UpdateEmployeeRequest) (*models.User, error) {
	_, err := s.userRepo.FindUserByIDAndTenant(context.Background(), userID, tenantID)
	if err != nil {
		return nil, errors.New("employee not found in your network")
	}
	updateData := map[string]interface{}{"updated_at": time.Now()}
	if req.FirstName != "" {
		updateData["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updateData["last_name"] = req.LastName
	}
	if req.Phone != "" {
		updateData["phone"] = req.Phone
	}
	if req.Role != "" {
		if req.Role == models.RoleSuperAdmin || req.Role == models.RoleFranchisor {
			return nil, errors.New("invalid role assignment")
		}
		updateData["role"] = req.Role
	}
	if req.ManagedBy != nil {
		updateData["managed_by"] = req.ManagedBy
	}

	if err := s.userRepo.UpdateUserFields(context.Background(), userID, updateData); err != nil {
		return nil, err
	}
	return s.userRepo.GetUserByID(context.Background(), userID)
}

func (s *UserService) DeleteEmployee(userID, tenantID uuid.UUID) error {
	_, err := s.userRepo.FindUserByIDAndTenant(context.Background(), userID, tenantID)
	if err != nil {
		return errors.New("employee not found in your network")
	}
	return s.userRepo.DeleteUser(context.Background(), userID)
}

// === МЕТОДЫ ДЛЯ СУПЕР-АДМИНА ===

func (s *UserService) GetAllUsersGlobal() ([]models.User, error) {
	return s.userRepo.FindAllGlobal(context.Background())
}

// === УПРАВЛЕНИЕ САЛОНАМИ ===

// CreateSalon создает салон внутри сети
func (s *UserService) CreateSalon(ctx context.Context, tenantID uuid.UUID, name, address string) (*models.Salon, error) {
	salon := &models.Salon{
		TenantID: tenantID,
		Name:     name,
		Address:  address,
	}
	salonRepo := repository.NewSalonRepository(s.db)
	if err := salonRepo.CreateSalon(ctx, salon); err != nil {
		return nil, err
	}
	return salon, nil
}

// GetMySalons возвращает список салонов сети с менеджерами
func (s *UserService) GetMySalons(ctx context.Context, tenantID uuid.UUID) ([]models.Salon, error) {
	salonRepo := repository.NewSalonRepository(s.db)

	salons, err := salonRepo.GetSalonsByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	if len(salons) == 0 {
		return salons, nil
	}

	salonIDs := make([]uuid.UUID, len(salons))
	for i, s := range salons {
		salonIDs[i] = s.ID
	}

	var managers []models.User
	if err := s.db.WithContext(ctx).Where("salon_id IN ?", salonIDs).Find(&managers).Error; err != nil {
		log.Printf("Error fetching managers for salons: %v", err)
		return salons, nil
	}

	managerMap := make(map[uuid.UUID]models.User)
	for _, m := range managers {
		if m.SalonID != nil {
			managerMap[*m.SalonID] = m
		}
	}

	for i := range salons {
		if mgr, ok := managerMap[salons[i].ID]; ok {
			salons[i].Manager = &mgr
		}
	}

	return salons, nil
}

// AssignManagerToSalon привязывает менеджера к конкретному салону
func (s *UserService) AssignManagerToSalon(ctx context.Context, managerID, salonID uuid.UUID) error {
	salonRepo := repository.NewSalonRepository(s.db)
	return salonRepo.UpdateUserSalon(ctx, managerID, salonID)
}

// GetSalonByID вспомогательный метод для получения салона
func (s *UserService) GetSalonByID(ctx context.Context, salonID uuid.UUID) (*models.Salon, error) {
	var salon models.Salon
	if err := s.db.WithContext(ctx).First(&salon, "id = ?", salonID).Error; err != nil {
		return nil, err
	}
	return &salon, nil
}

// UpdateSalon обновляет данные салона
func (s *UserService) UpdateSalon(ctx context.Context, salonID uuid.UUID, name, address string) (*models.Salon, error) {
	salon, err := s.GetSalonByID(ctx, salonID)
	if err != nil {
		return nil, err
	}

	salon.Name = name
	salon.Address = address

	if err := s.db.WithContext(ctx).Save(salon).Error; err != nil {
		return nil, err
	}
	return salon, nil
}

// DeleteSalon удаляет салон
func (s *UserService) DeleteSalon(ctx context.Context, salonID uuid.UUID) error {
	salonRepo := repository.NewSalonRepository(s.db)
	return salonRepo.DeleteSalon(ctx, salonID)
}
