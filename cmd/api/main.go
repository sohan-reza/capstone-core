package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sohan-reza/capstone-core/internal/config"
	"github.com/sohan-reza/capstone-core/internal/controller"
	"github.com/sohan-reza/capstone-core/internal/model"
	"github.com/sohan-reza/capstone-core/internal/repository"
	"github.com/sohan-reza/capstone-core/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/go-chi/chi/v5"
)

func main() {

	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	awsService, err := service.NewAWSService(cfg.AWS.BucketName, cfg.AWS.Region, cfg.AWS.AccessKeyID, cfg.AWS.SecretAccessKey)
	if err != nil {
		log.Fatalf("Failed to initialize AWS service: %v", err)
	}

	// Initialize database and repository
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.Name, cfg.Database.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate (for development)
	if err := db.AutoMigrate(&model.File{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fileRepo := repository.NewFileRepository(db)

	r := chi.NewRouter()

	s := &http.Server{
		Addr:           ":" + cfg.Server.Port,
		Handler:        r,
		ReadTimeout:    cfg.Server.TimeoutRead,
		WriteTimeout:   cfg.Server.TimeoutWrite,
		MaxHeaderBytes: 1 << 20,
	}

	uploadController := controller.NewUploadController(cfg, awsService, fileRepo)
	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Post("/upload", uploadController.HandleFileUpload)
		v1.Delete("/files", uploadController.HandleDeleteFile)
		v1.Get("/bucket/backup", uploadController.HandleDownloadBucket)
		v1.Get("/download", uploadController.GetFilesByTeamID)
	})

	s.ListenAndServe()
}
