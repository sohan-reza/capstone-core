package controller

import (
	"encoding/json"
	"net/http"
)

func (c *UploadController) GetFilesByTeamID(w http.ResponseWriter, r *http.Request) {
	teamID := r.URL.Query().Get("team_id")
	if teamID == "" {
		http.Error(w, "team_id parameter is required", http.StatusBadRequest)
		return
	}

	files, err := c.fileRepo.GetFilesByTeamID(teamID)
	if err != nil {
		http.Error(w, "failed to fetch files", http.StatusInternalServerError)
		return
	}

	if len(files) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}
