package controller

import (
	"encoding/json"
	"net/http"
)

type FileResponse struct {
	Status   string                 `json:"status"`
	Message  string                 `json:"message,omitempty"`
	FileName string                 `json:"file_filename"`
	FileSize int64                  `json:"file_size"`
	FileType string                 `json:"file_type"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type ErrorResponse struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	switch v := payload.(type) {
	case FileResponse:
		v.Status = "success"
		json.NewEncoder(w).Encode(v)
	case map[string]interface{}:
		if _, exists := v["status"]; !exists {
			v["status"] = "success"
		}
		json.NewEncoder(w).Encode(v)
	default:
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"payload": payload,
		})
	}
}

// func respondWithError(w http.ResponseWriter, statusCode int, message string, details map[string]interface{}) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(statusCode)

// 	response := ErrorResponse{
// 		Status:  "error",
// 		Message: message,
// 	}

// 	if details != nil {
// 		response.Details = details
// 	}

// 	json.NewEncoder(w).Encode(response)
// }

func respondWithError(w http.ResponseWriter, statusCode int, message string, errDetails interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Status:  "error",
		Message: message,
	}

	switch v := errDetails.(type) {
	case error:
		response.Error = v.Error()
	case map[string]interface{}:
		response.Details = v
	case nil:
		// No additional error details
	default:
		response.Details = map[string]interface{}{
			"error_details": v,
		}
	}

	json.NewEncoder(w).Encode(response)
}
