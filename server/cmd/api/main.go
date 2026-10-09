package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/joho/godotenv"

	usecases "milpa/aplication/use-cases"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/infrastructure/adapters/primary/api"
	"milpa/infrastructure/adapters/primary/api/handler"
	"milpa/infrastructure/adapters/primary/api/middleware"
	"milpa/infrastructure/adapters/primary/api/ws"
	"milpa/infrastructure/adapters/secondary/auth"
	"milpa/infrastructure/adapters/secondary/cache"
	repo "milpa/infrastructure/adapters/secondary/repository"
	"milpa/infrastructure/adapters/secondary/search"
	"milpa/infrastructure/adapters/secondary/storage"
	timepkg "milpa/infrastructure/adapters/secondary/time"
	"milpa/infrastructure/config"
	"milpa/infrastructure/database"
	elasticSsearch "milpa/infrastructure/searchService"
)

// El mantenedor oficial del wiring es chapi, al developer le da pereza la inyección de dependencias :)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env")
	}

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dsn := cfg.DatabaseURL

	elasticSearchClient, err := elasticSsearch.CreateESClient(cfg.ESClient)
	if err != nil {
		log.Fatalf("failed to create elasticsearch client: %v", err)
	}

	pool, err := database.CreatePool(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.CloseConnection(pool)

	log.Println("running database migrations...")
	database.MakeMigrations(context.Background(), dsn)
	log.Println("migrations complete")

	// The search index is bootstrapped on boot for the same reason the SQL
	// migrations are: Elasticsearch infers a mapping from the first document it
	// sees, which silently mis-types every later document, and an index that
	// was never created makes every search fail. Running it on an index that
	// already exists is a no-op and leaves the indexed documents alone.
	log.Printf("bootstrapping elasticsearch index %q...", cfg.ESClient.Index)
	if err := search.EnsureIndex(context.Background(), elasticSearchClient, cfg.ESClient.Index); err != nil {
		log.Fatalf("failed to bootstrap elasticsearch index: %v", err)
	}
	log.Println("elasticsearch index ready")

	jwtSecret := cfg.JWTSecret
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	serverPort := cfg.ServerPort

	hasher := auth.NewBcryptHasher(0)
	jwtProvider := auth.NewJWTProvider(jwtSecret, 24*time.Hour)
	clock := timepkg.NewClock()

	userRepo := repo.NewUserRepository(pool)
	companyRepo := repo.NewCompanyRepository(pool)
	offeringRepo := repo.NewOfferingRepository(pool)
	reviewRepo := repo.NewReviewRepository(pool)
	categoryRepo := repo.NewCategoryRepository(pool)
	unitOfMeasureRepo := repo.NewUnitOfMeasureRepository(pool)
	adminStatsRepo := repo.NewAdminStatsRepository(pool)
	inquiryRepo := repo.NewInquiryRepository(pool)
	liquidationRepo := repo.NewLiquidationRepository(pool)
	reportRepo := repo.NewReportRepository(pool)
	auditLogRepo := repo.NewAuditLogRepository(pool)
	conversationRepo := repo.NewConverationImpl(pool)
	messageRepo := repo.NewMessageRepositoryImpl(pool)
	supplyRequestRepo := repo.NewSupplyRequestRepository(pool)
	supplyOfferRepo := repo.NewSupplyOfferRepository(pool)
	matchRepo := repo.NewMatchRepository(pool)
	transactionRepo := repo.NewTransactionRepository(pool)
	supplierInventoryRepo := repo.NewSupplierInventoryRepository(pool)
	unitOfWork := repo.NewUnitOfWork(pool)

	searchRepo := search.NewElasticSearchImpl(elasticSearchClient, cfg.ESClient.Index)

	cacheClient := cache.NewCacheImpl(cfg.Redis.Addr, cfg.Redis.Password, 0, cfg.Redis.TLS)

	var userUC primary.UserUseCase = usecases.NewUserUseCase(userRepo, hasher, jwtProvider, clock)
	var companyUC primary.CompanyUseCase = usecases.NewCompanyUseCase(companyRepo, userRepo, categoryRepo, clock)
	var reviewUC primary.ReviewUseCase = usecases.NewReviewUseCase(reviewRepo, transactionRepo, matchRepo, supplyOfferRepo, supplyRequestRepo, companyRepo, clock)
	var categoryUC primary.CategoryUseCase = usecases.NewCategoryUseCase(categoryRepo)
	var unitOfMeasureUC primary.UnitOfMeasureUseCase = usecases.NewUnitOfMeasureUseCase(unitOfMeasureRepo)
	var adminStatsUC primary.AdminStatsUseCase = usecases.NewAdminStatsUseCase(adminStatsRepo)
	var inquiryUC primary.InquiryUseCase = usecases.NewInquiryUseCase(inquiryRepo, offeringRepo, clock)
	var searchUC primary.FuzzyUseCase = usecases.NewCachedSearchUseCase(usecases.NewSearchImpl(searchRepo), cacheClient)

	var offeringUC primary.OfferingUseCase = usecases.NewOfferingUseCase(offeringRepo, userRepo, categoryRepo, clock, searchRepo, searchUC.(port.Invalidator))
	var liquidationUC primary.LiquidationUseCase = usecases.NewLiquidationUseCase(liquidationRepo, userRepo, clock)

	var reportUC primary.ReportUseCase = usecases.NewReportUseCase(reportRepo, auditLogRepo, userRepo, offeringRepo, clock)
	var moderationUC primary.ModerationUseCase = usecases.NewModerationUseCase(userRepo, offeringRepo, auditLogRepo, clock)

	var conversationUC primary.ConversationUserUseCase = usecases.NewConversationUseCase(conversationRepo, offeringRepo, userRepo, clock)
	var messageUC primary.MessageUserCase = usecases.NewMessageUseCase(messageRepo, conversationRepo, clock)
	var transactionUC primary.TransactionUseCase = usecases.NewTransactionUseCase(transactionRepo, matchRepo, supplyRequestRepo, supplyOfferRepo, clock, unitOfWork)

	var supplyRequestUC primary.SupplyRequestUseCase = usecases.NewSupplyRequestUseCase(supplyRequestRepo, supplyOfferRepo, matchRepo, clock)
	var supplyOfferUC primary.SupplyOfferUseCase = usecases.NewSupplyOfferUseCase(supplyOfferRepo, supplyRequestRepo, matchRepo, clock)

	var recommendationUC primary.RecommendationUseCase = usecases.NewRecommendationUseCase(supplyOfferRepo, supplyRequestRepo, userRepo, supplierInventoryRepo, matchRepo, usecases.DefaultScoreFactors(reviewRepo))
	var matchUC primary.MatchUseCase = usecases.NewMatchUseCase(supplyRequestRepo, supplyOfferRepo, matchRepo, transactionRepo, recommendationUC, unitOfWork, cacheClient)

	categoryUC = usecases.NewCachedCategoryUseCase(categoryUC, cacheClient)
	companyUC = usecases.NewCachedCompanyUseCase(companyUC, cacheClient)
	offeringUC = usecases.NewCachedOfferingUseCase(offeringUC, cacheClient)
	reviewUC = usecases.NewCachedReviewUseCase(reviewUC, cacheClient)
	inquiryUC = usecases.NewCachedInquiryUseCase(inquiryUC, cacheClient)
	userUC = usecases.NewCachedUserUseCase(userUC, cacheClient)
	conversationUC = usecases.NewCachedConversationUseCase(conversationUC, cacheClient)
	messageUC = usecases.NewCachedMessageUseCase(messageUC, cacheClient, conversationRepo)

	worker := usecases.NewExpiryWorker(offeringRepo, searchRepo, cacheClient, searchUC.(port.Invalidator), clock)

	imageStore := storage.NewLocalImageStore("./uploads")

	userHandler := handler.NewUserHandler(userUC, imageStore)
	companyHandler := handler.NewCompanyHandler(companyUC)
	offeringHandler := handler.NewOfferingHandler(offeringUC, imageStore)
	reviewHandler := handler.NewReviewHandler(reviewUC)
	categoryHandler := handler.NewCategoryHandler(categoryUC)
	unitOfMeasureHandler := handler.NewUnitOfMeasureHandler(unitOfMeasureUC)
	adminStatsHandler := handler.NewAdminStatsHandler(adminStatsUC)
	inquiryHandler := handler.NewInquiryHandler(inquiryUC)
	liquidationHandler := handler.NewLiquidationHandler(liquidationUC)
	imageHandler := handler.NewImageHandler(imageStore)
	searchHandler := handler.NewSearchHandler(searchUC)
	reportHandler := handler.NewReportHandler(reportUC)
	moderationHandler := handler.NewModerationHandler(moderationUC)
	conversationHandler := handler.NewConversationHandler(conversationUC)
	messageHandler := handler.NewMessageHandler(messageUC)
	supplyRequestHandler := handler.NewSupplyRequestHandler(supplyRequestUC)
	supplyOfferHandler := handler.NewSupplyOfferHandler(supplyOfferUC)
	var inventoryUC primary.SupplierInventoryUseCase = usecases.NewSupplierInventoryUseCase(supplierInventoryRepo)
	inventoryHandler := handler.NewSupplierInventoryHandler(inventoryUC)
	matchHandler := handler.NewMatchHandler(matchUC)
	recommendationHandler := handler.NewRecommendationHandler(recommendationUC)
	transactionHandler := handler.NewTransactionHandler(transactionUC)

	hub := ws.NewHub()
	go hub.Run()

	workerCtx, workerCancel := context.WithCancel(context.Background())
	var workerWG sync.WaitGroup
	workerWG.Add(1)
	go func() {
		defer workerWG.Done()
		worker.Run(workerCtx)
	}()

	chatHandler := ws.NewHandler(hub, messageUC, conversationUC)

	authMW := middleware.NewAuthMiddleware(jwtProvider)
	suspensionMW := middleware.NewSuspensionMiddleware(userRepo)

	r := api.NewRouter(userHandler, companyHandler, offeringHandler, reviewHandler, categoryHandler, inquiryHandler, liquidationHandler, authMW, suspensionMW, imageHandler, searchHandler, reportHandler, moderationHandler, conversationHandler, messageHandler, chatHandler, supplyRequestHandler, supplyOfferHandler, inventoryHandler, matchHandler, recommendationHandler, unitOfMeasureHandler, adminStatsHandler)
	api.RegisterTransactionRoutes(r, transactionHandler, authMW, suspensionMW)

	srv := &http.Server{
		Addr:         ":" + serverPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server starting on port %s", serverPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("shutting down server...")
	workerCancel()
	workerWG.Wait()
	if err := srv.Shutdown(shutdown); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
}
