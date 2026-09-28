package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/infrastructure/adapters/primary/api/handler"
	"milpa/infrastructure/adapters/primary/api/middleware"
	"milpa/infrastructure/adapters/primary/api/ws"
	"milpa/internal/auth"
)

// stubPIIUserUC is a user use case that hands back both representations, so the
// assertions below are about what the transport does with them rather than about
// the use case, which is covered in tests/unitary/usecases.
type stubPIIUserUC struct {
	user dto.UserView
	err  error
}

func (s *stubPIIUserUC) Register(ctx context.Context, req dto.RegisterUserRequest) (*dto.PrivateUserDTO, error) {
	return nil, s.err
}

func (s *stubPIIUserUC) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	return nil, s.err
}

func (s *stubPIIUserUC) GetByID(ctx context.Context, id uuid.UUID) (dto.UserView, error) {
	if s.err != nil {
		return nil, s.err
	}
	// Mirrors the policy the real use case applies: self or admin gets the
	// contact card, everyone else gets the public profile.
	principal, ok := auth.FromContext(ctx)
	if ok && (principal.UserID == id || principal.Role == domain.RoleAdmin) {
		return s.user, nil
	}
	return s.publicView(), nil
}

func (s *stubPIIUserUC) publicView() dto.UserView {
	private, _ := s.user.(*dto.PrivateUserDTO)
	if private == nil {
		return s.user
	}
	return &dto.PublicUserDTO{
		ID:           private.ID,
		FirstName:    private.FirstName,
		LastName:     private.LastName,
		Role:         private.Role,
		Department:   private.Department,
		Municipality: private.Municipality,
		CreatedAt:    private.CreatedAt,
		UpdatedAt:    private.UpdatedAt,
	}
}

func (s *stubPIIUserUC) UpdateProfile(ctx context.Context, id uuid.UUID, req dto.UpdateUserRequest) error {
	return s.err
}

type stubPIICompanyUC struct {
	company dto.CompanyView
	err     error
}

func (s *stubPIICompanyUC) CreateCompany(ctx context.Context, req dto.RegisterCompanyRequest) (*dto.PrivateCompanyDTO, error) {
	return nil, s.err
}

func (s *stubPIICompanyUC) GetByID(ctx context.Context, id uuid.UUID) (dto.CompanyView, error) {
	if s.err != nil {
		return nil, s.err
	}
	principal, ok := auth.FromContext(ctx)
	if ok && (principal.UserID == s.ownerID() || principal.Role == domain.RoleAdmin) {
		return s.company, nil
	}
	return s.publicView(), nil
}

func (s *stubPIICompanyUC) GetByOwner(ctx context.Context, ownerID uuid.UUID) ([]dto.CompanyView, error) {
	if s.err != nil {
		return nil, s.err
	}
	principal, ok := auth.FromContext(ctx)
	if ok && (principal.UserID == ownerID || principal.Role == domain.RoleAdmin) {
		return []dto.CompanyView{s.company}, nil
	}
	return []dto.CompanyView{s.publicView()}, nil
}

func (s *stubPIICompanyUC) UpdateCompany(ctx context.Context, id uuid.UUID, req dto.UpdateCompanyRequest) error {
	return s.err
}

func (s *stubPIICompanyUC) ownerID() uuid.UUID {
	private, _ := s.company.(*dto.PrivateCompanyDTO)
	if private == nil {
		return uuid.Nil
	}
	return private.OwnerID
}

func (s *stubPIICompanyUC) publicView() dto.CompanyView {
	private, _ := s.company.(*dto.PrivateCompanyDTO)
	if private == nil {
		return s.company
	}
	return &dto.PublicCompanyDTO{
		ID:           private.ID,
		Name:         private.Name,
		CategoryID:   private.CategoryID,
		OwnerID:      private.OwnerID,
		Department:   private.Department,
		Municipality: private.Municipality,
		Description:  private.Description,
		Website:      private.Website,
		Verified:     private.Verified,
		CreatedAt:    private.CreatedAt,
		UpdatedAt:    private.UpdatedAt,
	}
}

var (
	piiOwnerID  = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	piiOtherID  = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	piiAdminID  = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	piiCompID   = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	piiUserID   = piiOwnerID
	piiCategory = uuid.MustParse("55555555-5555-5555-5555-555555555555")
)

func newPIIRouter(t *testing.T, userUC primary.UserUseCase, companyUC primary.CompanyUseCase) http.Handler {
	t.Helper()

	chat := ws.NewHandler(ws.NewHub(), nil, &stubConversationUC{})
	authMW := middleware.NewAuthMiddleware(piiJWT{})
	suspensionMW := middleware.NewSuspensionMiddleware(stubUserRepo{})

	return NewRouter(
		handler.NewUserHandler(userUC), handler.NewCompanyHandler(companyUC),
		nil, nil, nil, nil, nil,
		authMW, suspensionMW,
		nil, nil, nil, nil, nil, nil, chat,
		nil, nil, nil, nil,
	)
}

// piiJWT encodes the caller in the token as "<userID>:<role>" so one router can
// serve the anonymous, owner, admin and third-party cases. The stock stubJWT
// hands every token the same random identity, which cannot express them.
type piiJWT struct{}

func (piiJWT) GenerateToken(userID uuid.UUID, role domain.RoleOptions) (string, error) {
	return piiToken(userID, role), nil
}

func (piiJWT) ValidateToken(token string) (*port.JWTClaims, error) {
	rawID, rawRole, found := strings.Cut(token, ":")
	if !found {
		return nil, errors.New("invalid or expired token")
	}
	userID, err := uuid.Parse(rawID)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}
	role, err := strconv.Atoi(rawRole)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}
	return &port.JWTClaims{UserID: userID, Role: domain.RoleOptions(role)}, nil
}

func piiToken(userID uuid.UUID, role domain.RoleOptions) string {
	return userID.String() + ":" + strconv.Itoa(int(role))
}

func authorizedGet(path string, userID uuid.UUID, role domain.RoleOptions) *http.Request {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+piiToken(userID, role))
	return req
}

func privateUserFixture() *dto.PrivateUserDTO {
	return &dto.PrivateUserDTO{
		PublicUserDTO: dto.PublicUserDTO{
			ID:           piiUserID,
			FirstName:    "Carlos",
			LastName:     "Ramirez",
			Role:         domain.RoleProvider,
			Department:   "Leon",
			Municipality: "Leon",
		},
		Email:       "carlos@finca.com.ni",
		PhoneNumber: "555-1234",
		AddressLine: "Costado Sur del Parque Central",
	}
}

func privateCompanyFixture() *dto.PrivateCompanyDTO {
	return &dto.PrivateCompanyDTO{
		PublicCompanyDTO: dto.PublicCompanyDTO{
			ID:           piiCompID,
			Name:         "Finca Ramirez",
			CategoryID:   piiCategory,
			OwnerID:      piiOwnerID,
			Department:   "Leon",
			Municipality: "Leon",
			Description:  "Finca cafetalera",
			Website:      "https://finca.example.com",
			Verified:     true,
		},
		Email:       "ventas@finca.example.com",
		PhoneNumber: "555-9999",
		AddressLine: "Barrio San Francisco, Casa 12",
	}
}

func newPIIRouterWithFixtures(t *testing.T) http.Handler {
	t.Helper()
	return newPIIRouter(t,
		&stubPIIUserUC{user: privateUserFixture()},
		&stubPIICompanyUC{company: privateCompanyFixture()},
	)
}

// decodeData unwraps the response envelope so the assertions look at the payload
// an actual client receives.
func decodeData(t *testing.T, body string) map[string]interface{} {
	t.Helper()

	var envelope struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode response %q: %v", body, err)
	}
	if envelope.Data == nil {
		t.Fatalf("response has no data object: %q", body)
	}
	return envelope.Data
}

func assertNoContactFields(t *testing.T, where string, data map[string]interface{}) {
	t.Helper()

	for _, field := range []string{"email", "phone_number", "address_line", "address"} {
		if value, present := data[field]; present {
			t.Errorf("%s exposed %q = %v, want it absent", where, field, value)
		}
	}
}

// TestAnonymousGetsNoContactFields is the acceptance gate: an unauthenticated
// request to either public endpoint must not receive contact details.
func TestAnonymousGetsNoContactFields(t *testing.T) {
	t.Parallel()

	router := newPIIRouterWithFixtures(t)

	tests := []struct {
		name string
		path string
	}{
		{name: "user by id", path: "/api/v1/users/" + piiUserID.String()},
		{name: "company by id", path: "/api/v1/companies/" + piiCompID.String()},
		{name: "companies by owner", path: "/api/v1/companies/?owner_id=" + piiOwnerID.String()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
			}

			if strings.Contains(rr.Body.String(), "carlos@finca.com.ni") {
				t.Errorf("response contains the owner's email: %s", rr.Body.String())
			}
			if strings.Contains(rr.Body.String(), "555-1234") {
				t.Errorf("response contains the owner's phone: %s", rr.Body.String())
			}
			if strings.Contains(rr.Body.String(), "Costado Sur del Parque Central") {
				t.Errorf("response contains the owner's address line: %s", rr.Body.String())
			}

			var envelope struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			var single map[string]interface{}
			if err := json.Unmarshal(envelope.Data, &single); err == nil {
				assertNoContactFields(t, tt.name, single)
				return
			}

			var list []map[string]interface{}
			if err := json.Unmarshal(envelope.Data, &list); err != nil {
				t.Fatalf("decode response as object or list: %v", err)
			}
			if len(list) == 0 {
				t.Fatalf("response carried no entries: %s", rr.Body.String())
			}
			for i := range list {
				assertNoContactFields(t, tt.name+"["+string(rune('0'+i))+"]", list[i])
			}
		})
	}
}

// TestOwnUserGetsContactFields is the other half: the owner must still be able
// to read their own contact card, otherwise the marketplace has no way to work
// at all.
func TestOwnUserGetsContactFields(t *testing.T) {
	t.Parallel()

	router := newPIIRouterWithFixtures(t)

	req := authorizedGet("/api/v1/users/"+piiOwnerID.String(), piiOwnerID, domain.RoleProvider)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}

	data := decodeData(t, rr.Body.String())
	if data["email"] != "carlos@finca.com.ni" {
		t.Errorf("email = %v, want the owner's own email", data["email"])
	}
	if data["phone_number"] != "555-1234" {
		t.Errorf("phone_number = %v, want the owner's own phone", data["phone_number"])
	}
	if data["address_line"] != "Costado Sur del Parque Central" {
		t.Errorf("address_line = %v, want the owner's own address line", data["address_line"])
	}
	// The private view also carries the public representation.
	if data["first_name"] != "Carlos" || data["department"] != "Leon" {
		t.Errorf("private view lost the public fields: %v", data)
	}
}

func TestAdminGetsContactFields(t *testing.T) {
	t.Parallel()

	router := newPIIRouterWithFixtures(t)

	req := authorizedGet("/api/v1/users/"+piiUserID.String(), piiAdminID, domain.RoleAdmin)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}

	data := decodeData(t, rr.Body.String())
	if data["email"] != "carlos@finca.com.ni" {
		t.Errorf("email = %v, want the contact card for an admin", data["email"])
	}
	if data["phone_number"] != "555-1234" {
		t.Errorf("phone_number = %v, want the contact card for an admin", data["phone_number"])
	}
}

// TestThirdAuthenticatedUserGetsNoContactFields closes the loop: a token alone
// buys nothing, it has to be the right token.
func TestThirdAuthenticatedUserGetsNoContactFields(t *testing.T) {
	t.Parallel()

	router := newPIIRouterWithFixtures(t)

	req := authorizedGet("/api/v1/users/"+piiUserID.String(), piiOtherID, domain.RoleMIPYME)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}

	if strings.Contains(rr.Body.String(), "carlos@finca.com.ni") {
		t.Errorf("a third party received the owner's email: %s", rr.Body.String())
	}
	assertNoContactFields(t, "third party", decodeData(t, rr.Body.String()))
}

func TestOwnerOfCompanyGetsContactFields(t *testing.T) {
	t.Parallel()

	router := newPIIRouterWithFixtures(t)

	req := authorizedGet("/api/v1/companies/"+piiCompID.String(), piiOwnerID, domain.RoleProvider)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}

	data := decodeData(t, rr.Body.String())
	if data["email"] != "ventas@finca.example.com" {
		t.Errorf("email = %v, want the company's own email", data["email"])
	}
	if data["phone_number"] != "555-9999" {
		t.Errorf("phone_number = %v, want the company's own phone", data["phone_number"])
	}
	if data["address_line"] != "Barrio San Francisco, Casa 12" {
		t.Errorf("address_line = %v, want the company's own address line", data["address_line"])
	}
	// verified belongs to the public representation, so it travels with both.
	if data["verified"] != true {
		t.Errorf("verified = %v, want true on the private view too", data["verified"])
	}
}

func TestThirdUserGetsNoContactFieldsForCompany(t *testing.T) {
	t.Parallel()

	router := newPIIRouterWithFixtures(t)

	req := authorizedGet("/api/v1/companies/"+piiCompID.String(), piiOtherID, domain.RoleMIPYME)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}

	assertNoContactFields(t, "third party company read", decodeData(t, rr.Body.String()))
}

func TestThirdUserGetsNoContactFieldsForCompaniesByOwner(t *testing.T) {
	t.Parallel()

	router := newPIIRouterWithFixtures(t)

	req := authorizedGet("/api/v1/companies/?owner_id="+piiOwnerID.String(), piiOtherID, domain.RoleMIPYME)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rr.Code, rr.Body.String())
	}

	if strings.Contains(rr.Body.String(), "ventas@finca.example.com") {
		t.Errorf("a third party received the company's email through the owner listing: %s", rr.Body.String())
	}
}

// TestInvalidTokenDegradesToAnonymous pins that optional authentication never
// turns a public page into a 401, and never upgrades a request either.
func TestInvalidTokenDegradesToAnonymous(t *testing.T) {
	t.Parallel()

	router := newPIIRouterWithFixtures(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+piiUserID.String(), nil)
	req.Header.Set("Authorization", "Bearer garbage-token")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 for an unusable token on a public read", rr.Code)
	}
	assertNoContactFields(t, "invalid token", decodeData(t, rr.Body.String()))
}
