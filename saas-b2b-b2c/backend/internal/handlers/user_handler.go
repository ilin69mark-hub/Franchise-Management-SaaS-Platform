package handlers

import (
	"log"
	"net/http"
	"strings"

	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	service *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// GetEmployees
func (h *UserHandler) GetEmployees(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session, please login again"})
		return
	}

	var users []models.User

	if string(currentUser.Role) == string(models.RoleSuperAdmin) {
		users, err = h.service.GetAllUsersGlobal()
	} else {
		if currentUser.TenantID != nil {
			users, err = h.service.GetEmployees(*currentUser.TenantID)
		} else {
			users = []models.User{}
		}
	}

	if err != nil {
		log.Printf("ERROR GetUsers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetProfile
func (h *UserHandler) GetProfile(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	c.JSON(http.StatusOK, currentUser.ToProfileResponse())
}

// UpdateProfile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.service.UpdateProfile(currentUser.ID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated.ToProfileResponse())
}

// ChangePassword
func (h *UserHandler) ChangePassword(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.ChangePassword(currentUser.ID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Password changed"})
}

// CreateEmployee
func (h *UserHandler) CreateEmployee(c *gin.Context) {
	var req models.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	currentUser, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	// Определение TenantID
	targetTenantID := uuid.Nil

	if string(currentUser.Role) == string(models.RoleSuperAdmin) {
		// Супер админ может указать TenantID явно в запросе
		if req.TenantID != nil && *req.TenantID != uuid.Nil {
			targetTenantID = *req.TenantID
		}
		// Если не указан - останется Nil.
		// Сервис проверит: если роль franchiser -> вернет ошибку "tenant_id is required".
	} else {
		// Обычный пользователь всегда создает в рамках своего тенанта
		if currentUser.TenantID != nil {
			targetTenantID = *currentUser.TenantID
		}
	}

	// Передаем роль создателя как строку
	user, err := h.service.CreateEmployee(req, targetTenantID, currentUser.ID, string(currentUser.Role))
	if err != nil {
		log.Printf("ERROR CreateEmployee: %v", err)
		// REAUDIT-3: отказ по правам/дубль email = 403/409, а не 500.
		msg := err.Error()
		switch {
		case strings.Contains(msg, "permission denied"), strings.Contains(msg, "invalid role"):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case strings.Contains(msg, "limit reached"):
			c.JSON(http.StatusPaymentRequired, gin.H{"error": msg})
		case strings.Contains(msg, "already"), strings.Contains(msg, "duplicate"):
			// REAUDIT-4: нейтральный ответ (409 без признака "email занят" —
			// иначе сотрудник любого тенанта перечислял чужие email'ы).
			c.JSON(http.StatusConflict, gin.H{"error": "could not create employee"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create employee"})
		}
		return
	}
	c.JSON(http.StatusCreated, user)
}

// UpdateEmployee
func (h *UserHandler) UpdateEmployee(c *gin.Context) {
	idStr := c.Param("id")
	userID, _ := uuid.Parse(idStr)
	var req models.UpdateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	currentUser, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	var tid uuid.UUID
	if currentUser.TenantID != nil {
		tid = *currentUser.TenantID
	}
	// F4: запрет самоповышения — свою роль через HR-путь менять нельзя
	if userID == currentUser.ID && req.Role != "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot change own role"})
		return
	}

	user, err := h.service.UpdateEmployee(currentUser.ID, userID, tid, req, string(currentUser.Role))
	if err != nil {
		// REAUDIT-3: отказ по правам = 403, а не 500.
		if strings.Contains(err.Error(), "permission denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// DeleteEmployee
func (h *UserHandler) DeleteEmployee(c *gin.Context) {
	idStr := c.Param("id")
	userID, _ := uuid.Parse(idStr)
	currentUser, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	var tid uuid.UUID
	if currentUser.TenantID != nil {
		tid = *currentUser.TenantID
	}
	// F4: себя через HR-удаление удалять нельзя
	if userID == currentUser.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete self"})
		return
	}

	if err := h.service.DeleteEmployee(currentUser.ID, userID, tid, string(currentUser.Role)); err != nil {
		// REAUDIT-3: отказ по правам = 403, отсутствие цели = 404.
		if strings.Contains(err.Error(), "permission denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

// === УПРАВЛЕНИЕ САЛОНАМИ ===

// CreateSalon
func (h *UserHandler) CreateSalon(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	if currentUser.TenantID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not assigned to tenant"})
		return
	}

	var req struct {
		Name    string `json:"name" binding:"required"`
		Address string `json:"address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	salon, err := h.service.CreateSalon(c.Request.Context(), *currentUser.TenantID, req.Name, req.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, salon)
}

// GetMySalons
func (h *UserHandler) GetMySalons(c *gin.Context) {
	currentUser, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusOK, []models.Salon{})
		return
	}
	if currentUser.TenantID == nil {
		c.JSON(http.StatusOK, []models.Salon{})
		return
	}

	salons, err := h.service.GetMySalons(c.Request.Context(), *currentUser.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, salons)
}

// AssignManager — tenant-isolated
func (h *UserHandler) AssignManager(c *gin.Context) {
	var req struct {
		UserID  uuid.UUID `json:"user_id" binding:"required"`
		SalonID uuid.UUID `json:"salon_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	currentUser, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	// S6: назначать менеджера может только супер-админ или франчайзер
	// (менеджер/дилер не имеют права менять привязку салонов).
	switch currentUser.Role {
	case models.RoleSuperAdmin, models.RoleFranchisor:
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: role cannot assign managers"})
		return
	}
	// tenant isolation: both user and salon must be in caller's tenant (except super_admin)
	if currentUser.Role != "super_admin" && currentUser.TenantID != nil {
		// verify salon belongs to tenant and target user belongs to tenant
		salon, err := h.service.GetSalonByID(c.Request.Context(), req.SalonID)
		if err != nil || salon.TenantID != *currentUser.TenantID {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: salon not in your network"})
			return
		}
		targetUser, err := h.service.GetUserByID(c.Request.Context(), req.UserID)
		if err != nil || targetUser.TenantID == nil || *targetUser.TenantID != *currentUser.TenantID {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: user not in your network"})
			return
		}
	}

	if err := h.service.AssignManagerToSalon(c.Request.Context(), req.UserID, req.SalonID); err != nil {
		log.Printf("ERROR AssignManager: Failed to assign. Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("SUCCESS AssignManager: User %s assigned to Salon %s", req.UserID, req.SalonID)

	c.JSON(http.StatusOK, gin.H{"message": "Manager assigned to salon"})
}

// UpdateSalon — tenant-isolated
func (h *UserHandler) UpdateSalon(c *gin.Context) {
	idStr := c.Param("id")
	salonID, _ := uuid.Parse(idStr)
	currentUser, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	// verify salon belongs to tenant
	salonCheck, err := h.service.GetSalonByID(c.Request.Context(), salonID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Salon not found"})
		return
	}
	// REAUDIT-3: nil tenant = отказ (раньше проверка просто пропускалась).
	if currentUser.Role != "super_admin" && (currentUser.TenantID == nil || salonCheck.TenantID != *currentUser.TenantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req struct {
		Name    string `json:"name" binding:"required"`
		Address string `json:"address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	salon, err := h.service.UpdateSalon(c.Request.Context(), salonID, callerTenantID(currentUser), req.Name, req.Address)
	if err != nil {
		// REAUDIT-3: tenant-скоуп даёт 403, а не 500.
		if strings.Contains(err.Error(), "forbidden") {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, salon)
}

// DeleteSalon — tenant-isolated
func (h *UserHandler) DeleteSalon(c *gin.Context) {
	idStr := c.Param("id")
	salonID, _ := uuid.Parse(idStr)
	currentUser, err := getCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	salonCheck, err := h.service.GetSalonByID(c.Request.Context(), salonID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Salon not found"})
		return
	}
	// REAUDIT-3: nil tenant = отказ (раньше проверка просто пропускалась).
	if currentUser.Role != "super_admin" && (currentUser.TenantID == nil || salonCheck.TenantID != *currentUser.TenantID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if err := h.service.DeleteSalon(c.Request.Context(), salonID, callerTenantID(currentUser)); err != nil {
		// REAUDIT-3: tenant-скоуп даёт 403, а не 500.
		if strings.Contains(err.Error(), "forbidden") {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Salon deleted"})
}

// callerTenantID — tenant вызывающего; uuid.Nil = нет tenant (super_admin/legacy),
// и тогда tenant-скоуп в сервисе трактуется как "проверять нечего" и вызывающий
// обязан быть super_admin (это проверяет RequireRole).
func callerTenantID(user *models.User) uuid.UUID {
	if user.TenantID == nil {
		return uuid.Nil
	}
	return *user.TenantID
}
