package routes

import (
	"pocketful/internal/handler"
	"pocketful/internal/middleware"
	"pocketful/internal/repository"
	"pocketful/internal/service"

	"github.com/gin-gonic/gin"
)

// Setup registers all application routes on the given Gin engine.
func Setup(router *gin.Engine) {
	// Initialize repositories
	userRepo := repository.NewUserRepository()
	kycRepo := repository.NewKYCRepository()
	docRepo := repository.NewDocumentRepository()

	// Initialize services
	authService := service.NewAuthService(userRepo)
	kycService := service.NewKYCService(kycRepo, docRepo)

	// Initialize handlers
	userHandler := handler.NewUserHandler(authService)
	kycHandler := handler.NewKYCHandler(kycService)
	adminHandler := handler.NewAdminHandler(kycService)

	// ── Health Check ──────────────────────────────────────────────────────────
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "pocketful-kyc"})
	})

	// ── Public Routes (No Auth Required) ─────────────────────────────────────
	public := router.Group("/")
	{
		public.POST("/register", userHandler.Register)
		public.POST("/login", userHandler.Login)
	}

	// ── Authenticated User Routes ─────────────────────────────────────────────
	auth := router.Group("/")
	auth.Use(middleware.AuthRequired())
	{
		kyc := auth.Group("/kyc")
		{
			kyc.POST("/initiate", kycHandler.Initiate)
			kyc.POST("/upload", kycHandler.Upload)
			kyc.GET("/status", kycHandler.Status)
		}
	}

	// ── Admin Routes ──────────────────────────────────────────────────────────
	admin := router.Group("/admin")
	admin.Use(middleware.AuthRequired(), middleware.AdminRequired())
	{
		admin.POST("/verify", adminHandler.Verify)
	}
}
