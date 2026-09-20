package domain

import "errors"

var (
	ErrNotFound             = errors.New("resource not found")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrForbidden            = errors.New("forbidden")
	ErrInvalidInput         = errors.New("invalid input")
	ErrDuplicate            = errors.New("resource already exists")
	ErrEmailTaken           = errors.New("email already registered")
	ErrInvalidPrice         = errors.New("price must be greater than zero")
	ErrInvalidRating        = errors.New("rating must be between 1 and 5")
	ErrNameRequired         = errors.New("name is required")
	ErrInvalidOfferingType  = errors.New("offering type must be product or service")
	ErrMessageRequired      = errors.New("message is required")
	ErrEmailRequired        = errors.New("email is required")
	ErrFirstNameRequired    = errors.New("first name is required")
	ErrLastNameRequired     = errors.New("last name is required")
	ErrPasswordRequired     = errors.New("password is required")
	ErrDepartmentRequired   = errors.New("department is required")
	ErrMunicipalityRequired = errors.New("municipality is required")
	ErrAddressLineRequired  = errors.New("address line is required")
	ErrOwnerRequired        = errors.New("company must have an owner")
	ErrValidRoleRequired    = errors.New("valid user role is required ")
	ErrPhoneNumberRequired  = errors.New("A phone number is required")
	ErrUserNotFound         = errors.New("user not found")

	// Liquidation errors
	ErrProductNameRequired      = errors.New("product name is required")
	ErrInvalidQuantity          = errors.New("quantity must be greater than zero")
	ErrUnitOfMeasureRequired    = errors.New("unit of measure is required")
	ErrLiquidationNotOpen       = errors.New("liquidation is not open")
	ErrLiquidationCannotAssign  = errors.New("liquidation cannot be assigned in current status")
	ErrInvalidVisibility        = errors.New("visibility must be public or private")
)
