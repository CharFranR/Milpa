package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/auth"
	"milpa/infrastructure/adapters/secondary/cache"
	repo "milpa/infrastructure/adapters/secondary/repository"
	"milpa/infrastructure/adapters/secondary/search"
	timepkg "milpa/infrastructure/adapters/secondary/time"
	"milpa/infrastructure/config"
	"milpa/infrastructure/database"
	elasticSsearch "milpa/infrastructure/searchService"
	internalauth "milpa/internal/auth"
)

const demoFarmerEmail = "finca.elroble@milpa.com"

type demoOffering struct {
	name        string
	variety     string
	category    string
	price       float64
	quantity    float64
	latitude    float64
	longitude   float64
	description string
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env")
	}

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	elasticSearchClient, err := elasticSsearch.CreateESClient(cfg.ESClient)
	if err != nil {
		log.Fatalf("failed to create elasticsearch client: %v", err)
	}

	pool, err := database.CreatePool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.CloseConnection(pool)

	database.MakeMigrations(context.Background(), cfg.DatabaseURL)

	if err := search.EnsureIndex(context.Background(), elasticSearchClient, cfg.ESClient.Index); err != nil {
		log.Fatalf("failed to bootstrap elasticsearch index: %v", err)
	}

	hasher := auth.NewBcryptHasher(0)
	jwtProvider := auth.NewJWTProvider(cfg.JWTSecret, 24*time.Hour)
	clock := timepkg.NewClock()

	userRepo := repo.NewUserRepository(pool)
	offeringRepo := repo.NewOfferingRepository(pool)
	categoryRepo := repo.NewCategoryRepository(pool)
	supplierInventoryRepo := repo.NewSupplierInventoryRepository(pool)
	supplyRequestRepo := repo.NewSupplyRequestRepository(pool)
	supplyOfferRepo := repo.NewSupplyOfferRepository(pool)
	matchRepo := repo.NewMatchRepository(pool)
	transactionRepo := repo.NewTransactionRepository(pool)
	liquidationRepo := repo.NewLiquidationRepository(pool)
	reviewRepo := repo.NewReviewRepository(pool)
	unitOfWork := repo.NewUnitOfWork(pool)

	searchRepo := search.NewElasticSearchImpl(elasticSearchClient, cfg.ESClient.Index)
	cacheClient := cache.NewCacheImpl(resolveRedisAddr(), os.Getenv("REDIS_PASSWORD"), 0)

	userUC := usecases.NewUserUseCase(userRepo, hasher, jwtProvider, clock)
	cachedSearch := usecases.NewCachedSearchUseCase(usecases.NewSearchImpl(searchRepo), cacheClient)
	offeringUC := usecases.NewOfferingUseCase(offeringRepo, userRepo, clock, searchRepo, cachedSearch)
	inventoryUC := usecases.NewSupplierInventoryUseCase(supplierInventoryRepo)
	supplyRequestUC := usecases.NewSupplyRequestUseCase(supplyRequestRepo, supplyOfferRepo, matchRepo, clock)
	supplyOfferUC := usecases.NewSupplyOfferUseCase(supplyOfferRepo, supplyRequestRepo, matchRepo, clock)
	recommendationUC := usecases.NewRecommendationUseCase(supplyOfferRepo, supplyRequestRepo, userRepo, supplierInventoryRepo, matchRepo, usecases.DefaultScoreFactors(reviewRepo))
	matchUC := usecases.NewMatchUseCase(supplyRequestRepo, supplyOfferRepo, matchRepo, transactionRepo, recommendationUC, unitOfWork)
	transactionUC := usecases.NewTransactionUseCase(transactionRepo, matchRepo, supplyRequestRepo, supplyOfferRepo, clock, unitOfWork)
	liquidationUC := usecases.NewLiquidationUseCase(liquidationRepo, userRepo, clock)

	farmerID, farmerCreated, err := ensureFarmer(ctx, userRepo, userUC)
	if err != nil {
		log.Fatalf("failed to ensure demo farmer: %v", err)
	}

	unitIDs, err := loadUnitIDs(ctx, pool, "kg")
	if err != nil {
		log.Fatalf("failed to load units of measure: %v", err)
	}

	categoryIDs, err := loadCategoryIDs(ctx, categoryRepo, []string{"Frutales", "Cítricos", "Otros"})
	if err != nil {
		log.Fatalf("failed to load categories: %v", err)
	}

	farmerCtx := internalauth.WithPrincipal(ctx, internalauth.Principal{UserID: farmerID, Role: domain.RoleAgricultor})

	offeringsCreated, offeringsSkipped, err := seedOfferings(farmerCtx, offeringRepo, offeringUC, farmerID, unitIDs["kg"], categoryIDs)
	if err != nil {
		log.Fatalf("failed to seed offerings: %v", err)
	}

	inventoryUpserted, err := seedInventory(farmerCtx, inventoryUC, farmerID)
	if err != nil {
		log.Fatalf("failed to seed inventory: %v", err)
	}

	farmerState := "reused"
	if farmerCreated {
		farmerState = "created"
	}

	wholesale, err := seedWholesale(ctx, wholesaleSeedDeps{
		userRepo:      userRepo,
		userUC:        userUC,
		requestRepo:   supplyRequestRepo,
		requestUC:     supplyRequestUC,
		offerUC:       supplyOfferUC,
		matchUC:       matchUC,
		transactionUC: transactionUC,
		liquidationUC: liquidationUC,
		farmerID:      farmerID,
	})
	if err != nil {
		log.Fatalf("failed to seed wholesale flow: %v", err)
	}

	log.Printf(
		"seed summary: user=%s (%s, id=%s) offerings_created=%d offerings_skipped=%d inventory_upserted=%d",
		demoFarmerEmail, farmerState, farmerID, offeringsCreated, offeringsSkipped, inventoryUpserted,
	)
	log.Printf(
		"wholesale summary: buyer=%s (%s, id=%s) requests_created=%d requests_skipped=%d offers_created=%d deals_completed=%d liquidations_created=%d",
		demoBuyerEmail, wholesale.BuyerState, wholesale.BuyerID, wholesale.RequestsCreated,
		wholesale.RequestsSkipped, wholesale.OffersCreated, wholesale.DealsCompleted, wholesale.LiquidationsCreated,
	)
}

func ensureFarmer(ctx context.Context, userRepo *repo.UserRepositoryImpl, userUC *usecases.UserUseCaseImpl) (uuid.UUID, bool, error) {
	user, err := userRepo.FindByEmail(ctx, demoFarmerEmail)
	if err == nil {
		return user.ID, false, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return uuid.Nil, false, err
	}

	farmerLatitude := 12.1167
	farmerLongitude := -86.1667

	registered, err := userUC.Register(ctx, dto.RegisterUserRequest{
		Email:           demoFarmerEmail,
		FirstName:       "María",
		LastName:        "López",
		Role:            domain.RoleAgricultor,
		Address:         "Km 8 carretera a Masaya",
		Department:      "Masaya",
		Municipality:    "Masate",
		Latitude:        &farmerLatitude,
		Longitude:       &farmerLongitude,
		PhoneNumber:     "+505 8888 0001",
		Password:        "Password123!",
		ConfirmPassword: "Password123!",
	})
	if err != nil {
		return uuid.Nil, false, err
	}

	return registered.ID, true, nil
}

func loadUnitIDs(ctx context.Context, pool *pgxpool.Pool, requiredCode string) (map[string]uuid.UUID, error) {
	rows, err := pool.Query(ctx, "SELECT id, code FROM units_of_measure")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	unitIDs := make(map[string]uuid.UUID)
	for rows.Next() {
		var id uuid.UUID
		var code string
		if err := rows.Scan(&id, &code); err != nil {
			return nil, err
		}
		unitIDs[code] = id
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if _, ok := unitIDs[requiredCode]; !ok {
		return nil, fmt.Errorf("unit of measure %q not found", requiredCode)
	}
	return unitIDs, nil
}

func loadCategoryIDs(ctx context.Context, categoryRepo *repo.CategoryRepositoryImpl, required []string) (map[string]uuid.UUID, error) {
	categories, err := categoryRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	ids := make(map[string]uuid.UUID, len(categories))
	for _, category := range categories {
		ids[category.Name] = category.ID
	}
	for _, name := range required {
		if _, ok := ids[name]; !ok {
			return nil, fmt.Errorf("category %q not found", name)
		}
	}
	return ids, nil
}

func seedOfferings(ctx context.Context, offeringRepo *repo.OfferingRepositoryImpl, offeringUC *usecases.OfferingUseCaseImpl, farmerID uuid.UUID, kgID uuid.UUID, categoryIDs map[string]uuid.UUID) (int, int, error) {
	existing, err := offeringRepo.FindByUserID(ctx, farmerID, false)
	if err != nil {
		return 0, 0, err
	}

	catalogue := demoOfferings()
	if len(existing) > 0 {
		return 0, len(catalogue), nil
	}

	created := 0
	for _, item := range catalogue {
		categoryID := categoryIDs[item.category]
		unitID := kgID
		latitude := item.latitude
		longitude := item.longitude

		req := dto.CreateOfferingRequest{
			UserID:            farmerID,
			Type:              domain.OfferingProduct,
			Name:              item.name,
			Description:       item.description,
			Price:             item.price,
			Variety:           item.variety,
			UnitOfMeasureID:   &unitID,
			QuantityAvailable: item.quantity,
			CategoryID:        &categoryID,
			Latitude:          &latitude,
			Longitude:         &longitude,
		}

		if _, err := offeringUC.CreateOffering(ctx, req); err != nil {
			return created, 0, fmt.Errorf("create offering %q: %w", item.name, err)
		}
		created++
	}
	return created, 0, nil
}

func seedInventory(ctx context.Context, inventoryUC *usecases.SupplierInventoryUseCaseImpl, farmerID uuid.UUID) (int, error) {
	items := []struct {
		name     string
		quantity float64
	}{
		{name: "Tomate cherry", quantity: 120},
		{name: "Aguacate hass", quantity: 200},
		{name: "Café oro", quantity: 60},
		{name: "Miel de abeja", quantity: 90},
	}

	for _, item := range items {
		_, err := inventoryUC.Upsert(ctx, farmerID, dto.UpsertSupplierInventoryRequest{
			ProductName: item.name,
			Quantity:    item.quantity,
			AmountUnit:  domain.Kg,
		})
		if err != nil {
			return 0, fmt.Errorf("upsert inventory %q: %w", item.name, err)
		}
	}
	return len(items), nil
}

func demoOfferings() []demoOffering {
	return []demoOffering{
		{name: "Aguacate hass", variety: "Hass", category: "Frutales", price: 80, quantity: 200, latitude: 11.974, longitude: -86.094, description: "Aguacate hass de temporada, cosechado en Masaya."},
		{name: "Mango criollo", variety: "Criollo", category: "Frutales", price: 35, quantity: 300, latitude: 12.136, longitude: -86.251, description: "Mango criollo dulce para consumo fresco."},
		{name: "Naranja valencia", variety: "Valencia", category: "Cítricos", price: 25, quantity: 400, latitude: 11.93, longitude: -85.95, description: "Naranja valencia jugosa, ideal para jugo."},
		{name: "Mandarina criolla", variety: "Criolla", category: "Cítricos", price: 30, quantity: 250, latitude: 11.974, longitude: -86.094, description: "Mandarina criolla de cáscara fina."},
		{name: "Limón tahití", variety: "Tahití", category: "Cítricos", price: 40, quantity: 180, latitude: 11.93, longitude: -85.95, description: "Limón tahití fresco."},
		{name: "Café oro", variety: "Oro", category: "Otros", price: 250, quantity: 60, latitude: 13.09, longitude: -86.0, description: "Café oro de altura, tueste medio."},
		{name: "Miel de abeja", variety: "Multifloral", category: "Otros", price: 180, quantity: 90, latitude: 11.974, longitude: -86.094, description: "Miel multifloral pura."},
		{name: "Tomate cherry", variety: "Cherry", category: "Otros", price: 55, quantity: 120, latitude: 11.974, longitude: -86.094, description: "Tomate cherry dulce para ensaladas."},
	}
}

func resolveRedisAddr() string {
	if url := os.Getenv("REDIS_URL"); url != "" {
		host := strings.TrimPrefix(url, "redis://")
		if idx := strings.Index(host, "@"); idx != -1 {
			host = host[idx+1:]
		}
		if idx := strings.Index(host, "/"); idx != -1 {
			host = host[:idx]
		}
		return host
	}
	return os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT")
}
