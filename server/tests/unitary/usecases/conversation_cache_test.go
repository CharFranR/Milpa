package usecases_test

import (
	"slices"
	"testing"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
)

// The farmer's inbox is cached under their own key. Evicting only the buyer's
// key is what made a product inquiry take up to five minutes to show up for the
// agricultor: their list kept serving a snapshot from before the chat existed.
func TestConversationCacheEvictsBothInboxes(t *testing.T) {
	t.Parallel()

	cache := newFakeCache()
	uc := usecases.NewCachedConversationUseCase(
		usecases.NewConversationUseCase(newFakeConversationRepo(), newFakeOfferingRepo(), newFakeFarmerUserRepo(), newFakeTimer()),
		cache,
	)

	req := dto.CreateConversationDTO{FarmerID: testCompanyID, OfferingID: testOfferingID}
	result, err := uc.CreateConversation(principalCtx(), req)
	if err != nil {
		t.Fatalf("CreateConversation() error: %v", err)
	}

	for _, want := range []string{
		"conversations:byuser:" + result.BuyerID.String(),
		"conversations:byuser:" + result.FarmerID.String(),
	} {
		if !slices.Contains(cache.deletedKeys, want) {
			t.Errorf("deleted keys = %v, want %q", cache.deletedKeys, want)
		}
	}
}

// A rejected create changed nothing, so the cached lists are still valid.
func TestConversationCacheKeepsInboxesOnRejection(t *testing.T) {
	t.Parallel()

	cache := newFakeCache()
	uc := usecases.NewCachedConversationUseCase(
		usecases.NewConversationUseCase(newFakeConversationRepo(), newFakeOfferingRepo(), newFakeFarmerUserRepo(), newFakeTimer()),
		cache,
	)

	if _, err := uc.CreateConversation(farmerCtx(), dto.CreateConversationDTO{FarmerID: testCompanyID, OfferingID: testOfferingID}); err == nil {
		t.Fatal("CreateConversation() error = nil, want forbidden for an agricultor")
	}

	if len(cache.deletedKeys) != 0 {
		t.Errorf("deleted keys = %v, want none on a rejected create", cache.deletedKeys)
	}
}
