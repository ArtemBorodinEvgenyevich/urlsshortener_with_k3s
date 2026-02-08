package response

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func JSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			return
		}
	}
}

func Error(w http.ResponseWriter, logger *zap.Logger, code int, err error, msg string) {
	if logger != nil {
		logger.Error("HTTP Error",
			zap.Int("status_code", code),
			zap.Error(err),
			zap.String("message", msg),
		)
	}

	resp := ErrorResponse{
		Error:   http.StatusText(code),
		Message: msg,
	}

	JSON(w, code, resp)
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
