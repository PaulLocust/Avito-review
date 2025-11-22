// internal/controller/http/v1/response.go
package v1

import (
	"encoding/json"
	"net/http"

	"github.com/PaulLocust/Avito-review/internal/entity"
	"github.com/PaulLocust/Avito-review/internal/dto"
)

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, code entity.ErrorCode, message string) {
	errorResponse := dto.ErrorResponse{
		Error: struct {
			Code    dto.ErrorResponseErrorCode `json:"code"`
			Message string                     `json:"message"`
		}{
			Code:    dto.ErrorResponseErrorCode(code),
			Message: message,
		},
	}
	
	writeJSONResponse(w, statusCode, errorResponse)
}