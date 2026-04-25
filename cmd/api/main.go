package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pocketful/internal/config"
	"pocketful/internal/db"
	"pocketful/internal/middleware"
	"pocketful/internal/repository"
	"pocketful/internal/routes"
	"pocketful/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// ── Load Configuration ────────────────────────────────────────────────────
	config.Load()
	log.Println("✅ Configuration loaded")

	// ── Connect to MongoDB ─────────────────────────────────────────────────────
	db.Connect(config.AppConfig.MongoURI, config.AppConfig.DBName)
	defer db.Disconnect()

	// ── Create MongoDB Indexes ─────────────────────────────────────────────────
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userRepo := repository.NewUserRepository()
	if err := userRepo.CreateIndex(ctx); err != nil {
		log.Printf("Warning: Could not create user email index: %v", err)
	}

	// ── Seed Default Admin User ───────────────────────────────────────────────
	authService := service.NewAuthService(userRepo)
	if err := authService.SeedAdminUser(ctx, config.AppConfig.AdminEmail, config.AppConfig.AdminPassword); err != nil {
		log.Printf("Warning: Admin seed failed: %v", err)
	} else {
		log.Printf("✅ Admin user ready: %s", config.AppConfig.AdminEmail)
	}

	// ── Setup Gin Router ──────────────────────────────────────────────────────
	gin.SetMode(config.AppConfig.GinMode)

	router := gin.New()
	router.Use(gin.Recovery())        // Recover from panics
	router.Use(middleware.RequestLogger()) // Custom request logger

	// Register all routes
	routes.Setup(router)

	// ── Start HTTP Server with Graceful Shutdown ───────────────────────────────
	srv := &http.Server{
		Addr:         ":" + config.AppConfig.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in a goroutine so it doesn't block the shutdown logic below
	go func() {
		log.Printf("🚀 Pocketful KYC API running on http://localhost:%s", config.AppConfig.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// ── Graceful Shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}
