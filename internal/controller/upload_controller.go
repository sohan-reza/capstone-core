package controller

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sohan-reza/capstone-core/internal/config"
	"github.com/sohan-reza/capstone-core/internal/utils"
)

type UploadController struct {
	pdfController     *PDFController
	archiveController *ArchiveController
	uploadDir         string
}

func NewUploadController(cfg *config.Config) *UploadController {

	os.MkdirAll(cfg.Upload.Dir, 0755)

	return &UploadController{
		pdfController: NewPDFController(
			cfg.Upload.Dir,
			cfg.Plagiarism.APIEndpoint,
			cfg.Plagiarism.Threshold,
		),
		archiveController: &ArchiveController{},
		uploadDir:         cfg.Upload.Dir,
	}
}

func (c *UploadController) HandleFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "File too large or invalid form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	newFilename := generateUniqueFilename(header.Filename)
	filePath := filepath.Join(c.uploadDir, newFilename)

	if err := saveUploadedFile(file, filePath); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	switch utils.DetectFileType(header) {
	case utils.PDF:
		c.pdfController.HandleUpload(w, r, header.Filename, newFilename, filePath)
	case utils.Archive:
		c.archiveController.HandleUpload(w, r, header.Filename, newFilename, filePath)
	default:
		os.Remove(filePath)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnsupportedMediaType)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  false,
			"message": "Unsupported file type",
		})

	}
}

func generateUniqueFilename(original string) string {
	ext := filepath.Ext(original)
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes) + ext
}

func saveUploadedFile(file multipart.File, dstPath string) error {
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	return err
}
