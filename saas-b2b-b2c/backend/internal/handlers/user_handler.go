package handlers

import (
	"log"
	"net/http"

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

	log.Printf("DEBUG GetUsers: User %s (Role: %s)", currentUser.Email, currentUser.Role)

	var users []models.User

	if string(currentUser.Role) == string(models.RoleSuperAdmin) {
		log.Println("DEBUG GetUsers: Path -> SUPER ADMIN (Global)")
		users, err = h.service.GetAllUsersGlobal()
	} else {
		log.Println("DEBUG GetUsers: Path -> STANDARD USER (Tenant)")
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

	log.Printf("DEBUG GetUsers: Returning %d users", len(users))
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	user, err := h.service.UpdateEmployee(userID, tid, req)
	if err != nil {
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

	if err := h.service.DeleteEmployee(userID, tid); err != nil {
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

// AssignManager
func (h *UserHandler) AssignManager(c *gin.Context) {
	var req struct {
		UserID  uuid.UUID `json:"user_id" binding:"required"`
		SalonID uuid.UUID `json:"salon_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("DEBUG AssignManager: Attempting to assign User %s to Salon %s", req.UserID, req.SalonID)

	if err := h.service.AssignManagerToSalon(c.Request.Context(), req.UserID, req.SalonID); err != nil {
		log.Printf("ERROR AssignManager: Failed to assign. Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("SUCCESS AssignManager: User %s assigned to Salon %s", req.UserID, req.SalonID)

	c.JSON(http.StatusOK, gin.H{"message": "Manager assigned to salon"})
}

// UpdateSalon
func (h *UserHandler) UpdateSalon(c *gin.Context) {
	idStr := c.Param("id")
	salonID, _ := uuid.Parse(idStr)

	var req struct {
		Name    string `json:"name" binding:"required"`
		Address string `json:"address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	salon, err := h.service.UpdateSalon(c.Request.Context(), salonID, req.Name, req.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, salon)
}

// DeleteSalon
func (h *UserHandler) DeleteSalon(c *gin.Context) {
	idStr := c.Param("id")
	salonID, _ := uuid.Parse(idStr)

	if err := h.service.DeleteSalon(c.Request.Context(), salonID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Salon deleted"})
}
