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
	if status == http.StatusNoContent {
		w.WriteHeader(status)
		return
	}
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
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, domain.ErrUserNotFound), errors.Is(err, domain.ErrInventoryNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrUnauthorized), errors.Is(err, auth.ErrUnauthenticated):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrForbidden), errors.Is(err, domain.ErrUserSuspended):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrDuplicate), errors.Is(err, domain.ErrEmailTaken),
		errors.Is(err, domain.ErrLiquidationNotOpen), errors.Is(err, domain.ErrLiquidationCannotAssign),
		errors.Is(err, domain.ErrTransactionNotCompleted), errors.Is(err, domain.ErrReviewAlreadyExists):
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
		errors.Is(err, domain.ErrInvalidQuantity) ||
		errors.Is(err, domain.ErrInvalidOfferingType) ||
		errors.Is(err, domain.ErrNameRequired) ||
		errors.Is(err, domain.ErrCodeRequired) ||
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
		errors.Is(err, domain.ErrCannotSuspendAdmin) ||
		errors.Is(err, domain.ErrValidRoleRequired) ||
		errors.Is(err, domain.ErrPhoneNumberRequired) ||
		errors.Is(err, domain.ErrVarietyRequired) ||
		errors.Is(err, domain.ErrCategoryRequired) ||
		errors.Is(err, domain.ErrProductNameRequired) ||
		errors.Is(err, domain.ErrUnitOfMeasureRequired) ||
		errors.Is(err, domain.ErrInvalidVisibility) ||
		errors.Is(err, domain.ErrAuthorRequired) ||
		errors.Is(err, domain.ErrSelfReview) ||
		errors.Is(err, domain.ErrReviewTargetMismatch) ||
		errors.Is(err, domain.ErrInvalidReviewTargetType) ||
		errors.Is(err, domain.ErrTransactionRequired) ||
		errors.Is(err, domain.ErrReviewTargetPartyMismatch) ||
		errors.Is(err, domain.ErrConversationRequired) ||
		errors.Is(err, domain.ErrSenderMessageRequired) ||
		errors.Is(err, domain.ErrContentMessageRequired) ||
		errors.Is(err, domain.ErrFarmerConversationRequired) ||
		errors.Is(err, domain.ErrBuyerConversationRequired) ||
		errors.Is(err, domain.ErrOfferingconversationRequied) ||
		errors.Is(err, domain.ErrConversationTargetRequired)
}
