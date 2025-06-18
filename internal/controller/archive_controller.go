package controller

import (
	"net/http"
	"path/filepath"
)

type ArchiveController struct {
	uploadDir string
}

func (c *ArchiveController) HandleUpload(w http.ResponseWriter, r *http.Request, oldFilename string, newFilename string, filePath string) {
	file, header, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error retrieving file", err)
		return
	}
	defer file.Close()

	respondWithJSON(w, http.StatusOK, FileResponse{
		Status:   "success",
		Message:  "Archive processed successfully",
		FileName: oldFilename,
		FileSize: header.Size,
		FileType: filepath.Ext(oldFilename),
	})
}
