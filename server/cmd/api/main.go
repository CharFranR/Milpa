package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"

	usecases "milpa/aplication/use-cases"
	"milpa/domain/port/primary"
	"milpa/infrastructure/adapters/primary/api"
	"milpa/infrastructure/adapters/primary/api/handler"
	"milpa/infrastructure/adapters/primary/api/middleware"
	"milpa/infrastructure/adapters/secondary/auth"
	"milpa/infrastructure/adapters/secondary/cache"
	repo "milpa/infrastructure/adapters/secondary/repository"
	timepkg "milpa/infrastructure/adapters/secondary/time"
	"milpa/infrastructure/database"
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

	pool, err := database.CreatePool(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.CloseConnection(pool)

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

	cacheClient := cache.NewCacheImpl(
		os.Getenv("REDIS_HOST")+":"+os.Getenv("REDIS_PORT"),
		os.Getenv("REDIS_PASSWORD"),
		0,
	)

	var userUC primary.UserUseCase = usecases.NewUserUseCase(userRepo, hasher, jwtProvider, clock)
	var companyUC primary.CompanyUseCase = usecases.NewCompanyUseCase(companyRepo, userRepo, categoryRepo, clock)
	var offeringUC primary.OfferingUseCase = usecases.NewOfferingUseCase(offeringRepo, companyRepo, clock)
	var reviewUC primary.ReviewUseCase = usecases.NewReviewUseCase(reviewRepo, clock)
	var categoryUC primary.CategoryUseCase = usecases.NewCategoryUseCase(categoryRepo)
	var inquiryUC primary.InquiryUseCase = usecases.NewInquiryUseCase(inquiryRepo, clock)

	categoryUC = usecases.NewCachedCategoryUseCase(categoryUC, cacheClient)
	companyUC = usecases.NewCachedCompanyUseCase(companyUC, cacheClient)
	offeringUC = usecases.NewCachedOfferingUseCase(offeringUC, cacheClient)
	reviewUC = usecases.NewCachedReviewUseCase(reviewUC, cacheClient)
	inquiryUC = usecases.NewCachedInquiryUseCase(inquiryUC, cacheClient)
	userUC = usecases.NewCachedUserUseCase(userUC, cacheClient)

	userHandler := handler.NewUserHandler(userUC)
	companyHandler := handler.NewCompanyHandler(companyUC)
	offeringHandler := handler.NewOfferingHandler(offeringUC)
	reviewHandler := handler.NewReviewHandler(reviewUC)
	categoryHandler := handler.NewCategoryHandler(categoryUC)
	inquiryHandler := handler.NewInquiryHandler(inquiryUC)

	authMW := middleware.NewAuthMiddleware(jwtProvider)

	r := api.NewRouter(userHandler, companyHandler, offeringHandler, reviewHandler, categoryHandler, inquiryHandler, authMW)

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
