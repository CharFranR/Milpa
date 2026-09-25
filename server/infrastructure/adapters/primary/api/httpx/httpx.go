package httpx

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

type envelope struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func Respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{Data: data})
}

func RespondError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(envelope{Error: msg})
}

func StatusCode(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrUnauthorized), errors.Is(err, auth.ErrUnauthenticated):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrDuplicate), errors.Is(err, domain.ErrEmailTaken):
		return http.StatusConflict
	case IsValidationError(err):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func HandleError(w http.ResponseWriter, err error) {
	code := StatusCode(err)
	if code == http.StatusInternalServerError {
		log.Printf("ERROR: %v", err)
		RespondError(w, code, "internal server error")
		return
	}
	RespondError(w, code, err.Error())
}

func IsValidationError(err error) bool {
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
		errors.Is(err, domain.ErrOwnerRequired) ||
		errors.Is(err, domain.ErrReporterRequired) ||
		errors.Is(err, domain.ErrInvalidReportTargetType) ||
		errors.Is(err, domain.ErrInvalidReportStatus) ||
		errors.Is(err, domain.ErrTargetRequired) ||
		errors.Is(err, domain.ErrReasonRequired) ||
		errors.Is(err, domain.ErrSelfReport) ||
		errors.Is(err, domain.ErrReportAlreadyPending) ||
		errors.Is(err, domain.ErrReportAlreadyResolved) ||
		errors.Is(err, domain.ErrCannotSuspendSelf) ||
		errors.Is(err, domain.ErrCannotSuspendAdmin)
}
