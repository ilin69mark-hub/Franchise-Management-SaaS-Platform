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

	// Seed data
	if err := database.SeedUsers(db); err != nil {
		log.Printf("Seed users error: %v", err)
	}
	if err := database.SeedGoals(db); err != nil {
		log.Printf("Seed goals error: %v", err)
	}
	if err := database.SeedChecklists(db); err != nil {
		log.Printf("Seed checklists error: %v", err)
	}
	if err := database.SeedAlerts(db); err != nil {
		log.Printf("Seed alerts error: %v", err)
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
	checklistRepo := repository.NewChecklistRepository(db)

	authService := services.NewAuthService(db)
	userService := services.NewUserService(db)
	checklistService := services.NewChecklistService(checklistRepo)
	adminService := services.NewAdminService(db)
	notifService := services.NewNotificationService(notifRepo)
	leadService := services.NewLeadService(leadRepo)
	scheduleService := services.NewScheduleService(scheduleRepo)
	kpiService := services.NewKPIService(db, kpiRepo, scheduleRepo)
	goalService := services.NewGoalService(goalRepo)
	alertService := services.NewAlertService(nil, notifRepo, db)

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
	kpiHandler := handlers.NewKPIHandler(db, kpiService, scheduleService, alertService)
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
		protected.PUT("/goals/:id", goalHandler.UpdateGoal)

		protected.GET("/schedule", kpiHandler.GetSchedule)
		protected.POST("/schedule", kpiHandler.CreateEvent)
		protected.PUT("/schedule/:id/status", kpiHandler.UpdateEventStatus)

		// Dashboard Main (Salon Manager)
		protected.GET("/dashboard/main", kpiHandler.GetDashboardMain)
		protected.GET("/salon-manager/top-bar", kpiHandler.GetTopBar)
		protected.GET("/dashboard/funnel", kpiHandler.GetDashboardFunnel)
		protected.GET("/dashboard/team", kpiHandler.GetDashboardTeam)
		protected.GET("/dashboard/team/:id/history", kpiHandler.GetSalesRepHistory)
		protected.GET("/dashboard/products", kpiHandler.GetDashboardProducts)
		protected.GET("/alerts", kpiHandler.GetAlerts)
		protected.PATCH("/alerts/:id/read", kpiHandler.MarkAlertRead)
		protected.GET("/manager/targets", kpiHandler.GetManagerTargets)

		// Dashboard Dealer
		protected.GET("/dealer/summary", kpiHandler.GetDealerSummary)
		protected.GET("/dealer/finance", kpiHandler.GetDealerFinance)
		protected.GET("/dealer/funnel", kpiHandler.GetDealerFunnel)
		protected.GET("/dealer/products", kpiHandler.GetDealerProducts)
		protected.GET("/dealer/tasks", kpiHandler.GetDealerTasks)
		protected.PATCH("/dealer/tasks/:id", kpiHandler.UpdateDealerTask)
		protected.GET("/dealer/requests", kpiHandler.GetDealerRequests)
		protected.POST("/dealer/requests", kpiHandler.CreateDealerRequest)
		protected.GET("/dealer/marketing-budget", kpiHandler.GetDealerMarketingBudget)
		protected.GET("/dealer/alerts", kpiHandler.GetDealerAlerts)
		protected.PATCH("/dealer/alerts/:id/read", kpiHandler.MarkDealerAlertRead)
		protected.PATCH("/dealer/alerts/read-all", kpiHandler.MarkAllDealerAlertsRead)

		// Dashboard Franchiser
		protected.GET("/franchiser/summary", kpiHandler.GetFranchiserSummary)
		protected.GET("/franchiser/network", kpiHandler.GetFranchiserNetwork)
		protected.GET("/franchiser/network/territories", kpiHandler.GetTerritoriesHeatmap)
		protected.GET("/franchiser/health", kpiHandler.GetFranchiserHealth)
		protected.GET("/franchiser/team", kpiHandler.GetFranchiserTeam)
		protected.GET("/franchiser/team/:id/dynamics", kpiHandler.GetManagerDynamics)
		protected.GET("/franchiser/team/:id/dealers", kpiHandler.GetManagerDealers)
		protected.POST("/franchiser/team/plans", kpiHandler.SetManagerPlans)
		protected.GET("/franchiser/team/plans", kpiHandler.GetManagerPlans)
		protected.GET("/franchiser/dealers", kpiHandler.GetFranchiserDealers)
		protected.GET("/franchiser/dealers/health", kpiHandler.GetDealersHealth)
		protected.GET("/franchiser/dealers/migration", kpiHandler.GetDealersMigration)
		protected.GET("/franchiser/dealers/system-issues", kpiHandler.GetSystemIssues)
		protected.GET("/franchiser/dealers/geography", kpiHandler.GetDealersGeography)
		protected.GET("/franchiser/dealers/marketing-roi", kpiHandler.GetMarketingROI)
		protected.GET("/franchiser/requests", kpiHandler.GetFranchiserRequests)
		protected.GET("/franchiser/alerts", kpiHandler.GetFranchiserAlerts)
		protected.PATCH("/franchiser/alerts/:id/read", kpiHandler.MarkFranchiserAlertRead)
		protected.PATCH("/franchiser/alerts/read-all", kpiHandler.MarkAllFranchiserAlertsRead)
		protected.PATCH("/franchiser/alerts/:id/assign", kpiHandler.AssignAlert)
		protected.GET("/franchiser/alert-settings", kpiHandler.GetAlertSettings)
		protected.PUT("/franchiser/alert-settings", kpiHandler.UpdateAlertSettings)

		// Report B2B
		protected.GET("/franchiser/report/data", kpiHandler.GetReportData)
		protected.POST("/franchiser/report/generate-pdf", kpiHandler.GeneratePDF)
		protected.POST("/franchiser/report/send", kpiHandler.SendReport)
		protected.GET("/franchiser/report/history", kpiHandler.GetReportHistory)
		protected.POST("/franchiser/report/draft", kpiHandler.SaveDraft)
		protected.GET("/franchiser/report/draft", kpiHandler.GetDraft)

		// Dashboard Territory Manager
		protected.GET("/territory/summary", kpiHandler.GetTerritorySummary)
		protected.GET("/territory/funnel", kpiHandler.GetTerritoryFunnel)
		protected.GET("/territory/planfact", kpiHandler.GetTerritoryPlanFact)
		protected.GET("/territory/communications", kpiHandler.GetTerritoryCommunications)
		protected.GET("/territory/benchmarks", kpiHandler.GetTerritoryBenchmarks)

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
		protected.GET("/users/me", userHandler.GetProfile)
		protected.PUT("/users/me", userHandler.UpdateProfile)

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
		admin.GET("/tenants/:id", adminHandler.GetTenantByID)
		admin.POST("/tenants", adminHandler.CreateTenant)
		admin.PUT("/tenants/:id", adminHandler.UpdateTenant)
		admin.POST("/tenants/:id/block", adminHandler.BlockTenant)
		admin.POST("/tenants/:id/unblock", adminHandler.UnblockTenant)
		admin.DELETE("/tenants/:id", adminHandler.DeleteTenant)

		admin.GET("/plans", adminHandler.GetAllPlans)
		admin.POST("/plans", adminHandler.CreatePlan)
		admin.PUT("/plans/:id", adminHandler.UpdatePlan)
		admin.DELETE("/plans/:id", adminHandler.DeletePlan)

		admin.POST("/invoices", adminHandler.CreateInvoice)
		admin.GET("/invoices", adminHandler.GetAllInvoices)
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