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

type ConversationRepository interface {
	Save(ctx context.Context, convesation *domain.Conversation) error
	List(ctx context.Context, userID uuid.UUID) ([]domain.Conversation, error)
	ListMessage(ctx context.Context, conversastionID uuid.UUID) ([]domain.Message, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type MessageRepository interface {
	Save(ctx context.Context, message *domain.Message) error
	BulkSave(ctx context.Context, messages *[]domain.Message) error
	ListByConversationID(ctx context.Context, conversationID uuid.UUID) ([]domain.Message, error)
	GetMessageByID(ctx context.Context, id uuid.UUID) (*domain.Message, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type SupplyOfferRepository interface {
	Create(ctx context.Context, SupplyOffer *domain.SupplyOffer) error
	List(ctx context.Context, supplierID uuid.UUID) ([]domain.SupplyOffer, error)
	GetByID(ctx context.Context, SupplyOfferId uuid.UUID) (domain.SupplyOffer, error)
	Update(ctx context.Context, SuplyOffer *domain.SupplyOffer) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type SupplyRequestRepository interface {
	Create(ctx context.Context, supplyRequest *domain.SupplyRequest) error
	List(ctx context.Context, buyerID uuid.UUID) ([]domain.SupplyRequest, error)
	GetByID(ctx context.Context, supplyRequest uuid.UUID) (domain.SupplyRequest, error)
	Update(ctx context.Context, suplyRequest *domain.SupplyRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type MatchRepository interface {
	Create(ctx context.Context, match *domain.Match) error
	ListByOffer(ctx context.Context, supplyOfferID uuid.UUID) ([]domain.Match, error)
	ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]domain.Match, error)
	GetByID(ctx context.Context, matchID uuid.UUID) (*domain.Match, error)
	Update(ctx context.Context, match *domain.Match) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *domain.Transaction) error
	List(ctx context.Context, matchID uuid.UUID) ([]domain.Transaction, error)
	GetByID(ctx context.Context, transactionID uuid.UUID) (domain.Transaction, error)
	Update(ctx context.Context, transaction *domain.Transaction) error
	Delete(ctx context.Context, id uuid.UUID) error
}
