package controller

import (
	"fmt"
	"net/http"
)

func (c *UploadController) HandleDownloadBucket(w http.ResponseWriter, r *http.Request) {
	// Set headers for zip download
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"bucket-backup.zip\"")

	// Stream zip directly to response
	if err := c.awsService.DownloadBucketAsZip(w); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create zip: %v", err), http.StatusInternalServerError)
		return
	}
}
