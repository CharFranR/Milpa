package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/auth"
)

type OfferingUseCaseImpl struct {
	offeringRepo      port.OfferingRepository
	fuzzyRetrival     port.FuzzyRetrival
	userRepo          port.UserRepository
	timer             port.TimeProvider
	searchInvalidator port.Invalidator
}

func NewOfferingUseCase(offeringRepo port.OfferingRepository, userRepo port.UserRepository, timer port.TimeProvider, fuzzyRetrival port.FuzzyRetrival, searchInvalidator port.Invalidator) *OfferingUseCaseImpl {
	return &OfferingUseCaseImpl{
		offeringRepo:      offeringRepo,
		userRepo:          userRepo,
		timer:             timer,
		fuzzyRetrival:     fuzzyRetrival,
		searchInvalidator: searchInvalidator,
	}
}

func (uc *OfferingUseCaseImpl) CreateOffering(ctx context.Context, req dto.CreateOfferingRequest) (*dto.OfferingDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	if req.Type == domain.OfferingService {
		return nil, domain.ErrInvalidOfferingType
	}

	user, err := uc.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if user.ID != principal.UserID {
		return nil, domain.ErrForbidden
	}

	// RF-03: publishing requires a complete farmer profile. The check reuses
	// Address.IsComplete rather than restating the predicate, so the rule lives
	// in the domain and this call site cannot drift away from it. The user is
	// already loaded above, so this costs no extra query.
	if !user.Address.IsComplete() {
		return nil, fmt.Errorf("%w: a complete address (department, municipality and address line) is required to publish an offering", domain.ErrInvalidInput)
	}

	now := uc.timer.Now()

	offering, err := domain.NewOffering(req.UserID, req.Name, req.Type, now)
	if err != nil {
		return nil, err
	}

	offering.Description = req.Description
	if req.Price > 0 {
		offering.Price = req.Price
	}
	offering.ImageURL = req.ImageURL

	if err := offering.SetLocation(req.Latitude, req.Longitude, now); err != nil {
		return nil, err
	}
	offering.SetVariety(req.Variety, now)
	offering.SetUnitOfMeasure(req.UnitOfMeasureID, now)
	offering.SetQuantity(req.QuantityAvailable, now)
	offering.SetExpiry(req.ExpiresAt, now)
	offering.SetCategory(req.CategoryID, now)
	offering.SetCompany(req.CompanyID, now)

	if err := offering.RequirePublishable(); err != nil {
		return nil, err
	}

	if err := uc.offeringRepo.Save(ctx, offering); err != nil {
		return nil, err
	}

	// Build enriched index request for Elasticsearch
	indexReq := indexRequestFor(offering, user)

	// Save in elasticsearch
	err = uc.fuzzyRetrival.Index(ctx, indexReq)

	if err != nil {
		return nil, fmt.Errorf("CreateOffering.elasticsearch err: %w", err)
	}

	_ = uc.searchInvalidator.InvalidateAll(ctx)

	return offeringToDTO(offering), nil
}

func (uc *OfferingUseCaseImpl) GetByID(ctx context.Context, id uuid.UUID) (*dto.OfferingDTO, error) {
	offering, err := uc.offeringRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return offeringToDTO(offering), nil
}

func (uc *OfferingUseCaseImpl) GetByUserID(ctx context.Context, UserID uuid.UUID) ([]*dto.OfferingDTO, error) {
	offerings, err := uc.offeringRepo.FindByUserID(ctx, UserID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.OfferingDTO, len(offerings))
	for i := range offerings {
		dtos[i] = offeringToDTO(&offerings[i])
	}

	return dtos, nil
}

func (uc *OfferingUseCaseImpl) UpdateOffering(ctx context.Context, id uuid.UUID, req dto.UpdateOfferingRequest) error {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return err
	}

	if req.Type != nil && *req.Type == domain.OfferingService {
		return domain.ErrInvalidOfferingType
	}

	offering, err := uc.offeringRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Authentication alone is not authorisation: behind this route any valid
	// token reaches the handler, so without the ownership check one farmer could
	// rewrite another's product.
	if err := requireOfferingOwner(principal, offering); err != nil {
		return err
	}

	now := uc.timer.Now()

	if req.Type != nil {
		offering.Type = *req.Type
	}
	if req.Name != nil {
		offering.Name = *req.Name
	}
	if req.Description != nil {
		offering.UpdateDescription(*req.Description, now)
	}
	if req.Price != nil {
		if err := offering.UpdatePrice(*req.Price, now); err != nil {
			return err
		}
	}
	if req.ImageURL != nil {
		offering.UpdateImage(*req.ImageURL, now)
	}
	if req.Latitude != nil || req.Longitude != nil {
		if err := offering.SetLocation(req.Latitude, req.Longitude, now); err != nil {
			return err
		}
	}
	if req.Variety != nil {
		offering.SetVariety(*req.Variety, now)
	}
	if req.UnitOfMeasureID != nil {
		offering.SetUnitOfMeasure(req.UnitOfMeasureID, now)
	}
	if req.QuantityAvailable != nil {
		offering.SetQuantity(*req.QuantityAvailable, now)
	}
	if req.ExpiresAt != nil {
		offering.SetExpiry(req.ExpiresAt, now)
	}
	if req.CategoryID != nil {
		offering.SetCategory(req.CategoryID, now)
	}
	if req.CompanyID != nil {
		offering.SetCompany(req.CompanyID, now)
	}

	if err := offering.RequirePublishable(); err != nil {
		return err
	}

	offering.Touch(now)

	if err := uc.offeringRepo.Update(ctx, offering); err != nil {
		return err
	}

	// Update in Elasticsearch — rebuild the full index document
	user, err := uc.userRepo.FindByID(ctx, offering.UserID)
	if err != nil {
		// Log but don't fail — PG update already succeeded
		return nil
	}

	_ = uc.fuzzyRetrival.Update(ctx, id.String(), indexRequestFor(offering, user))

	_ = uc.searchInvalidator.InvalidateAll(ctx)

	return nil
}

func (uc *OfferingUseCaseImpl) DeleteOffering(ctx context.Context, id uuid.UUID) error {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return err
	}

	// The offering has to be loaded to know who owns it. Doing so also turns a
	// delete of an id that does not exist into ErrNotFound instead of a silent
	// success, which is the error the moderation caller already handles.
	offering, err := uc.offeringRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := requireOfferingOwner(principal, offering); err != nil {
		return err
	}

	err = uc.offeringRepo.Delete(ctx, id)

	if err != nil {
		return fmt.Errorf("Offering Delete error: %w", err)
	}

	err = uc.fuzzyRetrival.Delete(ctx, string(id.String()))

	if err != nil {
		return fmt.Errorf("Offering Fuzzy Delete error: %w", err)
	}

	_ = uc.searchInvalidator.InvalidateAll(ctx)

	return nil
}

func (uc *OfferingUseCaseImpl) DeactivateOffering(ctx context.Context, id uuid.UUID) (*dto.OfferingDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	offering, err := uc.offeringRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := requireOfferingOwner(principal, offering); err != nil {
		return nil, err
	}

	offering.Deactivate()

	if err := uc.offeringRepo.Update(ctx, offering); err != nil {
		return nil, err
	}

	_ = uc.searchInvalidator.InvalidateAll(ctx)

	return offeringToDTO(offering), nil
}

// requireOfferingOwner refuses a caller that does not own the offering.
//
// An admin is allowed through because the moderation endpoint
// (DELETE /api/v1/admin/offerings/{id}) delegates here; without that carve-out
// the admin delete would start failing the moment ownership was checked.
func requireOfferingOwner(principal auth.Principal, offering *domain.Offering) error {
	if offering.UserID != principal.UserID && principal.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	return nil
}

var _ primary.OfferingUseCase = (*OfferingUseCaseImpl)(nil)

// indexRequestFor projects an offering and its producer onto the search
// document. Create and update share it so the two paths can never disagree
// about the document shape. The geo_point is derived by the search adapter from
// the latitude and longitude set here, next to the mapping and the proximity
// sort that have to agree on the same field name.
func indexRequestFor(offering *domain.Offering, user *domain.User) *dto.IndexOfferingRequest {
	farmerVerified := user.HasRole(domain.RoleProvider) || user.HasRole(domain.RoleMIPYME)

	latitude, longitude := user.Address.Latitude, user.Address.Longitude
	if offering.HasLocation() {
		latitude, longitude = *offering.Latitude, *offering.Longitude
	}

	categoryID := ""
	if offering.CategoryID != nil {
		categoryID = offering.CategoryID.String()
	}

	return &dto.IndexOfferingRequest{
		ID:             offering.ID.String(),
		Name:           offering.Name,
		Description:    offering.Description,
		Price:          offering.Price,
		Type:           offering.Type.String(),
		CategoryID:     categoryID,
		ImageURL:       offering.ImageURL,
		UserID:         offering.UserID.String(),
		FarmerName:     user.FullName(),
		FarmerVerified: farmerVerified,
		Department:     user.Address.Department,
		Municipality:   user.Address.Municipality,
		Latitude:       latitude,
		Longitude:      longitude,
	}
}

func offeringToDTO(offering *domain.Offering) *dto.OfferingDTO {
	return &dto.OfferingDTO{
		ID:                offering.ID,
		UserID:            offering.UserID,
		Type:              offering.Type,
		Name:              offering.Name,
		Description:       offering.Description,
		Price:             offering.Price,
		ImageURL:          offering.ImageURL,
		Variety:           offering.Variety,
		UnitOfMeasureID:   offering.UnitOfMeasureID,
		QuantityAvailable: offering.QuantityAvailable,
		ExpiresAt:         offering.ExpiresAt,
		IsActive:          offering.IsActive,
		CategoryID:        offering.CategoryID,
		CompanyID:         offering.CompanyID,
		Latitude:          offering.Latitude,
		Longitude:         offering.Longitude,
		CreatedAt:         offering.CreatedAt,
		UpdatedAt:         offering.UpdatedAt,
	}
}
