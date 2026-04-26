package main

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/robfig/cron/v3"

	"franchise-saas-backend/internal/cache"
	"franchise-saas-backend/internal/database"
	"franchise-saas-backend/internal/docs"
	"franchise-saas-backend/internal/handlers"
	"franchise-saas-backend/internal/jobs"
	"franchise-saas-backend/internal/middleware"
	"franchise-saas-backend/internal/repository"
	"franchise-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: config.yaml not found, using env vars")
	}
	viper.AutomaticEnv()

	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	_, err = cache.ConnectRedis()
	if err != nil {
		log.Printf("Warning: Redis not available, rate limiting will use in-memory store")
	}

	notifRepo := repository.NewNotificationRepository(db)
	userRepo := repository.NewUserRepository(db)
	leadRepo := repository.NewLeadRepository(db)
	kpiRepo := repository.NewKPIRepository(db)
	scheduleRepo := repository.NewScheduleRepository(db)
	goalRepo := repository.NewGoalRepository(db)
	planRepo := repository.NewPlanRepository(db)
	planService := services.NewPlanService(planRepo)

	authService := services.NewAuthService(db)
	userService := services.NewUserService(db)
	checklistService := services.NewChecklistService(db)
	adminService := services.NewAdminService(db)
	notifService := services.NewNotificationService(notifRepo)
	leadService := services.NewLeadService(leadRepo)
	scheduleService := services.NewScheduleService(scheduleRepo)
	kpiService := services.NewKPIService(db, kpiRepo, scheduleRepo)
	goalService := services.NewGoalService(goalRepo)

	c := cron.New()
	paymentJob := jobs.NewPaymentJob(adminService, notifService)
	c.AddFunc("0 0 9 * * *", paymentJob.Run)
	c.Start()
	defer c.Stop()
	log.Println("Cron jobs started")

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	checklistHandler := handlers.NewChecklistHandler(checklistService)
	adminHandler := handlers.NewAdminHandler(adminService)
	notifHandler := handlers.NewNotificationHandler(notifService)
	leadHandler := handlers.NewLeadHandler(leadService)
	kpiHandler := handlers.NewKPIHandler(kpiService, scheduleService)
	goalHandler := handlers.NewGoalHandler(goalService)

	r := gin.Default()
	r.Use(middleware.CORS())

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

	r.GET("/api/docs", docs.SwaggerHandler)

	api := r.Group("/api/v1")
	api.Use(middleware.RateLimit(100, time.Minute))
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now().Format(time.RFC3339)})
		})
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/refresh", authHandler.RefreshToken)
	}

	protected := api.Group("/")
	protected.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})
	protected.Use(middleware.AuthMiddleware())
	protected.Use(middleware.LoggingMiddleware(userRepo))
	{
		protected.GET("/auth/me", userHandler.GetProfile)
		protected.PUT("/auth/me", userHandler.UpdateProfile)
		protected.POST("/auth/change-password", userHandler.ChangePassword)

		protected.GET("/stats/my", kpiHandler.GetMyStats)
		protected.GET("/stats/salon", kpiHandler.GetSalonStats)
		protected.POST("/goals", goalHandler.SetGoal)
		protected.GET("/goals", goalHandler.GetGoal)
		protected.GET("/goals/visible", goalHandler.GetVisibleGoals)
		protected.GET("/goals/by-date/:date", goalHandler.GetGoalByDate)
		protected.DELETE("/goals/:id", goalHandler.DeleteGoal)

		protected.GET("/schedule", kpiHandler.GetSchedule)
		protected.POST("/schedule", kpiHandler.CreateEvent)
		protected.PUT("/schedule/:id/status", kpiHandler.UpdateEventStatus)

		protected.GET("/stats/team/analytics", kpiHandler.GetTeamAnalytics)
		protected.GET("/schedule/all", kpiHandler.GetAllSchedule)
		protected.POST("/schedule/manager", kpiHandler.CreateEventForManager)
		protected.PUT("/schedule/:id", kpiHandler.UpdateEvent)
		protected.DELETE("/schedule/:id", kpiHandler.DeleteEvent)

		protected.GET("/checklists", checklistHandler.GetChecklists)
		protected.POST("/checklists", checklistHandler.CreateChecklist)
		protected.PUT("/checklists/:id/status", checklistHandler.UpdateStatus)
		protected.GET("/checklists/:id", checklistHandler.GetChecklistByID)
		protected.PUT("/checklists/:id", checklistHandler.UpdateChecklist)
		protected.DELETE("/checklists/:id", checklistHandler.DeleteChecklist)
		protected.POST("/checklists/:id/complete", checklistHandler.CompleteChecklist)

		protected.GET("/notifications", notifHandler.GetMyNotifications)
		protected.POST("/notifications/:id/read", notifHandler.MarkAsRead)
		protected.POST("/notifications/read-all", notifHandler.MarkAllAsRead)

		protected.GET("/users", userHandler.GetEmployees)
		protected.POST("/users", userHandler.CreateEmployee)
		protected.PUT("/users/:id", userHandler.UpdateEmployee)
		protected.DELETE("/users/:id", userHandler.DeleteEmployee)

		protected.GET("/leads", leadHandler.GetMyLeads)
		protected.POST("/leads", leadHandler.CreateLead)
		protected.GET("/leads/:id", leadHandler.GetLeadByID)
		protected.PUT("/leads/:id/status", leadHandler.UpdateLeadStatus)
		protected.POST("/leads/:id/activities", leadHandler.AddActivity)

		protected.POST("/salons", userHandler.CreateSalon)
		protected.GET("/salons", userHandler.GetMySalons)
		protected.POST("/salons/assign", userHandler.AssignManager)
		protected.PUT("/salons/:id", userHandler.UpdateSalon)
		protected.DELETE("/salons/:id", userHandler.DeleteSalon)

		handlers.NewPlanHandler(protected, planService)
	}

	admin := api.Group("/admin")
	admin.Use(func(c *gin.Context) { c.Set("db", db); c.Next() })
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

		admin.POST("/invoices", adminHandler.CreateInvoice)
		admin.PUT("/invoices/:id/pay", adminHandler.PayInvoice)
	}

	port := viper.GetString("server.port")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on port %s", port)
	defer cache.Close()
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}