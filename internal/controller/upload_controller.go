// package controller

// import (
// 	"crypto/rand"
// 	"encoding/hex"
// 	"encoding/json"
// 	"io"
// 	"mime/multipart"
// 	"net/http"
// 	"os"
// 	"path/filepath"

// 	"github.com/sohan-reza/capstone-core/internal/config"
// 	"github.com/sohan-reza/capstone-core/internal/utils"
// )

// type UploadController struct {
// 	pdfController     *PDFController
// 	archiveController *ArchiveController
// 	uploadDir         string
// }

// func NewUploadController(cfg *config.Config) *UploadController {

// 	os.MkdirAll(cfg.Upload.Dir, 0755)

// 	return &UploadController{
// 		pdfController: NewPDFController(
// 			cfg.Upload.Dir,
// 			cfg.Plagiarism.APIEndpoint,
// 			cfg.Plagiarism.Threshold,
// 		),
// 		archiveController: &ArchiveController{},
// 		uploadDir:         cfg.Upload.Dir,
// 	}
// }

// func (c *UploadController) HandleFileUpload(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	err := r.ParseMultipartForm(100 << 20)
// 	if err != nil {
// 		http.Error(w, "File too large or invalid form", http.StatusBadRequest)
// 		return
// 	}

// 	file, header, err := r.FormFile("file")
// 	if err != nil {
// 		http.Error(w, "Error retrieving file", http.StatusBadRequest)
// 		return
// 	}
// 	defer file.Close()

// 	newFilename := generateUniqueFilename(header.Filename)
// 	filePath := filepath.Join(c.uploadDir, newFilename)

// 	if err := saveUploadedFile(file, filePath); err != nil {
// 		http.Error(w, "Failed to save file", http.StatusInternalServerError)
// 		return
// 	}

// 	switch utils.DetectFileType(header) {
// 	case utils.PDF:
// 		c.pdfController.HandleUpload(w, r, header.Filename, newFilename, filePath)
// 	case utils.Archive:
// 		c.archiveController.HandleUpload(w, r, header.Filename, newFilename, filePath)
// 	default:
// 		os.Remove(filePath)
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusUnsupportedMediaType)
// 		json.NewEncoder(w).Encode(map[string]interface{}{
// 			"status":  false,
// 			"message": "Unsupported file type",
// 		})

// 	}
// }

// func generateUniqueFilename(original string) string {
// 	ext := filepath.Ext(original)
// 	randomBytes := make([]byte, 8)
// 	rand.Read(randomBytes)
// 	return hex.EncodeToString(randomBytes) + ext
// }

// func saveUploadedFile(file multipart.File, dstPath string) error {
// 	dst, err := os.Create(dstPath)
// 	if err != nil {
// 		return err
// 	}
// 	defer dst.Close()

//		_, err = io.Copy(dst, file)
//		return err
//	}
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
	"strconv"
	"time"

	"github.com/sohan-reza/capstone-core/internal/config"
	"github.com/sohan-reza/capstone-core/internal/model"
	"github.com/sohan-reza/capstone-core/internal/repository"
	"github.com/sohan-reza/capstone-core/internal/service"
	"github.com/sohan-reza/capstone-core/internal/utils"
)

type UploadController struct {
	pdfController     *PDFController
	archiveController *ArchiveController
	uploadDir         string
	awsService        service.AWSService
	fileRepo          repository.FileRepository
}

func NewUploadController(cfg *config.Config, awsService service.AWSService, fileRepo repository.FileRepository) *UploadController {
	os.MkdirAll(cfg.Upload.Dir, 0755)

	return &UploadController{
		pdfController: NewPDFController(
			cfg.Upload.Dir,
			cfg.Plagiarism.APIEndpoint,
			cfg.Plagiarism.Threshold,
		),
		archiveController: &ArchiveController{},
		uploadDir:         cfg.Upload.Dir,
		awsService:        awsService,
		fileRepo:          fileRepo,
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

	defer os.Remove(filePath)

	// Upload to AWS S3
	key, originalName, err := c.awsService.UploadFile(filePath, header.Filename, strconv.Itoa(time.Now().Year()), r.FormValue("intake"), r.FormValue("team_id"))
	if err != nil {
		http.Error(w, "Failed to upload to cloud storage", http.StatusInternalServerError)
		return
	}

	// Generate presigned URL
	downloadURL, err := c.awsService.GeneratePresignedURL(key)
	if err != nil {
		http.Error(w, "Failed to generate download link", http.StatusInternalServerError)
		return
	}

	// Save file metadata to database
	fileRecord := &model.File{
		OriginalName: originalName,
		StorageKey:   key,
		DownloadURL:  downloadURL,
		Size:         header.Size,
		TeamID:       r.FormValue("team_id"),
		FileType:     filepath.Ext(header.Filename)[1:],
		ContentType:  header.Header.Get("Content-Type"),
	}

	if err := c.fileRepo.Create(fileRecord); err != nil {
		http.Error(w, "Failed to save file metadata", http.StatusInternalServerError)
		return
	}

	// Process the file based on its type
	switch utils.DetectFileType(header) {
	case utils.PDF:
		c.pdfController.HandleUpload(w, r, header.Filename, newFilename, filePath)
	case utils.Archive:
		c.archiveController.HandleUpload(w, r, header.Filename, newFilename, filePath)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnsupportedMediaType)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  false,
			"message": "Unsupported file type",
		})
		return
	}

	// Include the download URL in the response if needed
	// You can modify your PDF/Archive controller responses to include this
}

func generateUniqueFilename(original string) string {
	ext := filepath.Ext(original)
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes) + ext
}

func saveUploadedFile(file multipart.File, dstPath string) error {
	dst, err := os.Create(dstPath)
	println(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	return err
}
