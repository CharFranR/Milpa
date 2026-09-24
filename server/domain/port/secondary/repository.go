package port

import (
	"context"

	"milpa/aplication/dto"
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
	FindByUserID(ctx context.Context, companyID uuid.UUID) ([]domain.Offering, error)
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
	FindByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.Inquiry, error)
	Save(ctx context.Context, inquiry *domain.Inquiry) error
	Update(ctx context.Context, inquiry *domain.Inquiry) error
}

type LiquidationRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Liquidation, error)
	FindBySupplier(ctx context.Context, supplierID uuid.UUID) ([]domain.Liquidation, error)
	FindOpen(ctx context.Context) ([]domain.Liquidation, error)
	Save(ctx context.Context, liquidation *domain.Liquidation) error
	Update(ctx context.Context, liquidation *domain.Liquidation) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ReportRepository interface {
	Save(ctx context.Context, report *domain.Report) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Report, error)
	FindAll(ctx context.Context, status string, targetType string, page, pageSize int) ([]domain.Report, int, error)
	Resolve(ctx context.Context, report *domain.Report) error
	ExistsPendingByTarget(ctx context.Context, reporterID uuid.UUID, targetType domain.ReportTargetType, targetID uuid.UUID) (bool, error)
}

type AuditLogRepository interface {
	Save(ctx context.Context, log *domain.AuditLog) error
	FindAll(ctx context.Context, action string, actorID string, targetType string, page, pageSize int) ([]domain.AuditLog, int, error)
}

type ImageStore interface {
	Upload(ctx context.Context, file []byte, filename string) (string, error)
	Load(ctx context.Context, filename string) (*dto.ImageDataDTO, error)
	Delete(ctx context.Context, filename string) error
}
