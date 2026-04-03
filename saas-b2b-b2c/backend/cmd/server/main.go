package main

import (
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/robfig/cron/v3"

	"franchise-saas-backend/internal/database"
	"franchise-saas-backend/internal/handlers"
	"franchise-saas-backend/internal/jobs"
	"franchise-saas-backend/internal/middleware"
	"franchise-saas-backend/internal/models"
	"franchise-saas-backend/internal/repository"
	"franchise-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func EnsureSuperAdmin(db *gorm.DB) {
	adminEmail := "admin@admin.ru"
	var user models.User

	if err := db.Where("email = ?", adminEmail).First(&user).Error; err != nil {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("qwerty"), bcrypt.DefaultCost)

		newUser := models.User{
			Email:        adminEmail,
			PasswordHash: string(hashedPassword),
			Role:         models.RoleSuperAdmin,
			FirstName:    "Super",
			LastName:     "Admin",
		}

		if err := db.Create(&newUser).Error; err != nil {
			log.Printf("Failed to create super admin: %v", err)
		} else {
			log.Println("Super Admin created: admin@admin.ru / qwerty")
		}
	}
}

func main() {
	// 1. Конфигурация
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: config.yaml not found, using env vars")
	}
	viper.AutomaticEnv()

	// 2. База данных
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// 2.5 Авто-создание суперадмина
	EnsureSuperAdmin(db)

	// 3. Инициализация репозиториев
	notifRepo := repository.NewNotificationRepository(db)
	userRepo := repository.NewUserRepository(db)
	leadRepo := repository.NewLeadRepository(db)
	kpiRepo := repository.NewKPIRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)

	// 4. Сервисы
	authService := services.NewAuthService(db)
	userService := services.NewUserService(db)
	checklistService := services.NewChecklistService(db)
	adminService := services.NewAdminService(db)
	notifService := services.NewNotificationService(notifRepo)
	leadService := services.NewLeadService(leadRepo)
	scheduleService := services.NewScheduleService(scheduleRepo)
	kpiService := services.NewKPIService(db, kpiRepo, scheduleRepo)

	// 5. Фоновые задачи
	c := cron.New()
	paymentJob := jobs.NewPaymentJob(adminService, notifService)
	c.AddFunc("0 0 9 * * *", paymentJob.Run)
	c.Start()
	defer c.Stop()
	log.Println("Cron jobs started")

	// 6. Хендлеры
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	checklistHandler := handlers.NewChecklistHandler(checklistService)
	adminHandler := handlers.NewAdminHandler(adminService)
	notifHandler := handlers.NewNotificationHandler(notifService)
	leadHandler := handlers.NewLeadHandler(leadService)
	kpiHandler := handlers.NewKPIHandler(kpiService, scheduleService)

	// 7. Роутер
	r := gin.Default()
	r.Use(CORSMiddleware())

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		status := "OK"
		if m.Alloc > 500*1024*1024 {
			status = "WARNING"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    status,
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "1.0.0-stage1",
			"memory_mb": m.Alloc / 1024 / 1024,
		})
	})

	// Публичные роуты
	api := r.Group("/api/v1")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/refresh", authHandler.RefreshToken)
	}

	// Защищенные роуты
	protected := api.Group("/")
	// !!! ВАЖНОЕ ИЗМЕНЕНИЕ: Добавляем DB в контекст !!!
	protected.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Сначала подключаем Middleware авторизации!
	protected.Use(middleware.AuthMiddleware())
	protected.Use(middleware.LoggingMiddleware(userRepo))
	{
		protected.GET("/auth/me", userHandler.GetProfile)
		protected.PUT("/auth/me", userHandler.UpdateProfile)
		protected.POST("/auth/change-password", userHandler.ChangePassword)

		// --- НОВОЕ: KPI и Расписание (Перенесено внутрь защищенного блока) ---
		protected.GET("/stats/my", kpiHandler.GetMyStats)
		protected.GET("/stats/salon", kpiHandler.GetSalonStats)
		protected.POST("/goals", kpiHandler.SetGoal) // Задать план

		protected.GET("/schedule", kpiHandler.GetSchedule)
		protected.POST("/schedule", kpiHandler.CreateEvent)
		protected.PUT("/schedule/:id/status", kpiHandler.UpdateEventStatus)

		protected.GET("/stats/team/analytics", kpiHandler.GetTeamAnalytics)
		protected.GET("/schedule/all", kpiHandler.GetAllSchedule)
		protected.POST("/schedule/manager", kpiHandler.CreateEventForManager)
		protected.PUT("/schedule/:id", kpiHandler.UpdateEvent)
		protected.DELETE("/schedule/:id", kpiHandler.DeleteEvent)
		// --------------------------------------------------------------------

		// Чек-листы
		protected.GET("/checklists", checklistHandler.GetChecklists)
		protected.POST("/checklists", checklistHandler.CreateChecklist)
		protected.PUT("/checklists/:id/status", checklistHandler.UpdateStatus)
		protected.GET("/checklists/:id", checklistHandler.GetChecklistByID)
		protected.PUT("/checklists/:id", checklistHandler.UpdateChecklist)
		protected.DELETE("/checklists/:id", checklistHandler.DeleteChecklist)
		protected.POST("/checklists/:id/complete", checklistHandler.CompleteChecklist)

		// Уведомления
		protected.GET("/notifications", notifHandler.GetMyNotifications)
		protected.POST("/notifications/:id/read", notifHandler.MarkAsRead)
		protected.POST("/notifications/read-all", notifHandler.MarkAllAsRead)

		// HR Модуль
		protected.GET("/users", userHandler.GetEmployees)
		protected.POST("/users", userHandler.CreateEmployee)
		protected.PUT("/users/:id", userHandler.UpdateEmployee)
		protected.DELETE("/users/:id", userHandler.DeleteEmployee)

		// CRM Модуль (Менеджер Салона)
		protected.GET("/leads", leadHandler.GetMyLeads)
		protected.POST("/leads", leadHandler.CreateLead)
		protected.GET("/leads/:id", leadHandler.GetLeadByID)
		protected.PUT("/leads/:id/status", leadHandler.UpdateLeadStatus)
		protected.POST("/leads/:id/activities", leadHandler.AddActivity)

		// УПРАВЛЕНИЕ САЛОНАМИ
		protected.POST("/salons", userHandler.CreateSalon)
		protected.GET("/salons", userHandler.GetMySalons)
		protected.POST("/salons/assign", userHandler.AssignManager)
		protected.PUT("/salons/:id", userHandler.UpdateSalon)
		protected.DELETE("/salons/:id", userHandler.DeleteSalon)
	}

	// Админ-панель
	admin := api.Group("/admin")
	admin.Use(func(c *gin.Context) { c.Set("db", db); c.Next() }) // Тоже добавляем для админки
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.SuperAdminMiddleware())
	{
		admin.GET("/stats", adminHandler.GetDashboardStats)
		admin.GET("/analytics", adminHandler.GetAnalytics)
		admin.GET("/product-analytics", adminHandler.GetProductAnalytics)
		admin.GET("/risks", adminHandler.GetRisks)
		admin.GET("/economics", adminHandler.GetUnitEconomics)
		admin.POST("/economics/marketing", adminHandler.UpdateMarketingSpend)

		admin.GET("/tenants", adminHandler.GetAllTenants)
		admin.GET("/tenants/payments", adminHandler.GetTenantsPaymentStatus)
		admin.POST("/tenants", adminHandler.CreateTenant)
		admin.PUT("/tenants/:id", adminHandler.UpdateTenant)
		admin.POST("/tenants/:id/block", adminHandler.BlockTenant)
		admin.DELETE("/tenants/:id", adminHandler.DeleteTenant)

		admin.GET("/plans", adminHandler.GetAllPlans)
		admin.POST("/plans", adminHandler.CreatePlan)
		admin.PUT("/plans/:id", adminHandler.UpdatePlan)
		admin.DELETE("/plans/:id", adminHandler.DeletePlan)

		// Invoices
		admin.POST("/invoices", adminHandler.CreateInvoice)
		admin.PUT("/invoices/:id/pay", adminHandler.PayInvoice)
	}

	// 8. Запуск сервера
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
