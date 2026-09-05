package port

import (
	"context"

	domain "milpa/domain/entities"

	"github.com/google/uuid"
)

type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByID(ctx context.Context, id string) (bool, error)
	Save(ctx context.Context, user *domain.User) (string, error)
	Update(ctx context.Context, user *domain.User) error
}

type CompanyRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Company, error)
	FindByOwner(ctx context.Context, ownerID uuid.UUID) ([]domain.Company, error)
	Save(ctx context.Context, company *domain.Company) error
	Update(ctx context.Context, company *domain.Company) error
}

type OfferingRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Offering, error)
	FindByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.Offering, error)
	Save(ctx context.Context, offering *domain.Offering) error
	Update(ctx context.Context, offering *domain.Offering) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ReviewRepository interface {
	FindByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.Review, error)
	FindByUser(ctx context.Context, userID uuid.UUID) ([]domain.Review, error)
	Save(ctx context.Context, review *domain.Review) error
}

type CategoryRepository interface {
	FindAll(ctx context.Context) ([]domain.Category, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	Save(ctx context.Context, category *domain.Category) error
}

type InquiryRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Inquiry, error)
	FindByUser(ctx context.Context, userID uuid.UUID) ([]domain.Inquiry, error)
	Save(ctx context.Context, inquiry *domain.Inquiry) error
	Update(ctx context.Context, inquiry *domain.Inquiry) error
}
