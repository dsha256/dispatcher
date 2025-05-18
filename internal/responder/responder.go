package responder

import (
	"encoding/json"
	"net/http"

	"github.com/dsha256/dispatcher/internal/types"
)

func WriteJSON(w http.ResponseWriter, status int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func WriteSuccess[T any](w http.ResponseWriter, status int, message string, data T) {
	WriteJSON(w, status, types.NewSuccessResponse(message, data))
}

func WriteError(w http.ResponseWriter, status int, err error) {
	WriteJSON(w, status, types.NewErrorResponse[string](err.Error()))
}
