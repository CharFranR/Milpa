package usecases_test

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/internal/auth"
)

// piiFixture is a user whose contact details are the ones the assertions look
// for: they must not appear in a public response and must appear in a private
// one.
func piiFixture() *domain.User {
	user := mustUser()
	user.Address = domain.Address{
		ID:           uuid.MustParse("88888888-8888-8888-8888-888888888888"),
		Department:   "Leon",
		Municipality: "Leon",
		AddressLine:  "Costado Sur del Parque Central",
	}
	return user
}

// TestUserGetByIDAppliesContactBoundary is the policy itself: which concrete
// representation each caller gets.
func TestUserGetByIDAppliesContactBoundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ctx       context.Context
		wantPrivt bool
	}{
		{name: "anonymous", ctx: context.Background()},
		{name: "own user", ctx: principalCtx(), wantPrivt: true},
		{name: "admin", ctx: reportAdminCtx(), wantPrivt: true},
		{name: "third authenticated user", ctx: principalCtxFor(testOtherID)},
		{name: "third party with provider role", ctx: auth.WithPrincipal(context.Background(), auth.Principal{UserID: testOtherID, Role: domain.RoleAgricultor})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userRepo := newFakeUserRepo()
			fixture := piiFixture()
			userRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
				return fixture, nil
			}
			uc := usecases.NewUserUseCase(userRepo, newFakeHasher(), newFakeJWT(), newFakeTimer())

			view, err := uc.GetByID(tt.ctx, testUserID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			_, isPrivate := view.(*dto.PrivateUserDTO)
			if isPrivate != tt.wantPrivt {
				t.Fatalf("view = %T, private = %v, want %v", view, isPrivate, tt.wantPrivt)
			}

			encoded, err := json.Marshal(view)
			if err != nil {
				t.Fatalf("marshal view: %v", err)
			}
			body := string(encoded)

			if !tt.wantPrivt {
				for _, secret := range []string{"email", "phone_number", "address_line"} {
					if strings.Contains(body, `"`+secret+`"`) {
						t.Errorf("public view exposed %q: %s", secret, body)
					}
				}
				if strings.Contains(body, "user@milpa.com.ni") {
					t.Errorf("public view exposed the email value: %s", body)
				}
				if strings.Contains(body, "Costado Sur del Parque Central") {
					t.Errorf("public view exposed the address line value: %s", body)
				}
				// The location a buyer searches by is still there.
				for _, field := range []string{`"department":"Leon"`, `"municipality":"Leon"`} {
					if !strings.Contains(body, field) {
						t.Errorf("public view lost %s: %s", field, body)
					}
				}
				return
			}

			for _, want := range []string{`"email":"user@milpa.com.ni"`, `"phone_number"`, `"address_line":"Costado Sur del Parque Central"`} {
				if !strings.Contains(body, want) {
					t.Errorf("private view missing %s: %s", want, body)
				}
			}
		})
	}
}

func TestCompanyGetByIDAppliesContactBoundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ctx       context.Context
		wantPrivt bool
	}{
		{name: "anonymous", ctx: context.Background()},
		{name: "own company", ctx: principalCtx(), wantPrivt: true},
		{name: "admin", ctx: reportAdminCtx(), wantPrivt: true},
		{name: "third authenticated user", ctx: principalCtxFor(testOtherID)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			companyRepo := newFakeCompanyRepo()
			fixture := mustCompany()
			fixture.Email = "info@milpa.com.ni"
			fixture.PhoneNumber = "555-9999"
			fixture.Address = domain.Address{
				Department:   "Matagalpa",
				Municipality: "Matagalpa",
				AddressLine:  "Barrio San Francisco, Casa 12",
			}
			companyRepo.findByID = func(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
				return fixture, nil
			}

			uc := usecases.NewCompanyUseCase(companyRepo, newFakeUserRepo(), newFakeCategoryRepo(), newFakeTimer())

			view, err := uc.GetByID(tt.ctx, testCompanyID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			_, isPrivate := view.(*dto.PrivateCompanyDTO)
			if isPrivate != tt.wantPrivt {
				t.Fatalf("view = %T, private = %v, want %v", view, isPrivate, tt.wantPrivt)
			}

			encoded, _ := json.Marshal(view)
			body := string(encoded)

			if !tt.wantPrivt {
				for _, secret := range []string{`"email"`, `"phone_number"`, `"address_line"`, "info@milpa.com.ni", "555-9999", "Casa 12"} {
					if strings.Contains(body, secret) {
						t.Errorf("public view exposed %q: %s", secret, body)
					}
				}
				// verified belongs to the public representation.
				if !strings.Contains(body, `"verified"`) {
					t.Errorf("public view lost verified: %s", body)
				}
				return
			}

			for _, want := range []string{`"email":"info@milpa.com.ni"`, `"phone_number":"555-9999"`, `"address_line":"Barrio San Francisco, Casa 12"`, `"verified"`} {
				if !strings.Contains(body, want) {
					t.Errorf("private view missing %s: %s", want, body)
				}
			}
		})
	}
}

func TestCompanyGetByOwnerAppliesContactBoundary(t *testing.T) {
	t.Parallel()

	companyRepo := newFakeCompanyRepo()
	fixture := mustCompany()
	fixture.Email = "info@milpa.com.ni"
	companyRepo.findByOwner = func(ctx context.Context, ownerID uuid.UUID) ([]domain.Company, error) {
		return []domain.Company{*fixture}, nil
	}

	uc := usecases.NewCompanyUseCase(companyRepo, newFakeUserRepo(), newFakeCategoryRepo(), newFakeTimer())

	tests := []struct {
		name      string
		ctx       context.Context
		wantPrivt bool
	}{
		{name: "anonymous", ctx: context.Background()},
		{name: "owner", ctx: principalCtx(), wantPrivt: true},
		{name: "admin", ctx: reportAdminCtx(), wantPrivt: true},
		{name: "third party", ctx: principalCtxFor(testOtherID)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			views, err := uc.GetByOwner(tt.ctx, testUserID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(views) != 1 {
				t.Fatalf("views = %d, want 1", len(views))
			}

			_, isPrivate := views[0].(*dto.PrivateCompanyDTO)
			if isPrivate != tt.wantPrivt {
				t.Fatalf("view = %T, private = %v, want %v", views[0], isPrivate, tt.wantPrivt)
			}

			encoded, _ := json.Marshal(views[0])
			if !tt.wantPrivt && strings.Contains(string(encoded), "info@milpa.com.ni") {
				t.Errorf("public listing exposed the company email: %s", encoded)
			}
		})
	}
}

// TestPublicDTOCannotCarryContactFields states the invariant in the type system:
// the public representation has no field to leak through, so the boundary is
// structural rather than a field that is sometimes populated. The assertion is
// reflection because a struct literal naming those fields would not compile —
// which is exactly the guarantee.
func TestPublicDTOCannotCarryContactFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   any
		forbid  []string
		require []string
	}{
		{
			name:    "public user",
			value:   dto.PublicUserDTO{ID: testUserID, FirstName: "John", Department: "Leon"},
			forbid:  []string{"Email", "PhoneNumber", "AddressLine"},
			require: []string{"ID", "FirstName", "LastName", "Department", "Municipality", "Role"},
		},
		{
			name:    "public company",
			value:   dto.PublicCompanyDTO{ID: testCompanyID, Name: "Milpa S.A."},
			forbid:  []string{"Email", "PhoneNumber", "AddressLine"},
			require: []string{"ID", "Name", "OwnerID", "Verified", "Department", "Municipality"},
		},
		{
			name:    "private user adds exactly the contact card",
			value:   dto.PrivateUserDTO{},
			require: []string{"Email", "PhoneNumber", "AddressLine", "PublicUserDTO"},
		},
		{
			name:    "private company adds exactly the contact card",
			value:   dto.PrivateCompanyDTO{},
			require: []string{"Email", "PhoneNumber", "AddressLine", "PublicCompanyDTO"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			typ := reflect.TypeOf(tt.value)
			fields := map[string]bool{}
			for i := range typ.NumField() {
				fields[typ.Field(i).Name] = true
			}

			for _, name := range tt.forbid {
				if fields[name] {
					t.Errorf("%s has a %s field, so a public response can carry it", typ.Name(), name)
				}
			}
			for _, name := range tt.require {
				if !fields[name] {
					t.Errorf("%s is missing the %s field", typ.Name(), name)
				}
			}
		})
	}
}

// TestPublicViewSerialisesNoContactFields is the JSON-level counterpart.
func TestPublicViewSerialisesNoContactFields(t *testing.T) {
	t.Parallel()

	encoded, err := json.Marshal(&dto.PublicUserDTO{ID: testUserID, FirstName: "John", Department: "Leon", Municipality: "Leon"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, field := range []string{"email", "phone_number", "address_line", "address"} {
		if _, present := decoded[field]; present {
			t.Errorf("PublicUserDTO serialised %q: %s", field, encoded)
		}
	}
	for _, field := range []string{"id", "first_name", "department", "municipality"} {
		if _, present := decoded[field]; !present {
			t.Errorf("PublicUserDTO lost %q: %s", field, encoded)
		}
	}
}
