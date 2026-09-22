package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"milpa/aplication/dto"
	usecases "milpa/aplication/use-cases"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/infrastructure/adapters/primary/api"
	"milpa/infrastructure/adapters/primary/api/handler"
	"milpa/infrastructure/adapters/primary/api/middleware"
	"milpa/infrastructure/adapters/secondary/auth"
	"milpa/infrastructure/adapters/secondary/cache"
	repo "milpa/infrastructure/adapters/secondary/repository"
	"milpa/infrastructure/adapters/secondary/search"
	"milpa/infrastructure/adapters/secondary/storage"
	timepkg "milpa/infrastructure/adapters/secondary/time"
	"milpa/infrastructure/database"
	elasticSsearch "milpa/infrastructure/searchService"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("DB_SSLMODE"),
	)

	MaxIdleConnsPerHost, _ := strconv.Atoi(os.Getenv("ESCLIENT_MAXID"))

	ClientData := dto.ESClient{
		Username:            os.Getenv("ESCLIENT_USER"),
		Password:            os.Getenv("ESCLIENT_PASSWORD"),
		Endpoint1:           os.Getenv("ESCLIENT_ENDPOINT1"),
		Endpoint2:           os.Getenv("ESCLIENT_ENDPOINT2"),
		MaxIdleConnsPerHost: MaxIdleConnsPerHost,
	}

	elasticSearchClient, err := elasticSsearch.CreateESClient(ClientData)

	log.Println("main elasticSearchClient error: %w", err)

	pool, err := database.CreatePool(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.CloseConnection(pool)

	log.Println("running database migrations...")
	database.MakeMigrations(context.Background(), dsn)
	log.Println("migrations complete")

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	hasher := auth.NewBcryptHasher(0)
	jwtProvider := auth.NewJWTProvider(jwtSecret, 24*time.Hour)
	clock := timepkg.NewClock()

	userRepo := repo.NewUserRepository(pool)
	companyRepo := repo.NewCompanyRepository(pool)
	offeringRepo := repo.NewOfferingRepository(pool)
	reviewRepo := repo.NewReviewRepository(pool)
	categoryRepo := repo.NewCategoryRepository(pool)
	inquiryRepo := repo.NewInquiryRepository(pool)
	liquidationRepo := repo.NewLiquidationRepository(pool)
	reportRepo := repo.NewReportRepository(pool)
	auditLogRepo := repo.NewAuditLogRepository(pool)

	searchRepo := search.NewElasticSearchImpl(elasticSearchClient, ClientData.Index)

	cacheClient := cache.NewCacheImpl(
		resolveRedisAddr(),
		os.Getenv("REDIS_PASSWORD"),
		0,
	)

	var userUC primary.UserUseCase = usecases.NewUserUseCase(userRepo, hasher, jwtProvider, clock)
	var companyUC primary.CompanyUseCase = usecases.NewCompanyUseCase(companyRepo, userRepo, categoryRepo, clock)
	var reviewUC primary.ReviewUseCase = usecases.NewReviewUseCase(reviewRepo, clock)
	var categoryUC primary.CategoryUseCase = usecases.NewCategoryUseCase(categoryRepo)
	var inquiryUC primary.InquiryUseCase = usecases.NewInquiryUseCase(inquiryRepo, clock)
	var searchUC primary.FuzzyUseCase = usecases.NewCachedSearchUseCase(usecases.NewSearchImpl(searchRepo), cacheClient)

	var offeringUC primary.OfferingUseCase = usecases.NewOfferingUseCase(offeringRepo, userRepo, clock, searchRepo, searchUC.(port.Invalidator))
	var liquidationUC primary.LiquidationUseCase = usecases.NewLiquidationUseCase(liquidationRepo, userRepo, clock)

	var reportUC primary.ReportUseCase = usecases.NewReportUseCase(reportRepo, auditLogRepo, userRepo, offeringRepo, clock)
	var moderationUC primary.ModerationUseCase = usecases.NewModerationUseCase(userRepo, offeringRepo, auditLogRepo, offeringUC, clock)

	categoryUC = usecases.NewCachedCategoryUseCase(categoryUC, cacheClient)
	companyUC = usecases.NewCachedCompanyUseCase(companyUC, cacheClient)
	offeringUC = usecases.NewCachedOfferingUseCase(offeringUC, cacheClient)
	reviewUC = usecases.NewCachedReviewUseCase(reviewUC, cacheClient)
	inquiryUC = usecases.NewCachedInquiryUseCase(inquiryUC, cacheClient)
	userUC = usecases.NewCachedUserUseCase(userUC, cacheClient)

	imageStore := storage.NewLocalImageStore("./uploads")

	userHandler := handler.NewUserHandler(userUC)
	companyHandler := handler.NewCompanyHandler(companyUC)
	offeringHandler := handler.NewOfferingHandler(offeringUC, imageStore)
	reviewHandler := handler.NewReviewHandler(reviewUC)
	categoryHandler := handler.NewCategoryHandler(categoryUC)
	inquiryHandler := handler.NewInquiryHandler(inquiryUC)
	liquidationHandler := handler.NewLiquidationHandler(liquidationUC)
	imageHandler := handler.NewImageHandler(imageStore)
	searchHandler := handler.NewSearchHandler(searchUC)
	reportHandler := handler.NewReportHandler(reportUC)
	moderationHandler := handler.NewModerationHandler(moderationUC)

	authMW := middleware.NewAuthMiddleware(jwtProvider)
	suspensionMW := middleware.NewSuspensionMiddleware(userRepo)

	r := api.NewRouter(userHandler, companyHandler, offeringHandler, reviewHandler, categoryHandler, inquiryHandler, liquidationHandler, authMW, suspensionMW, imageHandler, searchHandler, reportHandler, moderationHandler)

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
	if err := srv.Shutdown(shutdown); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
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
