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
		// Роли для дашбордов дилера/салон-менеджера (общие для /dealer, /dashboard, /salon-manager).
		dealerDashRoles := middleware.RequireRole("dealer", "salon_manager", "franchiser", "franchiser_manager", "super_admin")
		// Роли для франшизных разделов.
		franchiserRoles := middleware.RequireRole("franchiser", "franchiser_manager", "super_admin")

		protected.GET("/auth/me", userHandler.GetProfile)
		protected.PUT("/auth/me", userHandler.UpdateProfile)
		protected.POST("/auth/change-password", userHandler.ChangePassword)
		protected.POST("/auth/logout", authHandler.Logout)

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

		protected.GET("/alerts", kpiHandler.GetAlerts)
		protected.PATCH("/alerts/:id/read", kpiHandler.MarkAlertRead)

		// Dashboard Main (Salon Manager)
		dashboardGroup := protected.Group("/dashboard")
		dashboardGroup.Use(dealerDashRoles)
		{
			dashboardGroup.GET("/main", kpiHandler.GetDashboardMain)
			dashboardGroup.GET("/funnel", kpiHandler.GetDashboardFunnel)
			dashboardGroup.GET("/team", kpiHandler.GetDashboardTeam)
			dashboardGroup.GET("/team/:id/history", kpiHandler.GetSalesRepHistory)
			dashboardGroup.GET("/products", kpiHandler.GetDashboardProducts)
		}

		salonManagerGroup := protected.Group("/salon-manager")
		salonManagerGroup.Use(dealerDashRoles)
		{
			salonManagerGroup.GET("/top-bar", kpiHandler.GetTopBar)
		}

		// Dashboard Dealer
		dealerGroup := protected.Group("/dealer")
		dealerGroup.Use(dealerDashRoles)
		{
			dealerGroup.GET("/summary", kpiHandler.GetDealerSummary)
			dealerGroup.GET("/finance", kpiHandler.GetDealerFinance)
			dealerGroup.GET("/funnel", kpiHandler.GetDealerFunnel)
			dealerGroup.GET("/products", kpiHandler.GetDealerProducts)
			dealerGroup.GET("/tasks", kpiHandler.GetDealerTasks)
			dealerGroup.PATCH("/tasks/:id", kpiHandler.UpdateDealerTask)
			dealerGroup.GET("/requests", kpiHandler.GetDealerRequests)
			dealerGroup.POST("/requests", kpiHandler.CreateDealerRequest)
			dealerGroup.GET("/marketing-budget", kpiHandler.GetDealerMarketingBudget)
			dealerGroup.GET("/alerts", kpiHandler.GetDealerAlerts)
			dealerGroup.PATCH("/alerts/:id/read", kpiHandler.MarkDealerAlertRead)
			dealerGroup.PATCH("/alerts/read-all", kpiHandler.MarkAllDealerAlertsRead)
		}

		// Dashboard Franchiser
		franchiserGroup := protected.Group("/franchiser")
		franchiserGroup.Use(franchiserRoles)
		{
			franchiserGroup.GET("/summary", kpiHandler.GetFranchiserSummary)
			franchiserGroup.GET("/network", kpiHandler.GetFranchiserNetwork)
			franchiserGroup.GET("/network/territories", kpiHandler.GetTerritoriesHeatmap)
			franchiserGroup.GET("/health", kpiHandler.GetFranchiserHealth)
			franchiserGroup.GET("/team", kpiHandler.GetFranchiserTeam)
			franchiserGroup.GET("/team/:id/dynamics", kpiHandler.GetManagerDynamics)
			franchiserGroup.GET("/team/:id/dealers", kpiHandler.GetManagerDealers)
			franchiserGroup.POST("/team/plans", kpiHandler.SetManagerPlans)
			franchiserGroup.GET("/team/plans", kpiHandler.GetManagerPlans)
			franchiserGroup.GET("/dealers", kpiHandler.GetFranchiserDealers)
			franchiserGroup.GET("/dealers/health", kpiHandler.GetDealersHealth)
			franchiserGroup.GET("/dealers/migration", kpiHandler.GetDealersMigration)
			franchiserGroup.GET("/dealers/system-issues", kpiHandler.GetSystemIssues)
			franchiserGroup.GET("/dealers/geography", kpiHandler.GetDealersGeography)
			franchiserGroup.GET("/dealers/marketing-roi", kpiHandler.GetMarketingROI)
			franchiserGroup.GET("/requests", kpiHandler.GetFranchiserRequests)
			franchiserGroup.GET("/alerts", kpiHandler.GetFranchiserAlerts)
			franchiserGroup.PATCH("/alerts/:id/read", kpiHandler.MarkFranchiserAlertRead)
			franchiserGroup.PATCH("/alerts/read-all", kpiHandler.MarkAllFranchiserAlertsRead)
			franchiserGroup.PATCH("/alerts/:id/assign", kpiHandler.AssignAlert)
			franchiserGroup.GET("/alert-settings", kpiHandler.GetAlertSettings)
			franchiserGroup.PUT("/alert-settings", kpiHandler.UpdateAlertSettings)

			// Report B2B
			franchiserGroup.GET("/report/data", kpiHandler.GetReportData)
			franchiserGroup.POST("/report/generate-pdf", kpiHandler.GeneratePDF)
			franchiserGroup.POST("/report/send", kpiHandler.SendReport)
			franchiserGroup.GET("/report/history", kpiHandler.GetReportHistory)
			franchiserGroup.POST("/report/draft", kpiHandler.SaveDraft)
			franchiserGroup.GET("/report/draft", kpiHandler.GetDraft)
		}

		managerGroup := protected.Group("/manager")
		managerGroup.Use(franchiserRoles)
		{
			managerGroup.GET("/targets", kpiHandler.GetManagerTargets)
		}

		// Dashboard Territory Manager
		territoryGroup := protected.Group("/territory")
		territoryGroup.Use(franchiserRoles)
		{
			territoryGroup.GET("/summary", kpiHandler.GetTerritorySummary)
			territoryGroup.GET("/funnel", kpiHandler.GetTerritoryFunnel)
			territoryGroup.GET("/planfact", kpiHandler.GetTerritoryPlanFact)
			territoryGroup.GET("/communications", kpiHandler.GetTerritoryCommunications)
			territoryGroup.GET("/benchmarks", kpiHandler.GetTerritoryBenchmarks)
		}

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
