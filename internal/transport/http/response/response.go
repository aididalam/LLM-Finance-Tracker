package response

import (
	"encoding/json"
	"net/http"
)

func ResSuccess(w http.ResponseWriter, data any, status ...int) {
	code := http.StatusOK
	if len(status) > 0 {
		code = status[0]
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{"data": data})
}

func ResError(w http.ResponseWriter, err any, status ...int) {
	code := http.StatusBadRequest
	if len(status) > 0 {
		code = status[0]
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{"error": err})
}
