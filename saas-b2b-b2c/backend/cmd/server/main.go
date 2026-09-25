package main

import (
	"context"

	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"franchise-saas-backend/config"
	"franchise-saas-backend/internal/cache"
	"franchise-saas-backend/internal/database"
	"franchise-saas-backend/internal/docs"
	"franchise-saas-backend/internal/handlers"
	"franchise-saas-backend/internal/jobs"
	"franchise-saas-backend/internal/middleware"
	"franchise-saas-backend/internal/repository"
	"franchise-saas-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// isProdEnv — прод-режим: GIN_MODE=release или APP_ENV=production.
// Сиды, дефолты и прочий демо-мусор в проде запрещены.
func isProdEnv() bool {
	if gin.Mode() == gin.ReleaseMode {
		return true
	}
	return os.Getenv("APP_ENV") == "production"
}

// seedDemoData — демо-данные для dev/stage (никогда в проде, см. выше).
func seedDemoData(db *gorm.DB) {
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
	if err := database.SeedProductsAnalytics(db); err != nil {
		log.Printf("Seed products analytics error: %v", err)
	}
}

func main() {
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: config.yaml not found, using env vars")
	}
	viper.AutomaticEnv()

	// F5/F14: fail-closed на старте — без JWT_SECRET и DB_PASSWORD процесс
	// не слушает порт (раньше стартовал и отдавал /health 200 OK без секретов).
	_ = config.LoadConfig()

	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// RE-AUDIT: сиды демо-аккаунтов (включая super_admin) — только вне прода.
	// Раньше выполнялись при каждом старте везде, а сгенерированный пароль
	// печатался в stderr (credential в логах).
	// REAUDIT-3: демо-данные (включая super_admin) — только явный SEED_DEMO=true.
	// Прежний гейт "не prod" включал их в staging/k8s, где APP_ENV/GIN_MODE не заданы.
	if isProdEnv() {
		log.Printf("prod mode: skipping demo seeds")
	} else if !strings.EqualFold(strings.TrimSpace(os.Getenv("SEED_DEMO")), "true") {
		log.Println("demo seeds skipped (set SEED_DEMO=true to enable)")
	} else {
		seedDemoData(db)
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
	checklistService := services.NewChecklistService(checklistRepo).WithUserRepository(userRepo)
	adminService := services.NewAdminService(db)
	notifService := services.NewNotificationService(notifRepo)
	leadService := services.NewLeadService(leadRepo)
	scheduleService := services.NewScheduleService(scheduleRepo)
	kpiService := services.NewKPIService(db, kpiRepo, scheduleRepo)
	goalService := services.NewGoalService(goalRepo)
	alertService := services.NewAlertService(nil, notifRepo, db)

	c := cron.New(cron.WithSeconds())
	paymentJob := jobs.NewPaymentJob(adminService, notifService)
	_, _ = c.AddFunc("0 0 9 * * *", paymentJob.Run)
	logRotation := jobs.NewLogRotationJob(db)
	_, _ = c.AddFunc("0 0 3 * * *", logRotation.Run)
	_, _ = c.AddFunc("0 0 * * * *", func() {
		if _, err := authService.CleanupAuthSessions(context.Background()); err != nil {
			log.Printf("auth session cleanup failed: %v", err)
		}
	})
	c.Start()
	defer c.Stop()
	log.Println("Cron jobs started (payment 09:00, log rotation 03:00)")

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	checklistHandler := handlers.NewChecklistHandler(checklistService)
	adminHandler := handlers.NewAdminHandler(adminService)
	notifHandler := handlers.NewNotificationHandler(notifService)
	leadHandler := handlers.NewLeadHandler(leadService)
	kpiHandler := handlers.NewKPIHandler(db, kpiService, scheduleService, alertService)
	goalHandler := handlers.NewGoalHandler(goalService)

	middleware.SetIdentityResolver(func(ctx context.Context, userID, sessionID string, authVersion int64) (string, string, bool) {
		uid, err := uuid.Parse(userID)
		if err != nil {
			return "", "", false
		}
		sid, err := uuid.Parse(sessionID)
		if err != nil {
			return "", "", false
		}
		identity, err := authService.ResolveAccessIdentity(ctx, uid, sid, authVersion)
		if err != nil {
			return "", "", false
		}
		tenantID := ""
		if identity.TenantID != nil {
			tenantID = identity.TenantID.String()
		}
		return string(identity.Role), tenantID, true
	})

	r := gin.Default()
	// Только доверенные прокси (nginx) могут устанавливать X-Forwarded-For — защита от обхода rate-limit через подделку XFF
	_ = r.SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.1/32"})
	r.Use(middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		// RE-AUDIT: минимум для неаутентифицированного эндпоинта —
		// memory_mb/version светили внутренности для recon.
		status := "OK"
		code := http.StatusOK
		dbStatus := "ok"
		redisStatus := "ok"
		if err := db.Exec("SELECT 1").Error; err != nil {
			dbStatus = "down"
			status = "degraded"
			code = http.StatusServiceUnavailable
		}
		if cache.Client != nil {
			if err := cache.Client.Ping(c.Request.Context()).Err(); err != nil {
				redisStatus = "down"
				if status == "OK" {
					status = "degraded"
				}
			}
		} else {
			redisStatus = "not_configured"
		}
		c.JSON(code, gin.H{
			"status":    status,
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "1.0.0-stage1",
			"db":        dbStatus,
			"redis":     redisStatus,
		})
	})

	r.GET("/api/docs", docs.SwaggerHandler)

	api := r.Group("/api/v1")
	api.Use(middleware.RateLimit(100, time.Minute))
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now().Format(time.RFC3339)})
		})
		api.POST("/auth/refresh", authHandler.RefreshToken)
	}
	// Строгий лимит на auth: защита от брутфорса (F9: Redis-first, общий на
	// все реплики; in-memory RateLimit здесь врал при масштабировании).
	authLimited := api.Group("/auth")
	authLimited.Use(middleware.RateLimitAuth(10, time.Minute))
	{
		authLimited.POST("/register", authHandler.Register)
		authLimited.POST("/logout", middleware.CSRF(), authHandler.Logout)
	}
	authLoginLimited := api.Group("/auth")
	authLoginLimited.Use(middleware.RateLimitAuth(5, time.Minute))
	{
		authLoginLimited.POST("/login", authHandler.Login)
	}

	protected := api.Group("/")
	protected.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})
	protected.Use(middleware.AuthMiddleware())
	// F7: double-submit CSRF для cookie-сессий (Bearer-запросы exempt внутри).
	protected.Use(middleware.CSRF())
	protected.Use(middleware.LoggingMiddleware(userRepo))
	{
		// Роли для дашбордов дилера/салон-менеджера (общие для /dealer, /dashboard, /salon-manager).
		dealerDashRoles := middleware.RequireRole("dealer", "salon_manager", "franchiser", "franchiser_manager", "super_admin")
		// Роли для франшизных разделов.
		franchiserRoles := middleware.RequireRole("franchiser", "franchiser_manager", "super_admin")
		// HR/салоны: salon_manager никого не создаёт/не удаляет (см. F3/F4).
		hrRoles := middleware.RequireRole("franchiser", "franchiser_manager", "dealer", "super_admin")

		protected.GET("/auth/csrf", authHandler.GetCSRF)
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

		usersGroup := protected.Group("/users")
		usersGroup.Use(hrRoles)
		{
			usersGroup.GET("", userHandler.GetEmployees)
			usersGroup.POST("", userHandler.CreateEmployee)
			usersGroup.PUT("/:id", userHandler.UpdateEmployee)
			usersGroup.DELETE("/:id", userHandler.DeleteEmployee)
		}
		protected.GET("/users/me", userHandler.GetProfile)
		protected.PUT("/users/me", userHandler.UpdateProfile)

		protected.GET("/leads", leadHandler.GetMyLeads)
		protected.POST("/leads", leadHandler.CreateLead)
		protected.GET("/leads/:id", leadHandler.GetLeadByID)
		protected.PUT("/leads/:id/status", leadHandler.UpdateLeadStatus)
		protected.POST("/leads/:id/activities", leadHandler.AddActivity)

		salonsGroup := protected.Group("/salons")
		salonsGroup.Use(hrRoles)
		{
			salonsGroup.POST("", userHandler.CreateSalon)
			salonsGroup.GET("", userHandler.GetMySalons)
			salonsGroup.POST("/assign", userHandler.AssignManager)
			salonsGroup.PUT("/:id", userHandler.UpdateSalon)
			salonsGroup.DELETE("/:id", userHandler.DeleteSalon)
		}

		handlers.NewPlanHandler(protected, planService)
	}

	admin := api.Group("/admin")
	admin.Use(func(c *gin.Context) { c.Set("db", db); c.Next() })
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.CSRF())
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
	defer func() { _ = cache.Close() }()
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
