package usecases

import (
	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

func isFarmer(principal auth.Principal) bool {
	return principal.Role == domain.RoleAgricultor
}

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

func isMayorista(principal auth.Principal) bool {
	switch principal.Role {
	case domain.RoleCompradorMayoristaDetallista, domain.RoleCompradorMayoristaCorporativo:
		return true
	default:
		return false
	}
}
