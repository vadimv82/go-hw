package main

import (
	"cruder/internal/controller"
	"cruder/internal/handler"
	"cruder/internal/middleware"
	"cruder/internal/repository"
	"cruder/internal/service"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
	}

	dbConn, err := repository.NewPostgresConnection(dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	repositories := repository.NewRepository(dbConn.DB())
	services := service.NewService(repositories)
	controllers := controller.NewController(services)
	r := gin.Default()

	r.Use(middleware.JSONLoggingMiddleware())

	r.Use(middleware.APIKeyAuthMiddleware())

	handler.New(r, controllers.Users)
	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
