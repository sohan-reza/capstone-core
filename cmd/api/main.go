package main

import (
	"log"
	"net/http"

	"github.com/sohan-reza/capstone-core/internal/config"
	"github.com/sohan-reza/capstone-core/internal/controller"

	"github.com/go-chi/chi/v5"
)

func main() {

	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting server on port %s", cfg.Server.Port)

	r := chi.NewRouter()

	s := &http.Server{
		Addr:           ":" + cfg.Server.Port,
		Handler:        r,
		ReadTimeout:    cfg.Server.TimeoutRead,
		WriteTimeout:   cfg.Server.TimeoutWrite,
		MaxHeaderBytes: 1 << 20,
	}

	uploadController := controller.NewUploadController(cfg)
	r.Post("/upload", uploadController.HandleFileUpload)

	s.ListenAndServe()
}
