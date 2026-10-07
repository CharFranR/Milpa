package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
)

type collectingFuzzyRetrival struct {
	documents []*dto.IndexOfferingRequest
	indexErr  error
}

func (f *collectingFuzzyRetrival) Search(ctx context.Context, query *dto.SearchQuery) (*dto.SearchResponse, error) {
	return &dto.SearchResponse{}, nil
}

func (f *collectingFuzzyRetrival) Index(ctx context.Context, p *dto.IndexOfferingRequest) error {
	if f.indexErr != nil {
		return f.indexErr
	}
	f.documents = append(f.documents, p)
	return nil
}

func (f *collectingFuzzyRetrival) Update(ctx context.Context, id string, p *dto.IndexOfferingRequest) error {
	return nil
}

func (f *collectingFuzzyRetrival) Delete(ctx context.Context, id string) error {
	return nil
}

func TestReindexAll_indexes_every_active_offering(t *testing.T) {
	second := *mustOffering()
	second.ID = uuid.New()
	second.Name = "Tomatoes"

	repo := newFakeOfferingRepo()
	repo.findAll = func(ctx context.Context) ([]domain.Offering, error) {
		return []domain.Offering{*mustOffering(), second}, nil
	}
	fuzzy := &collectingFuzzyRetrival{}

	uc := usecases.NewOfferingUseCase(repo, newFakeUserRepo(), newFakeTimer(), fuzzy, &fakeInvalidator{})

	indexed, err := uc.ReindexAll(context.Background())

	if err != nil {
		t.Fatalf("ReindexAll returned error: %v", err)
	}
	if indexed != 2 {
		t.Fatalf("indexed = %d, want 2", indexed)
	}
	if len(fuzzy.documents) != 2 {
		t.Fatalf("documents indexed = %d, want 2", len(fuzzy.documents))
	}
	if fuzzy.documents[0].Name != "Organic Corn" || fuzzy.documents[1].Name != "Tomatoes" {
		t.Fatalf("indexed names = %q, %q", fuzzy.documents[0].Name, fuzzy.documents[1].Name)
	}
	if fuzzy.documents[0].FarmerName == "" {
		t.Fatal("the document should carry the farmer data, like CreateOffering does")
	}
}

func TestReindexAll_fails_when_the_offerings_cannot_be_read(t *testing.T) {
	repo := newFakeOfferingRepo()
	repo.findAll = func(ctx context.Context) ([]domain.Offering, error) {
		return nil, errors.New("db down")
	}
	fuzzy := &collectingFuzzyRetrival{}

	uc := usecases.NewOfferingUseCase(repo, newFakeUserRepo(), newFakeTimer(), fuzzy, &fakeInvalidator{})

	indexed, err := uc.ReindexAll(context.Background())

	if err == nil {
		t.Fatal("ReindexAll should propagate the repository error")
	}
	if indexed != 0 {
		t.Fatalf("indexed = %d, want 0", indexed)
	}
	if len(fuzzy.documents) != 0 {
		t.Fatalf("nothing should be indexed, got %d", len(fuzzy.documents))
	}
}

func TestReindexAll_keeps_going_when_a_document_fails(t *testing.T) {
	repo := newFakeOfferingRepo()
	repo.findAll = func(ctx context.Context) ([]domain.Offering, error) {
		return []domain.Offering{*mustOffering()}, nil
	}
	fuzzy := &collectingFuzzyRetrival{indexErr: errors.New("elasticsearch unreachable")}

	uc := usecases.NewOfferingUseCase(repo, newFakeUserRepo(), newFakeTimer(), fuzzy, &fakeInvalidator{})

	indexed, err := uc.ReindexAll(context.Background())

	if err == nil {
		t.Fatal("ReindexAll should report that the index is not populated")
	}
	if indexed != 0 {
		t.Fatalf("indexed = %d, want 0", indexed)
	}
}