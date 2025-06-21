package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time" // Added for cron.WithLocation(time.UTC)

	"github.com/gin-contrib/cors" // Import CORS middleware
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3" // Added for cron jobs
	"github.com/zayyadi/finance-tracker/internal/database"
	"github.com/zayyadi/finance-tracker/internal/handlers"
	"github.com/zayyadi/finance-tracker/internal/models" // Added for direct AutoMigrate
	"github.com/zayyadi/finance-tracker/internal/services"
)

func main() {
	// Load .env file if it exists (useful for local development)
	err := godotenv.Load()
	if err == nil { // Can be nil if file doesn't exist, which is fine
		log.Println("Loaded .env file")
	} else {
		log.Printf("Warning: Could not load .env file (this is normal if not present): %v\n", err)
	}

	// Initialize database connection
	// os.Setenv("DB_HOST", "localhost")
	// os.Setenv("DB_USER", "testuser")
	// os.Setenv("DB_PASSWORD", "testpassword")
	// os.Setenv("DB_NAME", "finance_tracker_db")
	// os.Setenv("DB_PORT", "5432")
	// os.Setenv("DB_SSLMODE", "disable")

	// os.Setenv("JWT_SECRET_KEY", "your-super-secret-jwt-key-that-is-very-long-and-random")
	// os.Setenv("OPENROUTER_API_KEY", "YOUR_DUMMY_OPENROUTER_API_KEY_FOR_TESTING")

	if err := database.ConnectDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	db := database.GetDB()
	if db == nil {
		log.Fatalf("Failed to get GORM DB instance after connecting")
	}

	// Perform Auto-Migration directly here
	log.Println("Starting GORM auto-migration...")
	err = db.AutoMigrate(
		&models.User{}, // Ensure User table is created first if others depend on it
		&models.Income{},
		&models.Expense{},
		&models.Savings{},
		&models.Debt{},
		&models.FinancialSummary{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate GORM models: %v", err)
	} else {
		log.Println("GORM models auto-migrated successfully.")
	}

	defer database.CloseDB()

	router := gin.Default()

	// CORS Middleware Configuration
	// This allows your Vue dev server (e.g., http://localhost:5173) to make requests
	// to your Go backend (e.g., http://localhost:8080).
	// For production, you might want to restrict origins.
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173", "http://127.0.0.1:5173"} // Add your Vue dev server URL
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(config))

	// HTML template loading and static file serving for templates are removed.
	// The frontend will be a separate JS application.

	// Instantiate services
	// userService := services.NewUserService(db) // Removed UserService
	incomeService := services.NewIncomeService(db)
	expenseService := services.NewExpenseService(db)
	savingsService := services.NewSavingsService(db)
	debtService := services.NewDebtService(db)
	summaryService := services.NewSummaryService(db)
	aiAdviceService := services.NewAIAdviceService()
	reportService := services.NewReportService(incomeService, expenseService)
	notificationService := services.NewNotificationService(db) // Instantiate NotificationService
	analyticsService := services.NewAnalyticsService(db)    // New AnalyticsService

	// Instantiate handlers
	// authHandler := handlers.NewAuthHandler(userService) // Removed AuthHandler
	incomeHandler := handlers.NewIncomeHandler(incomeService, summaryService) // Added summaryService
	expenseHandler := handlers.NewExpenseHandler(expenseService, summaryService) // Added summaryService
	savingsHandler := handlers.NewSavingsHandler(savingsService)
	debtHandler := handlers.NewDebtHandler(debtService)
	summaryHandler := handlers.NewSummaryHandler(summaryService)
	aiAdviceHandler := handlers.NewAIAdviceHandler(aiAdviceService, summaryService)
	reportHandler := handlers.NewReportHandler(reportService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService) // New AnalyticsHandler
	// viewHandler := handlers.NewViewHandler() // Removed as Go no longer serves HTML pages

	// Frontend Page Routes are removed.
	// These will be handled by your JavaScript framework's router.
	// Example: router.GET("/", viewHandler.ShowHomePage) - REMOVED
	// Example: router.GET("/dashboard", viewHandler.ShowDashboardPage) - REMOVED

	// apiPublic := router.Group("/api") // This group is effectively empty now
	// {
	// apiPublic.POST("/register", authHandler.RegisterUser) // Removed
	// apiPublic.POST("/login", authHandler.LoginUser)       // Removed
	// }

	// All API routes are now under /api/v1 and are public due to AuthMiddleware removal.
	// For this project, assuming all data modification/retrieval requires auth.
	// If there were public info APIs, they would remain in a non-protected group.

	// API routes are now directly under /api/v1 (no auth middleware)
	apiV1 := router.Group("/api/v1")
	{
		// No /profile route needed

		incomeRoutes := apiV1.Group("/income")
		{
			incomeRoutes.POST("", incomeHandler.CreateIncomeHandler)
			incomeRoutes.GET("/:id", incomeHandler.GetIncomeHandler)
			incomeRoutes.GET("", incomeHandler.ListIncomesHandler)
			incomeRoutes.PUT("/:id", incomeHandler.UpdateIncomeHandler)
			incomeRoutes.DELETE("/:id", incomeHandler.DeleteIncomeHandler)
		}

		expenseRoutes := apiV1.Group("/expenses") // Changed from apiProtected to apiV1
		{
			expenseRoutes.POST("", expenseHandler.CreateExpenseHandler)
			expenseRoutes.GET("/:id", expenseHandler.GetExpenseHandler)
			expenseRoutes.GET("", expenseHandler.ListExpensesHandler)
			expenseRoutes.PUT("/:id", expenseHandler.UpdateExpenseHandler)
			expenseRoutes.DELETE("/:id", expenseHandler.DeleteExpenseHandler)
		}

		savingsRoutes := apiV1.Group("/savings") // Changed from apiProtected to apiV1
		{
			savingsRoutes.POST("", savingsHandler.CreateSavingsHandler)
			savingsRoutes.GET("/:id", savingsHandler.GetSavingsHandler)
			savingsRoutes.GET("", savingsHandler.ListSavingsHandler)
			savingsRoutes.PUT("/:id", savingsHandler.UpdateSavingsHandler)
			savingsRoutes.DELETE("/:id", savingsHandler.DeleteSavingsHandler)
		}

		debtRoutes := apiV1.Group("/debts") // Changed from apiProtected to apiV1
		{
			debtRoutes.POST("", debtHandler.CreateDebtHandler)
			debtRoutes.GET("/:id", debtHandler.GetDebtHandler)
			debtRoutes.GET("", debtHandler.ListDebtsHandler)
			debtRoutes.PUT("/:id", debtHandler.UpdateDebtHandler)
			debtRoutes.DELETE("/:id", debtHandler.DeleteDebtHandler)
		}

		summaryRoutes := apiV1.Group("/summary") // Changed from apiProtected to apiV1
		{
			summaryRoutes.GET("/monthly", summaryHandler.GetMonthlySummaryHandler)
			summaryRoutes.GET("/weekly", summaryHandler.GetWeeklySummaryHandler)
			summaryRoutes.GET("/yearly", summaryHandler.GetYearlySummaryHandler)
		}

		apiV1.GET("/advice", aiAdviceHandler.GetAdviceHandler) // Changed from apiProtected to apiV1

		reportRoutes := apiV1.Group("/reports") // Changed from apiProtected to apiV1
		{
			reportRoutes.GET("/csv", reportHandler.GenerateCSVReportHandler)
			reportRoutes.GET("/pdf", reportHandler.GeneratePDFReportHandler)
		}

		analyticsRoutes := apiV1.Group("/analytics")
		{
			analyticsRoutes.GET("/expense-categories", analyticsHandler.GetExpenseBreakdownHandler)
			analyticsRoutes.GET("/income-expense-trend", analyticsHandler.GetIncomeExpenseTrendHandler)
		}
	}

	// webAuthGroup related code was already removed effectively by making /dashboard direct.
	// No further changes needed for webAuthGroup specifically.

	// Initialize and start Cron scheduler
	cronScheduler := cron.New(cron.WithLocation(time.UTC), cron.WithSeconds()) // Use UTC, enable seconds

	// Schedule CheckDueDatesAndGoals to run daily at 3 AM UTC
	_, errCron := cronScheduler.AddFunc("0 0 3 * * *", func() {
		// For testing, run every minute: "* * * * *"
		// _, errCron := cronScheduler.AddFunc("* * * * *", func() {
		log.Println("Cron Job: Running scheduled task CheckDueDatesAndGoals")
		notificationService.CheckDueDatesAndGoals()
	})
	if errCron != nil {
		log.Fatalf("Error adding cron job CheckDueDatesAndGoals: %v", errCron)
	}

	cronScheduler.Start()
	log.Println("Cron scheduler started. Daily checks scheduled for 3:00 AM UTC.")
	// In a real application, consider graceful shutdown of the scheduler:
	// defer cronScheduler.Stop() // This needs careful handling with server lifecycle

	// --- Frontend Static File Serving (for production builds) ---
	// This section serves the built Vue.js application.
	// Assumes your Vue app is built into a 'frontend/dist' directory
	// relative to your Go executable's location.
	staticFilesPath := "./frontend/dist" // Adjust if your build output is elsewhere

	// Serve static assets (JS, CSS, images, etc.) from the 'assets' subdirectory
	// Vite typically places these in an 'assets' folder within 'dist'.
	router.StaticFS("/assets", http.Dir(filepath.Join(staticFilesPath, "assets")))

	// Serve other static files like favicon.ico from the root of the dist folder
	router.StaticFile("/favicon.ico", filepath.Join(staticFilesPath, "favicon.ico"))
	// Add other specific static files if needed (e.g., manifest.json, robots.txt)

	// For Single Page Application (SPA) routing:
	// All non-API, non-static-file routes should serve your main index.html.
	// This allows the Vue router to handle client-side navigation.
	router.NoRoute(func(c *gin.Context) {
		// Check if it's an API call, if so, let Gin's default 404 handle it or return custom API 404
		if len(c.Request.URL.Path) > 4 && c.Request.URL.Path[:5] == "/api/" {
			c.JSON(http.StatusNotFound, gin.H{"code": "API_ENDPOINT_NOT_FOUND", "message": "The requested API endpoint does not exist."})
			return
		}
		c.File(filepath.Join(staticFilesPath, "index.html"))
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
