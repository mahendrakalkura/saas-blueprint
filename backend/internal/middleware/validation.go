package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type ErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

func ValidateJSON(next http.HandlerFunc, v interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(v); err != nil {
			respondJSON(w, http.StatusBadRequest, ErrorResponse{
				Error: "Invalid JSON payload",
			})
			return
		}

		if err := validate.Struct(v); err != nil {
			fields := make(map[string]string)
			for _, err := range err.(validator.ValidationErrors) {
				fields[err.Field()] = getValidationMessage(err)
			}

			respondJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
				Error:  "Validation failed",
				Fields: fields,
			})
			return
		}

		next.ServeHTTP(w, r)
	}
}

func getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "uuid":
		return "Invalid UUID format"
	default:
		return "Invalid value"
	}
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
