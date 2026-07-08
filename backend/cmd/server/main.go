package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/electrum/dynamic-pricing-engine/internal/application"
	"github.com/electrum/dynamic-pricing-engine/internal/domain/pricing"
	"github.com/electrum/dynamic-pricing-engine/internal/infrastructure/auth"
	infraconfig "github.com/electrum/dynamic-pricing-engine/internal/infrastructure/config"
	"github.com/electrum/dynamic-pricing-engine/internal/infrastructure/postgres"
	presentation "github.com/electrum/dynamic-pricing-engine/internal/presentation/http"
)

func main() {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/pricing_config.json"
	}

	// -----------------------------------------------------------------------
	// Database
	// -----------------------------------------------------------------------
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		pool, err := postgres.InitDB(ctx)
		if err != nil {
			log.Printf("⚠ Database init failed: %v — running without DB", err)
		} else {
			defer postgres.Close()
			_ = pool
			log.Println("✓ Database connected")
		}
	} else {
		log.Println("⚠ DATABASE_URL not set — running without database")
	}

	// -----------------------------------------------------------------------
	// Infrastructure adapters (repositories)
	// -----------------------------------------------------------------------
	vehicleRepo := postgres.NewVehicleRepo()
	zoneRepo := postgres.NewZoneRepo()
	configRepo := postgres.NewConfigRepo()
	auditRepo := postgres.NewAuditRepo()

	// JSON file config store (hot-reload via watcher)
	jsonStore := infraconfig.NewJSONStore(configPath)
	if err := jsonStore.Load(ctx); err != nil {
		log.Printf("⚠ Config file load failed: %v — using defaults", err)
	}
	cfg, _ := jsonStore.GetActive(ctx)
	if cfg != nil {
		log.Printf("✓ Config loaded (version %d)", cfg.Version)
	}

	// Start config file watcher (20s interval)
	watcherCtx, watcherCancel := context.WithCancel(context.Background())
	defer watcherCancel()
	go jsonStore.Watcher(watcherCtx, 20*time.Second, func(cfg *pricing.PricingConfig) {
		log.Printf("⚡ Config hot-reloaded (version %d)", cfg.Version)
	})

	// -----------------------------------------------------------------------
	// Auth
	// -----------------------------------------------------------------------
	userRepo := auth.NewUserRepo()
	jwtSvc := auth.NewJWTService(userRepo)

	// -----------------------------------------------------------------------
	// Application layer (usecase)
	// -----------------------------------------------------------------------
	pricingUseCase := application.NewPricingUseCase(
		vehicleRepo, zoneRepo, configRepo, auditRepo,
	)

	// -----------------------------------------------------------------------
	// Presentation layer (HTTP)
	// -----------------------------------------------------------------------
	handler := presentation.NewHandler(pricingUseCase, jwtSvc)

	// -----------------------------------------------------------------------
	// Router
	// -----------------------------------------------------------------------
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "electrum-dynamic-pricing-engine",
			"version":   "2.0.0",
			"arch":      "clean-architecture",
		})
	})

	v1 := router.Group("/api/v1")
	{
		// Public
		v1.POST("/auth/login", handler.Login)

		// Protected: pricing (requires JWT)
		pricingGroup := v1.Group("/pricing")
		pricingGroup.Use(presentation.AuthMiddleware(jwtSvc))
		{
			pricingGroup.GET("", handler.GetPricing)
			pricingGroup.GET("/breakdown", handler.GetBreakdown)
		}

		// Config (read = JWT only, write = admin)
		configGroup := v1.Group("/config")
		configGroup.Use(presentation.AuthMiddleware(jwtSvc))
		{
			configGroup.GET("", handler.GetConfig)
		}

		// Protected: admin (requires JWT + admin role)
		adminGroup := v1.Group("/admin")
		adminGroup.Use(presentation.AuthMiddleware(jwtSvc), presentation.AdminMiddleware())
		{
			adminGroup.PUT("/config", handler.UpdateConfig)
			adminGroup.GET("/config/history", handler.GetConfigHistory)
			adminGroup.GET("/pricing/audit", handler.GetAuditLogs)
		}

		// Zones (public read-only)
		zonesGroup := v1.Group("/zones")
		{
			zonesGroup.GET("", handler.GetZones)
		}

		// Vehicles (public read-only)
		vehiclesGroup := v1.Group("/vehicles")
		{
			vehiclesGroup.GET("", handler.GetVehicles)
		}
	}

	// -----------------------------------------------------------------------
	// Start server
	// -----------------------------------------------------------------------
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("✓ Dynamic Pricing Engine (clean-arch) running on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
