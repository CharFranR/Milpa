package usecases

import (
	"context"
	"slices"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/auth"
)

type ReviewUseCaseImpl struct {
	reviewRepo        port.ReviewRepository
	txRepo            port.TransactionRepository
	matchRepo         port.MatchRepository
	supplyOfferRepo   port.SupplyOfferRepository
	supplyRequestRepo port.SupplyRequestRepository
	companyRepo       port.CompanyRepository
	timer             port.TimeProvider
}

func NewReviewUseCase(reviewRepo port.ReviewRepository, txRepo port.TransactionRepository, matchRepo port.MatchRepository, supplyOfferRepo port.SupplyOfferRepository, supplyRequestRepo port.SupplyRequestRepository, companyRepo port.CompanyRepository, timer port.TimeProvider) *ReviewUseCaseImpl {
	return &ReviewUseCaseImpl{
		reviewRepo:        reviewRepo,
		txRepo:            txRepo,
		matchRepo:         matchRepo,
		supplyOfferRepo:   supplyOfferRepo,
		supplyRequestRepo: supplyRequestRepo,
		companyRepo:       companyRepo,
		timer:             timer,
	}
}

func (uc *ReviewUseCaseImpl) CreateReview(ctx context.Context, req dto.CreateReviewRequest) (*dto.ReviewDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	if req.TransactionID == uuid.Nil {
		return nil, domain.ErrTransactionRequired
	}

	transaction, err := uc.txRepo.GetByID(ctx, req.TransactionID)
	if err != nil {
		return nil, err
	}
	if transaction.Status != domain.TransactionCompleted {
		return nil, domain.ErrTransactionNotCompleted
	}

	match, err := uc.matchRepo.GetByID(ctx, transaction.MatchID)
	if err != nil {
		return nil, err
	}

	offer, err := uc.supplyOfferRepo.GetByID(ctx, match.SupplyOffer)
	if err != nil {
		return nil, err
	}
	request, err := uc.supplyRequestRepo.GetByID(ctx, match.SupplyRequest)
	if err != nil {
		return nil, err
	}

	if principal.UserID != request.BuyerID && principal.UserID != offer.SupplierID {
		return nil, domain.ErrForbidden
	}

	targetType, targetID, err := resolveReviewTarget(req)
	if err != nil {
		return nil, err
	}
	companyID := uuid.Nil
	if targetType == domain.ReviewTargetCompany {
		companyID = targetID
	}

	other := offer.SupplierID
	if principal.UserID == offer.SupplierID {
		other = request.BuyerID
	}

	switch targetType {
	case domain.ReviewTargetUser:
		if targetID != other {
			return nil, domain.ErrReviewTargetPartyMismatch
		}
	case domain.ReviewTargetCompany:
		if err := uc.checkCompanyTarget(ctx, principal.UserID, other, targetID); err != nil {
			return nil, err
		}
	}

	now := uc.timer.Now()

	review, err := domain.NewReview(principal.UserID, targetType, targetID, companyID, req.Rating, req.Comment, now, req.TransactionID)
	if err != nil {
		return nil, err
	}

	exists, err := uc.reviewRepo.ExistsByTransactionAndAuthor(ctx, req.TransactionID, principal.UserID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrReviewAlreadyExists
	}

	if err := uc.reviewRepo.Save(ctx, review); err != nil {
		return nil, err
	}

	return reviewToDTO(review), nil
}

func (uc *ReviewUseCaseImpl) checkCompanyTarget(ctx context.Context, authorID, otherID, targetID uuid.UUID) error {
	companies, err := uc.companyRepo.FindByOwner(ctx, otherID)
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(companies, func(company domain.Company) bool { return company.ID == targetID }) {
		return domain.ErrReviewTargetPartyMismatch
	}

	authorCompanies, err := uc.companyRepo.FindByOwner(ctx, authorID)
	if err != nil {
		return err
	}
	if slices.ContainsFunc(authorCompanies, func(company domain.Company) bool { return company.ID == targetID }) {
		return domain.ErrReviewTargetPartyMismatch
	}

	return nil
}

func resolveReviewTarget(req dto.CreateReviewRequest) (domain.ReviewTargetType, uuid.UUID, error) {
	hasLegacy := req.CompanyID != uuid.Nil
	hasExplicit := req.TargetID != uuid.Nil || req.TargetType != ""

	if hasLegacy && hasExplicit {
		return "", uuid.Nil, domain.ErrReviewTargetMismatch
	}
	if hasLegacy {
		return domain.ReviewTargetCompany, req.CompanyID, nil
	}
	if req.TargetID == uuid.Nil {
		return "", uuid.Nil, domain.ErrTargetRequired
	}

	targetType := domain.ReviewTargetType(req.TargetType)
	if !domain.ValidReviewTargetType(targetType) {
		return "", uuid.Nil, domain.ErrInvalidReviewTargetType
	}
	return targetType, req.TargetID, nil
}

func (uc *ReviewUseCaseImpl) FindByUser(ctx context.Context, userID uuid.UUID) ([]*dto.ReviewDTO, error) {
	reviews, err := uc.reviewRepo.FindByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.ReviewDTO, len(reviews))
	for i := range reviews {
		dtos[i] = reviewToDTO(&reviews[i])
	}

	return dtos, nil
}

func (uc *ReviewUseCaseImpl) FindByCompany(ctx context.Context, companyID uuid.UUID) ([]*dto.ReviewDTO, error) {
	reviews, err := uc.reviewRepo.FindByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.ReviewDTO, len(reviews))
	for i := range reviews {
		dtos[i] = reviewToDTO(&reviews[i])
	}

	return dtos, nil
}

func (uc *ReviewUseCaseImpl) FindByTarget(ctx context.Context, targetType domain.ReviewTargetType, targetID uuid.UUID) ([]*dto.ReviewDTO, error) {
	if !domain.ValidReviewTargetType(targetType) {
		return nil, domain.ErrInvalidReviewTargetType
	}
	if targetID == uuid.Nil {
		return nil, domain.ErrTargetRequired
	}

	reviews, err := uc.reviewRepo.FindByTarget(ctx, targetType, targetID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.ReviewDTO, len(reviews))
	for i := range reviews {
		dtos[i] = reviewToDTO(&reviews[i])
	}

	return dtos, nil
}

func (uc *ReviewUseCaseImpl) GetAverageRating(ctx context.Context, targetType domain.ReviewTargetType, targetID uuid.UUID) (*dto.ReviewAverageDTO, error) {
	if !domain.ValidReviewTargetType(targetType) {
		return nil, domain.ErrInvalidReviewTargetType
	}
	if targetID == uuid.Nil {
		return nil, domain.ErrTargetRequired
	}

	average, count, err := uc.reviewRepo.AverageRating(ctx, targetType, targetID)
	if err != nil {
		return nil, err
	}

	return &dto.ReviewAverageDTO{
		TargetType: string(targetType),
		TargetID:   targetID,
		Average:    average,
		Count:      count,
	}, nil
}

var _ primary.ReviewUseCase = (*ReviewUseCaseImpl)(nil)

func reviewToDTO(review *domain.Review) *dto.ReviewDTO {
	return &dto.ReviewDTO{
		ID:            review.ID,
		UserID:        review.AuthorID,
		CompanyID:     review.CompanyID,
		Rating:        review.Rating,
		Comment:       review.Comment,
		CreatedAt:     review.CreatedAt,
		TargetType:    string(review.TargetType),
		TargetID:      review.TargetID,
		TransactionID: review.TransactionID,
	}
}
