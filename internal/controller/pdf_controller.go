package controller

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-resty/resty/v2"
)

type PDFController struct {
	uploadDir     string
	plagiarismAPI string
	client        *resty.Client
	threshold     int
}

func NewPDFController(uploadDir, apiEndpoint string, threshold int) *PDFController {
	return &PDFController{
		uploadDir:     uploadDir,
		plagiarismAPI: apiEndpoint,
		client:        resty.New(),
		threshold:     threshold,
	}
}

func (c *PDFController) HandleUpload(w http.ResponseWriter, r *http.Request, oldFilename string, newFilename string, filePath string) {
	file, header, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error retrieving file", err)
		return
	}
	defer file.Close()

	isValid, plagiarismPercent, err := c.checkPlagiarism(filePath)
	if err != nil {
		os.Remove(filePath)
		respondWithError(w, http.StatusServiceUnavailable, "Plagiarism service unavailable", err)
		return
	}

	if !isValid {
		os.Remove(filePath)
		respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status": "error",
			"error":  "Plagiarism check failed",
			"details": map[string]interface{}{
				"message":   "n% or less plagiarism is accepted",
				"type":      "plagiarism",
				"detected":  plagiarismPercent,
				"threshold": c.threshold,
			},
		})
		return
	}

	respondWithJSON(w, http.StatusOK, FileResponse{
		Status:   "success",
		Message:  "PDF processed successfully",
		OldName:  oldFilename,
		NewName:  newFilename,
		FileSize: header.Size,
		FileType: filepath.Ext(oldFilename),
		Metadata: map[string]interface{}{
			"plagiarism_checked": true,
			"plagiarism_percent": plagiarismPercent,
		},
	})
}

func (c *PDFController) checkPlagiarism(filePath string) (bool, float64, error) {
	resp, err := c.client.R().
		SetFile("file", filePath).
		SetResult(&struct {
			Percentage float64 `json:"matchPercent"`
		}{}).
		Post(c.plagiarismAPI + "/check/plagiarism")

	if err != nil {
		return false, 0, err
	}

	if resp.StatusCode() != http.StatusOK {
		return false, 0, nil
	}

	result := resp.Result().(*struct {
		Percentage float64 `json:"matchPercent"`
	})

	isAcceptable := result.Percentage < float64(c.threshold)

	return isAcceptable, result.Percentage, nil
}
