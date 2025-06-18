package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (c *UploadController) HandleDeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract file key from URL path or query parameter
	fileKey := r.URL.Query().Get("key")
	if fileKey == "" {
		http.Error(w, "Missing file key parameter", http.StatusBadRequest)
		return
	}

	// Delete from S3
	if err := c.awsService.DeleteFile(fileKey); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete file: %v", err), http.StatusInternalServerError)
		return
	}

	// Delete from database
	if err := c.fileRepo.DeleteByKey(fileKey); err != nil {
		log.Printf("Warning: File deleted from S3 but not from DB: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  true,
		"message": "File deleted successfully",
		"key":     fileKey,
	})
}
