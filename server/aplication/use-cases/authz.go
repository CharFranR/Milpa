package usecases

import (
	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

// isFarmer reports whether the caller publishes as an agricultor: supply
// offers, liquidations, offerings and supplier inventory are all supply-side
// publications.
func isFarmer(principal auth.Principal) bool {
	return principal.Role == domain.RoleAgricultor
}

// isBuyer reports whether the caller is any of the three comprador roles.
func isBuyer(principal auth.Principal) bool {
	switch principal.Role {
	case domain.RoleCompradorMinorista,
		domain.RoleCompradorMayoristaDetallista,
		domain.RoleCompradorMayoristaCorporativo:
		return true
	default:
		return false
	}
}

// isMayorista reports whether the caller is one of the two mayorista roles,
// the only ones that publish supply requests and the only ones a restricted
// liquidation is offered to.
func isMayorista(principal auth.Principal) bool {
	switch principal.Role {
	case domain.RoleCompradorMayoristaDetallista, domain.RoleCompradorMayoristaCorporativo:
		return true
	default:
		return false
	}
}
