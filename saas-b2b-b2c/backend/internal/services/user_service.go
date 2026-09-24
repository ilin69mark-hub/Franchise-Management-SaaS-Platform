package services

import (
	"context"
	"errors"
	"log"
	"strings"
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
	updateData := map[string]interface{}{}

	if req.FirstName != "" {
		updateData["first_name"] = req.FirstName
	}
	if req.LastName != "" {
		updateData["last_name"] = req.LastName
	}
	if req.Phone != "" {
		updateData["phone"] = req.Phone
	}
	if req.DisplayName != nil {
		updateData["display_name"] = *req.DisplayName
	}
	if req.Position != nil {
		updateData["position"] = *req.Position
	}
	if req.Bio != nil {
		updateData["bio"] = *req.Bio
	}
	if req.Quote != nil {
		updateData["quote"] = *req.Quote
	}
	if req.UserStatus != nil {
		updateData["user_status"] = *req.UserStatus
	}
	if req.AvailableForQuestions != nil {
		updateData["available_for_questions"] = *req.AvailableForQuestions
	}
	if req.ContactsEmailVisible != nil {
		updateData["contacts_email_visible"] = *req.ContactsEmailVisible
	}
	if req.ContactsPhoneVisible != nil {
		updateData["contacts_phone_visible"] = *req.ContactsPhoneVisible
	}
	if req.ContactsPhone != nil {
		updateData["contacts_phone"] = *req.ContactsPhone
	}
	if req.ContactsTelegram != nil {
		updateData["contacts_telegram"] = *req.ContactsTelegram
	}
	if req.ContactsWhatsApp != nil {
		updateData["contacts_whatsapp"] = *req.ContactsWhatsApp
	}
	if req.ContactsWorkingHours != nil {
		updateData["contacts_working_hours"] = *req.ContactsWorkingHours
	}

	updateData["updated_at"] = time.Now()
	err := s.userRepo.UpdateUserFields(context.Background(), userID, updateData)
	if err != nil {
		return nil, err
	}
	return s.userRepo.GetUserByID(context.Background(), userID)
}

func (s *UserService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	// F6: смена на "123" запрещена — та же политика, что при регистрации.
	if len(newPassword) < 12 || len(newPassword) > 128 {
		return errors.New("password must be 12..72 characters")
	}
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

// CreateEmployee — strict whitelist per creator role + tenant isolation
func (s *UserService) CreateEmployee(req models.CreateEmployeeRequest, tenantID uuid.UUID, creatorID uuid.UUID, creatorRole string) (*models.User, error) {
	// 1. Строгая проверка прав — whitelist
	allowedByCreator := map[string]map[models.Role]bool{
		string(models.RoleSuperAdmin): {
			models.RoleFranchisor:        true,
			models.RoleFranchisorManager: true,
			models.RoleDealer:            true,
			models.RoleDealerManager:     true,
		},
		string(models.RoleFranchisor): {
			models.RoleFranchisorManager: true,
			models.RoleDealer:            true,
		},
		string(models.RoleFranchisorManager): {
			models.RoleDealer: true,
		},
		string(models.RoleDealer): {
			models.RoleDealerManager: true,
		},
	}
	if allowed, ok := allowedByCreator[creatorRole]; !ok || !allowed[req.Role] {
		// super_admin уже покрыт map, остальные строго по whitelist
		if creatorRole != string(models.RoleSuperAdmin) || !allowed[req.Role] {
			return nil, errors.New("permission denied: cannot create user with this role")
		}
	}
	// 2. Валидация TenantID для Франчайзера
	if req.Role == models.RoleFranchisor && tenantID == uuid.Nil {
		return nil, errors.New("tenant_id is required for franchiser role")
	}
	if tenantID == uuid.Nil && req.Role != models.RoleSuperAdmin {
		// tenant_id must not be Nil for non-super_admin tenants — prevents uuid.Nil bypass
		return nil, errors.New("tenant_id is required")
	}

	// 2b. ManagedBy must be in same tenant if provided
	if req.ManagedBy != nil {
		mgr, err := s.userRepo.GetUserByID(context.Background(), *req.ManagedBy)
		if err != nil {
			return nil, errors.New("managed_by user not found")
		}
		if tenantID != uuid.Nil && mgr.TenantID != nil && *mgr.TenantID != tenantID {
			return nil, errors.New("managed_by must be in same tenant")
		}
		if tenantID != uuid.Nil && mgr.TenantID == nil {
			return nil, errors.New("managed_by must be in same tenant")
		}
	}

	// 3. Подготовка данных (F6: HR-пароли — та же политика, binding-валидацию можно обойти прямым вызовом)
	// S4: email нормализуем как в AuthService (уникальность — LOWER(email)).
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if len(req.Password) < 12 || len(req.Password) > 72 {
		return nil, errors.New("password must be 12..72 characters")
	}
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

// allowedManage — кого какой создатель вправе вести (F4: тот же whitelist,
// что и на создании; на обновлении/удалении его не было — дыра эскалации).
var allowedManage = map[string]map[models.Role]bool{
	string(models.RoleSuperAdmin): {
		models.RoleFranchisor: true, models.RoleFranchisorManager: true,
		models.RoleDealer: true, models.RoleDealerManager: true,
	},
	string(models.RoleFranchisor): {
		models.RoleFranchisorManager: true, models.RoleDealer: true,
	},
	string(models.RoleFranchisorManager): {
		models.RoleDealer: true,
	},
	string(models.RoleDealer): {
		models.RoleDealerManager: true,
	},
}

func (s *UserService) UpdateEmployee(userID, tenantID uuid.UUID, req models.UpdateEmployeeRequest, callerRole string) (*models.User, error) {
	// S8: super_admin без сети грузит напрямую (иначе fail-closed ломал
	// управление: поиск в tenant Nil ничего не находил).
	var target *models.User
	var err error
	if callerRole == string(models.RoleSuperAdmin) {
		target, err = s.userRepo.GetUserByID(context.Background(), userID)
	} else {
		target, err = s.userRepo.FindUserByIDAndTenant(context.Background(), userID, tenantID)
	}
	if err != nil {
		return nil, errors.New("employee not found in your network")
	}
	// super_admin трогает только super_admin
	if target.Role == models.RoleSuperAdmin && callerRole != string(models.RoleSuperAdmin) {
		return nil, errors.New("permission denied: cannot manage super_admin")
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
		// F4: роль менять можно только в пределах своего whitelist
		if callerRole == "" {
			return nil, errors.New("permission denied: cannot change role")
		}
		if allowed, ok := allowedManage[callerRole]; !ok || !allowed[req.Role] {
			return nil, errors.New("permission denied: cannot assign this role")
		}
		updateData["role"] = req.Role
	}
	if req.ManagedBy != nil {
		// F4: ManagedBy — только из того же tenant (как на создании)
		if callerRole == "" {
			return nil, errors.New("permission denied: cannot reassign manager")
		}
		mgr, err := s.userRepo.GetUserByID(context.Background(), *req.ManagedBy)
		if err != nil {
			return nil, errors.New("managed_by user not found")
		}
		if tenantID != uuid.Nil && mgr.TenantID != nil && *mgr.TenantID != tenantID {
			return nil, errors.New("managed_by must be in same tenant")
		}
		if tenantID != uuid.Nil && mgr.TenantID == nil {
			return nil, errors.New("managed_by must be in same tenant")
		}
		updateData["managed_by"] = req.ManagedBy
	}

	if err := s.userRepo.UpdateUserFields(context.Background(), userID, updateData); err != nil {
		return nil, err
	}
	return s.userRepo.GetUserByID(context.Background(), userID)
}

func (s *UserService) DeleteEmployee(userID, tenantID uuid.UUID, callerRole string) error {
	var target *models.User
	var err error
	if callerRole == string(models.RoleSuperAdmin) {
		target, err = s.userRepo.GetUserByID(context.Background(), userID)
	} else {
		target, err = s.userRepo.FindUserByIDAndTenant(context.Background(), userID, tenantID)
	}
	if err != nil {
		return errors.New("employee not found in your network")
	}
	// super_admin удаляет только super_admin; остальные — только роли из своего whitelist
	if target.Role == models.RoleSuperAdmin && callerRole != string(models.RoleSuperAdmin) {
		return errors.New("permission denied: cannot manage super_admin")
	}
	if callerRole == "" {
		return errors.New("permission denied")
	}
	if allowed, ok := allowedManage[callerRole]; !ok || !allowed[target.Role] {
		return errors.New("permission denied: cannot delete user with this role")
	}
	return s.userRepo.DeleteUser(context.Background(), userID)
}

// === МЕТОДЫ ДЛЯ СУПЕР-АДМИНА ===

func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return s.userRepo.GetUserByID(ctx, userID)
}

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
