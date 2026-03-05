package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func ParseID(id string) openapi_types.UUID {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.NewSHA1(uuid.NameSpaceURL, []byte(id))
	}
	return parsed
}

func WriteError(w http.ResponseWriter, status int, code ErrorCode, message, requestID string) {
	WriteJSON(w, status, ErrorResponse{
		Error: struct {
			Code      ErrorCode               `json:"code"`
			Details   *map[string]interface{} `json:"details,omitempty"`
			Message   string                  `json:"message"`
			RequestId string                  `json:"request_id"`
		}{
			Code:      code,
			Message:   message,
			RequestId: requestID,
		},
	})
}
