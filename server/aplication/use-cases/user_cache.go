package usecases

import (
	"context"
	"time"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/auth"

	"github.com/google/uuid"
)

type CachedUserUseCase struct {
	next  primary.UserUseCase
	cache port.Cache
}

func NewCachedUserUseCase(next primary.UserUseCase, cache port.Cache) *CachedUserUseCase {
	return &CachedUserUseCase{
		next:  next,
		cache: cache,
	}
}

func (uc *CachedUserUseCase) Register(ctx context.Context, req dto.RegisterUserRequest) (*dto.PrivateUserDTO, error) {
	return uc.next.Register(ctx, req)
}

func (uc *CachedUserUseCase) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	return uc.next.Login(ctx, req)
}

// cachedUserView is the union of the two representations a read can resolve to.
// It is not itself a UserView: it is the envelope the cache stores, so a read
// that resolved to one view can be written and read back without the cache
// having to know which one the policy will pick.
type cachedUserView struct {
	Public  *dto.PublicUserDTO  `json:"public"`
	Private *dto.PrivateUserDTO `json:"private"`
}

func cacheUserView(view dto.UserView) cachedUserView {
	switch v := view.(type) {
	case *dto.PrivateUserDTO:
		return cachedUserView{Private: v}
	default:
		public, _ := view.(*dto.PublicUserDTO)
		return cachedUserView{Public: public}
	}
}

func (uc *CachedUserUseCase) GetByID(ctx context.Context, id uuid.UUID) (dto.UserView, error) {
	var cached cachedUserView

	_, err := uc.cache.Remember(
		ctx,
		viewerCacheKey(ctx, "user:", id),
		5*time.Minute,
		&cached,
		func() error {
			result, err := uc.next.GetByID(ctx, id)
			if err != nil {
				return err
			}

			cached = cacheUserView(result)
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if cached.Private != nil {
		return cached.Private, nil
	}
	if cached.Public != nil {
		return cached.Public, nil
	}
	return nil, nil
}

func (uc *CachedUserUseCase) UpdateProfile(ctx context.Context, id uuid.UUID, req dto.UpdateUserRequest) error {
	err := uc.next.UpdateProfile(ctx, id, req)
	if err != nil {
		return err
	}

	_ = uc.cache.DeleteByPrefix(ctx, "user:"+id.String()+":")

	return nil
}

var _ primary.UserUseCase = (*CachedUserUseCase)(nil)

// viewerCacheKey scopes a cached read to the viewer it was produced for.
//
// A read behind this key resolves to either a public profile or a private
// contact card depending on who is asking, so a key derived only from the
// entity id would let the first viewer's response be served to the next one —
// which would switch the contact-detail boundary back off for as long as the
// entry lived. Keying by the viewer means an entry is only ever returned to the
// caller it was written for. Anonymous callers share the one "anonymous" slot,
// which is safe precisely because none of them can ever be entitled: being the
// owner or an admin both require a principal.
func viewerCacheKey(ctx context.Context, prefix string, id uuid.UUID) string {
	if principal, ok := auth.FromContext(ctx); ok {
		return prefix + id.String() + ":as-" + principal.UserID.String()
	}
	return prefix + id.String() + ":anonymous"
}
