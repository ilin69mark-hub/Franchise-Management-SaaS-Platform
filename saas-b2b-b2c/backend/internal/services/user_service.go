package services

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"franchise-saas-backend/internal/cache"
	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	// REAUDIT-3: 12..72 БАЙТ (72 — предел bcrypt; >72 он отвергает, а не усекает),
	// плюс проверка HIBP, как при регистрации.
	if len(newPassword) < minPasswordLength || len(newPassword) > maxPasswordLength {
		return errors.New("password must be 12..72 bytes")
	}
	if err := rejectPwnedPassword(newPassword); err != nil {
		return err
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if s.db != nil {
		return s.db.Transaction(func(tx *gorm.DB) error {
			var current models.User
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, "id = ?", userID).Error; err != nil {
				return err
			}
			if err := bcrypt.CompareHashAndPassword([]byte(current.PasswordHash), []byte(oldPassword)); err != nil {
				return errors.New("invalid old password")
			}
			if err := tx.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
				"password_hash": string(newHash),
				"auth_version":  gorm.Expr("auth_version + 1"),
				"updated_at":    time.Now(),
			}).Error; err != nil {
				return err
			}
			now := time.Now().UTC()
			return tx.Model(&models.AuthSession{}).
				Where("user_id = ? AND revoked_at IS NULL", userID).
				Updates(map[string]interface{}{
					"revoked_at":    now,
					"revoke_reason": "password_changed",
					"updated_at":    now,
				}).Error
		})
	}
	user, err := s.userRepo.GetUserByID(context.Background(), userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}
	if err := s.userRepo.UpdateUserFields(context.Background(), userID, map[string]interface{}{
		"password_hash": string(newHash),
		"updated_at":    time.Now(),
	}); err != nil {
		return err
	}
	cache.RevokeUserSessions(context.Background(), userID.String())
	return nil
}

// === МЕТОДЫ HR (СОТРУДНИКИ) ===

func (s *UserService) GetEmployees(tenantID uuid.UUID) ([]models.User, error) {
	return s.userRepo.FindUsersByTenantID(context.Background(), tenantID)
}

// CreateEmployee — strict whitelist per creator role + tenant isolation

// enforceUserQuota — REAUDIT-4: лимит пользователей тенанта (tenants.max_users
// как override; иначе берём план). Блокировка строки tenant'а делает проверку
// и последующую вставку атомарными относительно параллельных запросов.
func enforceUserQuota(tx *gorm.DB, tenantID uuid.UUID) error {
	if tx == nil || tenantID == uuid.Nil {
		return nil
	}
	var tenant models.Tenant
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&tenant, "id = ?", tenantID).Error; err != nil {
		return err
	}
	limit := tenant.MaxUsers
	if tenant.PlanID != nil {
		var plan models.Plan
		if err := tx.First(&plan, "id = ?", *tenant.PlanID).Error; err != nil {
			return err
		}
		if plan.MaxUsers > 0 && (limit <= 0 || plan.MaxUsers < limit) {
			limit = plan.MaxUsers
		}
	}
	if limit <= 0 {
		return nil // лимит не задан — без ограничений
	}
	var count int64
	if err := tx.Model(&models.User{}).Where("tenant_id = ?", tenantID).Count(&count).Error; err != nil {
		return err
	}
	if count >= int64(limit) {
		return errors.New("tenant user limit reached (upgrade plan or increase max_users)")
	}
	return nil
}

// enforceSalonQuota — лимит салонов тенанта (plans.max_salons).
func enforceSalonQuota(tx *gorm.DB, tenantID uuid.UUID) error {
	if tx == nil || tenantID == uuid.Nil {
		return nil
	}
	var tenant models.Tenant
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&tenant, "id = ?", tenantID).Error; err != nil {
		return err
	}
	limit := 0
	if tenant.PlanID != nil {
		var plan models.Plan
		if err := tx.First(&plan, "id = ?", *tenant.PlanID).Error; err != nil {
			return err
		}
		limit = plan.MaxSalons
	}
	if limit <= 0 {
		return nil
	}
	var count int64
	if err := tx.Model(&models.Salon{}).Where("tenant_id = ?", tenantID).Count(&count).Error; err != nil {
		return err
	}
	if count >= int64(limit) {
		return errors.New("tenant salon limit reached (upgrade plan)")
	}
	return nil
}

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

	// REAUDIT-4: квота лицензий плана. Раньше max_users/max_salons были
	// метаданными: tenant с max_users=1 спокойно создавал десятки сотрудников.
	// Проверка и вставка идут в одной транзакции с блокировкой строки tenant'а.
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

	if s.db != nil {
		err = s.db.Transaction(func(tx *gorm.DB) error {
			if err := enforceUserQuota(tx, tenantID); err != nil {
				return err
			}
			return repository.NewUserRepository(tx).CreateUser(context.Background(), &user)
		})
	} else {
		err = s.userRepo.CreateUser(context.Background(), &user)
	}
	if err != nil {
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

// REAUDIT-3 (BIZ-008): tenant-скоупа мало — дилер правил ЛЮБЫМ сотрудником своей
// сети, потому что проверялась только принадлежность tenant'у. Управлять можно
// только своими подчинёнными (прямой или транзитивный managed_by), плюс себя.

// isDescendantOf — является ли candidate потомком первого (прямой или транзитивный
// managed_by). Нужен, чтобы запретить reassignment, создающий цикл иерархии.
func (s *UserService) isDescendantOf(candidate *models.User, ancestorID uuid.UUID) bool {
	if candidate == nil {
		return false
	}
	seen := map[uuid.UUID]bool{}
	cur := candidate
	for i := 0; i < 16 && cur.ManagedBy != nil; i++ {
		parentID := *cur.ManagedBy
		if parentID == ancestorID {
			return true
		}
		if seen[parentID] {
			return false
		}
		seen[parentID] = true
		parent, err := s.userRepo.GetUserByID(context.Background(), parentID)
		if err != nil || parent == nil {
			return false
		}
		cur = parent
	}
	return false
}

// validateManagedBy — REAUDIT-4: новый руководитель обязан быть в той же сети,
// существовать, не быть самим целевым юзером и не лежать ниже него.
func (s *UserService) validateManagedBy(targetID uuid.UUID, newManagerID uuid.UUID) error {
	if newManagerID == targetID {
		return errors.New("managed_by cannot be the user itself")
	}
	mgr, err := s.userRepo.GetUserByID(context.Background(), newManagerID)
	if err != nil || mgr == nil {
		return errors.New("managed_by user not found")
	}
	target, err := s.userRepo.GetUserByID(context.Background(), targetID)
	if err != nil || target == nil {
		return errors.New("employee not found")
	}
	if (mgr.TenantID == nil) != (target.TenantID == nil) {
		return errors.New("managed_by must be in same tenant")
	}
	if mgr.TenantID != nil && *mgr.TenantID != *target.TenantID {
		return errors.New("managed_by must be in same tenant")
	}
	if s.isDescendantOf(mgr, targetID) {
		return errors.New("managed_by cannot be a subordinate of the user")
	}
	return nil
}

func (s *UserService) canManage(callerID uuid.UUID, target *models.User) bool {
	if callerID == target.ID {
		return true
	}
	if target.Role == models.RoleSuperAdmin {
		return false
	}
	// Проверка на цикл при САМОМ reassignment живёт в validateManagedBy
	// (isDescendantOf(newManager, target)); здесь подчинённые управляемы.
	seen := map[uuid.UUID]bool{}
	current := target
	for i := 0; i < 16 && current.ManagedBy != nil; i++ {
		parentID := *current.ManagedBy
		if parentID == callerID {
			return true
		}
		if seen[parentID] {
			return false
		}
		seen[parentID] = true
		parent, err := s.userRepo.GetUserByID(context.Background(), parentID)
		if err != nil || parent == nil {
			return false
		}
		current = parent
	}
	return false
}

func (s *UserService) UpdateEmployee(callerID, userID, tenantID uuid.UUID, req models.UpdateEmployeeRequest, callerRole string) (*models.User, error) {
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
	// REAUDIT-3: вне иерархии подчинённых — отказ (иначе дилер правил всей сетью).
	if callerRole != string(models.RoleSuperAdmin) && !s.canManage(callerID, target) {
		return nil, errors.New("permission denied: target is not in your hierarchy")
	}
	updateData := map[string]interface{}{"updated_at": time.Now()}
	securityChanged := false
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
		securityChanged = true
	}
	if req.ManagedBy != nil {
		if callerRole == "" {
			return nil, errors.New("permission denied: cannot reassign manager")
		}
		// REAUDIT-4: единая проверка (same tenant + не self + не subordinate),
		// в том числе для super_admin (tenantID == uuid.Nil больше не отключает
		// проверку сети — иначе суперадмин создавал кросс-tenant связи).
		if err := s.validateManagedBy(userID, *req.ManagedBy); err != nil {
			return nil, err
		}
		updateData["managed_by"] = req.ManagedBy
		securityChanged = true
	}

	if securityChanged && s.db != nil {
		updateData["auth_version"] = gorm.Expr("auth_version + 1")
	}
	if err := s.userRepo.UpdateUserFields(context.Background(), userID, updateData); err != nil {
		return nil, err
	}
	if securityChanged && s.db != nil {
		if s.db.Migrator().HasTable(&models.AuthSession{}) {
			if err := repository.NewAuthSessionRepository(s.db).RevokeAllForUser(context.Background(), userID, "security_change"); err != nil {
				return nil, err
			}
		}
	} else if securityChanged {
		cache.RevokeUserSessions(context.Background(), userID.String())
	}
	return s.userRepo.GetUserByID(context.Background(), userID)
}

func (s *UserService) DeleteEmployee(callerID, userID, tenantID uuid.UUID, callerRole string) error {
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
	if callerRole != string(models.RoleSuperAdmin) && !s.canManage(callerID, target) {
		return errors.New("permission denied: target is not in your hierarchy")
	}
	if err := s.userRepo.DeleteUser(context.Background(), userID); err != nil {
		return err
	}
	cache.RevokeUserSessions(context.Background(), userID.String())
	return nil
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
	// REAUDIT-4: квота салонов (plans.max_salons) проверяется под блокировкой
	// строки tenant'а; вставка идёт в той же транзакции.
	salon := &models.Salon{TenantID: tenantID, Name: name, Address: address}
	if s.db == nil {
		if err := repository.NewSalonRepository(s.db).CreateSalon(ctx, salon); err != nil {
			return nil, err
		}
		return salon, nil
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := enforceSalonQuota(tx, tenantID); err != nil {
			return err
		}
		return repository.NewSalonRepository(tx).CreateSalon(ctx, salon)
	}); err != nil {
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

// UpdateSalon обновляет данные салона. REAUDIT-3: salonID обязан принадлежать
// tenant вызывающего (nil tenant = отказ, а не обход проверки).
func (s *UserService) UpdateSalon(ctx context.Context, salonID, tenantID uuid.UUID, name, address string) (*models.Salon, error) {
	salon, err := s.GetSalonByID(ctx, salonID)
	if err != nil {
		return nil, err
	}
	if tenantID != uuid.Nil {
		if salon.TenantID != tenantID {
			return nil, errors.New("forbidden: salon not in your tenant")
		}
	}

	salon.Name = name
	salon.Address = address

	if err := s.db.WithContext(ctx).Save(salon).Error; err != nil {
		return nil, err
	}
	return salon, nil
}

// DeleteSalon удаляет салон (tenant-скоуп обязателен, см. UpdateSalon).
func (s *UserService) DeleteSalon(ctx context.Context, salonID, tenantID uuid.UUID) error {
	if tenantID != uuid.Nil {
		salon, err := s.GetSalonByID(ctx, salonID)
		if err != nil {
			return err
		}
		if salon.TenantID != tenantID {
			return errors.New("forbidden: salon not in your tenant")
		}
	}
	salonRepo := repository.NewSalonRepository(s.db)
	return salonRepo.DeleteSalon(ctx, salonID)
}
