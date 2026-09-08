package usecases_test

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
	"milpa/internal/auth"
)

var (
	errFake        = errors.New("fake error")
	fixedTime      = time.Date(2026, 8, 12, 10, 0, 0, 0, time.UTC)
	testUserID     = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	testCompanyID  = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	testOfferingID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	testReviewID   = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	testCategoryID = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	testInquiryID  = uuid.MustParse("66666666-6666-6666-6666-666666666666")
	testOtherID    = uuid.MustParse("77777777-7777-7777-7777-777777777777")
)

func principalCtx() context.Context {
	return auth.WithPrincipal(context.Background(), auth.Principal{UserID: testUserID, Role: domain.RoleMIPYME})
}

func strPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

func offeringTypePtr(t domain.OfferingType) *domain.OfferingType {
	return &t
}

func inquiryStatusPtr(s domain.InquiryStatus) *domain.InquiryStatus {
	return &s
}

func mustUser() *domain.User {
	user, err := domain.NewUser("user@milpa.com.ni", "John", "Doe", fixedTime)
	if err != nil {
		panic(err)
	}
	user.ID = testUserID
	return user
}

func mustCompany() *domain.Company {
	company, err := domain.NewCompany(domain.User{ID: testUserID}, "Milpa S.A.", fixedTime)
	if err != nil {
		panic(err)
	}
	company.ID = testCompanyID
	return company
}

func mustOffering() *domain.Offering {
	offering, err := domain.NewOffering(testCompanyID, "Organic Corn", domain.OfferingProduct, fixedTime)
	if err != nil {
		panic(err)
	}
	offering.ID = testOfferingID
	offering.Description = "Fresh organic corn"
	offering.Price = 10.0
	offering.ImageURL = "http://images.milpa.com/corn.png"
	return offering
}

func mustReview() *domain.Review {
	review, err := domain.NewReview(testUserID, testCompanyID, 5, "Great quality", fixedTime)
	if err != nil {
		panic(err)
	}
	review.ID = testReviewID
	return review
}

func mustInquiry() *domain.Inquiry {
	inquiry, err := domain.NewInquiry(testUserID, testOfferingID, "Is it in stock?", fixedTime)
	if err != nil {
		panic(err)
	}
	inquiry.ID = testInquiryID
	return inquiry
}

func mustCategory() *domain.Category {
	return &domain.Category{ID: testCategoryID, Name: "Grains", Description: "Grain products"}
}

type fakeUserRepo struct {
	findByID      func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	findByEmail   func(ctx context.Context, email string) (*domain.User, error)
	existsByEmail func(ctx context.Context, email string) (bool, error)
	existsByID    func(ctx context.Context, id string) (bool, error)
	save          func(ctx context.Context, user *domain.User) (string, error)
	update        func(ctx context.Context, user *domain.User) error
	saved         []*domain.User
	updated       []*domain.User
	existedEmails []string
	existedID     []string
}

func newFakeUserRepo() *fakeUserRepo {
	f := &fakeUserRepo{}
	f.findByID = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
		user := mustUser()
		user.ID = id
		return user, nil
	}
	f.findByEmail = func(ctx context.Context, email string) (*domain.User, error) {
		user := mustUser()
		user.Email = email
		return user, nil
	}
	f.existsByEmail = func(ctx context.Context, email string) (bool, error) {
		f.existedEmails = append(f.existedEmails, email)
		return false, nil
	}

	f.existsByID = func(ctx context.Context, id string) (bool, error) {
		f.existedID = append(f.existedID, id)
		return false, nil
	}

	f.save = func(ctx context.Context, user *domain.User) (string, error) {
		f.saved = append(f.saved, user)
		return "", nil
	}
	f.update = func(ctx context.Context, user *domain.User) error {
		f.updated = append(f.updated, user)
		return nil
	}
	return f
}

func (f *fakeUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return f.findByID(ctx, id)
}

func (f *fakeUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return f.findByEmail(ctx, email)
}

func (f *fakeUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return f.existsByEmail(ctx, email)
}

func (f *fakeUserRepo) ExistsByID(ctx context.Context, id string) (bool, error) {
	return f.existsByEmail(ctx, id)
}

func (f *fakeUserRepo) Save(ctx context.Context, user *domain.User) (string, error) {
	return f.save(ctx, user)
}

func (f *fakeUserRepo) Update(ctx context.Context, user *domain.User) error {
	return f.update(ctx, user)
}

type fakeCompanyRepo struct {
	findByID    func(ctx context.Context, id uuid.UUID) (*domain.Company, error)
	findByOwner func(ctx context.Context, ownerID uuid.UUID) ([]domain.Company, error)
	save        func(ctx context.Context, company *domain.Company) error
	update      func(ctx context.Context, company *domain.Company) error
	saved       []*domain.Company
	updated     []*domain.Company
}

func newFakeCompanyRepo() *fakeCompanyRepo {
	f := &fakeCompanyRepo{}
	f.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
		company := mustCompany()
		company.ID = id
		return company, nil
	}
	f.findByOwner = func(ctx context.Context, ownerID uuid.UUID) ([]domain.Company, error) {
		return []domain.Company{*mustCompany()}, nil
	}
	f.save = func(ctx context.Context, company *domain.Company) error {
		f.saved = append(f.saved, company)
		return nil
	}
	f.update = func(ctx context.Context, company *domain.Company) error {
		f.updated = append(f.updated, company)
		return nil
	}
	return f
}

func (f *fakeCompanyRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
	return f.findByID(ctx, id)
}

func (f *fakeCompanyRepo) FindByOwner(ctx context.Context, ownerID uuid.UUID) ([]domain.Company, error) {
	return f.findByOwner(ctx, ownerID)
}

func (f *fakeCompanyRepo) Save(ctx context.Context, company *domain.Company) error {
	return f.save(ctx, company)
}

func (f *fakeCompanyRepo) Update(ctx context.Context, company *domain.Company) error {
	return f.update(ctx, company)
}

type fakeOfferingRepo struct {
	findByID      func(ctx context.Context, id uuid.UUID) (*domain.Offering, error)
	findByCompany func(ctx context.Context, companyID uuid.UUID) ([]domain.Offering, error)
	save          func(ctx context.Context, offering *domain.Offering) error
	update        func(ctx context.Context, offering *domain.Offering) error
	delete        func(ctx context.Context, id uuid.UUID) error
	saved         []*domain.Offering
	updated       []*domain.Offering
	deleted       []uuid.UUID
}

func newFakeOfferingRepo() *fakeOfferingRepo {
	f := &fakeOfferingRepo{}
	f.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Offering, error) {
		offering := mustOffering()
		offering.ID = id
		return offering, nil
	}
	f.findByCompany = func(ctx context.Context, companyID uuid.UUID) ([]domain.Offering, error) {
		return []domain.Offering{*mustOffering()}, nil
	}
	f.save = func(ctx context.Context, offering *domain.Offering) error {
		f.saved = append(f.saved, offering)
		return nil
	}
	f.update = func(ctx context.Context, offering *domain.Offering) error {
		f.updated = append(f.updated, offering)
		return nil
	}
	f.delete = func(ctx context.Context, id uuid.UUID) error {
		f.deleted = append(f.deleted, id)
		return nil
	}
	return f
}

func (f *fakeOfferingRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Offering, error) {
	return f.findByID(ctx, id)
}

func (f *fakeOfferingRepo) FindByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.Offering, error) {
	return f.findByCompany(ctx, companyID)
}

func (f *fakeOfferingRepo) Save(ctx context.Context, offering *domain.Offering) error {
	return f.save(ctx, offering)
}

func (f *fakeOfferingRepo) Update(ctx context.Context, offering *domain.Offering) error {
	return f.update(ctx, offering)
}

func (f *fakeOfferingRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return f.delete(ctx, id)
}

type fakeReviewRepo struct {
	findByCompany func(ctx context.Context, companyID uuid.UUID) ([]domain.Review, error)
	findByUser    func(ctx context.Context, userID uuid.UUID) ([]domain.Review, error)
	save          func(ctx context.Context, review *domain.Review) error
	saved         []*domain.Review
}

func newFakeReviewRepo() *fakeReviewRepo {
	f := &fakeReviewRepo{}
	f.findByCompany = func(ctx context.Context, companyID uuid.UUID) ([]domain.Review, error) {
		return []domain.Review{*mustReview()}, nil
	}
	f.findByUser = func(ctx context.Context, userID uuid.UUID) ([]domain.Review, error) {
		return []domain.Review{*mustReview()}, nil
	}
	f.save = func(ctx context.Context, review *domain.Review) error {
		f.saved = append(f.saved, review)
		return nil
	}
	return f
}

func (f *fakeReviewRepo) FindByCompany(ctx context.Context, companyID uuid.UUID) ([]domain.Review, error) {
	return f.findByCompany(ctx, companyID)
}

func (f *fakeReviewRepo) FindByUser(ctx context.Context, userID uuid.UUID) ([]domain.Review, error) {
	return f.findByUser(ctx, userID)
}

func (f *fakeReviewRepo) Save(ctx context.Context, review *domain.Review) error {
	return f.save(ctx, review)
}

type fakeCategoryRepo struct {
	findAll  func(ctx context.Context) ([]domain.Category, error)
	findByID func(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	save     func(ctx context.Context, category *domain.Category) error
	saved    []*domain.Category
}

func newFakeCategoryRepo() *fakeCategoryRepo {
	f := &fakeCategoryRepo{}
	f.findAll = func(ctx context.Context) ([]domain.Category, error) {
		return []domain.Category{*mustCategory()}, nil
	}
	f.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
		category := mustCategory()
		category.ID = id
		return category, nil
	}
	f.save = func(ctx context.Context, category *domain.Category) error {
		f.saved = append(f.saved, category)
		return nil
	}
	return f
}

func (f *fakeCategoryRepo) FindAll(ctx context.Context) ([]domain.Category, error) {
	return f.findAll(ctx)
}

func (f *fakeCategoryRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	return f.findByID(ctx, id)
}

func (f *fakeCategoryRepo) Save(ctx context.Context, category *domain.Category) error {
	return f.save(ctx, category)
}

type fakeInquiryRepo struct {
	findByID   func(ctx context.Context, id uuid.UUID) (*domain.Inquiry, error)
	findByUser func(ctx context.Context, userID uuid.UUID) ([]domain.Inquiry, error)
	save       func(ctx context.Context, inquiry *domain.Inquiry) error
	update     func(ctx context.Context, inquiry *domain.Inquiry) error
	saved      []*domain.Inquiry
	updated    []*domain.Inquiry
}

func newFakeInquiryRepo() *fakeInquiryRepo {
	f := &fakeInquiryRepo{}
	f.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Inquiry, error) {
		inquiry := mustInquiry()
		inquiry.ID = id
		return inquiry, nil
	}
	f.findByUser = func(ctx context.Context, userID uuid.UUID) ([]domain.Inquiry, error) {
		return []domain.Inquiry{*mustInquiry()}, nil
	}
	f.save = func(ctx context.Context, inquiry *domain.Inquiry) error {
		f.saved = append(f.saved, inquiry)
		return nil
	}
	f.update = func(ctx context.Context, inquiry *domain.Inquiry) error {
		f.updated = append(f.updated, inquiry)
		return nil
	}
	return f
}

func (f *fakeInquiryRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Inquiry, error) {
	return f.findByID(ctx, id)
}

func (f *fakeInquiryRepo) FindByUser(ctx context.Context, userID uuid.UUID) ([]domain.Inquiry, error) {
	return f.findByUser(ctx, userID)
}

func (f *fakeInquiryRepo) Save(ctx context.Context, inquiry *domain.Inquiry) error {
	return f.save(ctx, inquiry)
}

func (f *fakeInquiryRepo) Update(ctx context.Context, inquiry *domain.Inquiry) error {
	return f.update(ctx, inquiry)
}

type fakeHasher struct {
	hash    func(password string) (string, error)
	compare func(hash, password string) error
	hashed  []string
}

func newFakeHasher() *fakeHasher {
	f := &fakeHasher{}
	f.hash = func(password string) (string, error) {
		f.hashed = append(f.hashed, password)
		return "hashed-" + password, nil
	}
	f.compare = func(hash, password string) error {
		return nil
	}
	return f
}

func (f *fakeHasher) Hash(password string) (string, error) {
	return f.hash(password)
}

func (f *fakeHasher) Compare(hash, password string) error {
	return f.compare(hash, password)
}

type fakeJWT struct {
	generateToken func(userID uuid.UUID, role domain.RoleOptions) (string, error)
	validateToken func(token string) (*port.JWTClaims, error)
	tokenUserID   uuid.UUID
	tokenRole     domain.RoleOptions
}

func newFakeJWT() *fakeJWT {
	f := &fakeJWT{}
	f.generateToken = func(userID uuid.UUID, role domain.RoleOptions) (string, error) {
		f.tokenUserID = userID
		f.tokenRole = role
		return "signed-token", nil
	}
	f.validateToken = func(token string) (*port.JWTClaims, error) {
		return &port.JWTClaims{UserID: testUserID, Role: domain.RolePending}, nil
	}
	return f
}

func (f *fakeJWT) GenerateToken(userID uuid.UUID, role domain.RoleOptions) (string, error) {
	return f.generateToken(userID, role)
}

func (f *fakeJWT) ValidateToken(token string) (*port.JWTClaims, error) {
	return f.validateToken(token)
}

type fakeTimer struct {
	now time.Time
}

func (f fakeTimer) Now() time.Time {
	return f.now
}

func newFakeTimer() fakeTimer {
	return fakeTimer{now: fixedTime}
}
