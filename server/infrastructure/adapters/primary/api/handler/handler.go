package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	domain "milpa/domain/entities"
)

type envelope struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{Data: data})
}

func respondError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{Error: msg})
}

func statusCode(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrDuplicate), errors.Is(err, domain.ErrEmailTaken):
		return http.StatusConflict
	case isValidationError(err):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func isValidationError(err error) bool {
	return errors.Is(err, domain.ErrInvalidInput) ||
		errors.Is(err, domain.ErrInvalidPrice) ||
		errors.Is(err, domain.ErrInvalidRating) ||
		errors.Is(err, domain.ErrInvalidOfferingType) ||
		errors.Is(err, domain.ErrNameRequired) ||
		errors.Is(err, domain.ErrMessageRequired) ||
		errors.Is(err, domain.ErrEmailRequired) ||
		errors.Is(err, domain.ErrFirstNameRequired) ||
		errors.Is(err, domain.ErrLastNameRequired) ||
		errors.Is(err, domain.ErrPasswordRequired) ||
		errors.Is(err, domain.ErrDepartmentRequired) ||
		errors.Is(err, domain.ErrMunicipalityRequired) ||
		errors.Is(err, domain.ErrAddressLineRequired) ||
		errors.Is(err, domain.ErrOwnerRequired)
}

func handleError(w http.ResponseWriter, err error) {
	code := statusCode(err)
	if code == http.StatusInternalServerError {
		respondError(w, code, "internal server error")
		return
	}
	respondError(w, code, err.Error())
}
