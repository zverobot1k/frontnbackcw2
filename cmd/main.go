package main

import (
	"log"
	"net/http"
	"os"

	"somewebproject/internal/config"
	"somewebproject/internal/middleware"
	"somewebproject/internal/migration"
	"somewebproject/internal/repository"
	"somewebproject/internal/router"
	"somewebproject/internal/service"
	"somewebproject/internal/transport"
)

func main() {
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("init db: %v", err)
	}

	if err := migration.Run(db); err != nil {
		log.Fatalf("run migration: %v", err)
	}

	userRepo := repository.NewUserRepo(db)
	productRepo := repository.NewProductRepo(db)

	jwtSecret := getEnv("JWT_SECRET", "dev_secret")
	authService := service.NewAuthService(userRepo, jwtSecret)
	userService := service.NewUserService(userRepo)
	productService := service.NewProductService(productRepo)

	authHandler := transport.NewHandler(authService, userService)
	productHandler := transport.NewProductHandler(productService)
	authMiddleware := middleware.NewAuthMiddleware(jwtSecret, userRepo)
	requireRoles := middleware.RequireRoles

	handler := router.NewRouter(authHandler, productHandler, authMiddleware, requireRoles)

	port := getEnv("APP_PORT", "8080")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	log.Printf("server started on :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen and serve: %v", err)
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
