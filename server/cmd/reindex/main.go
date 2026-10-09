package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"

	usecases "milpa/aplication/use-cases"
	"milpa/infrastructure/adapters/secondary/cache"
	repo "milpa/infrastructure/adapters/secondary/repository"
	"milpa/infrastructure/adapters/secondary/search"
	timepkg "milpa/infrastructure/adapters/secondary/time"
	"milpa/infrastructure/config"
	"milpa/infrastructure/database"
	elasticSsearch "milpa/infrastructure/searchService"
)

// reindex repuebla el indice de busqueda de Elasticsearch con los offerings
// que ya estan en Postgres. Solo CreateOffering y UpdateOffering escriben en el
// indice, asi que un indice recreado (o un volumen de ES que arranca vacio)
// deja el catalogo entero invisible para /search sin otra via que editar cada
// producto a mano.
//
// Uso: ./reindex   (docker compose run --rm api ./reindex)
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system env")
	}

	cfg := config.Load()

	// Sin timeout corto: con muchas ofertas la indexacion lleva su tiempo.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
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

	log.Printf("bootstrapping elasticsearch index %q...", cfg.ESClient.Index)
	if err := search.EnsureIndex(ctx, elasticSearchClient, cfg.ESClient.Index); err != nil {
		log.Fatalf("failed to bootstrap elasticsearch index: %v", err)
	}

	// Sin decoradores de cache en las lecturas: esto es un one-shot. La
	// invalidacion si hace falta al final, porque /search se cachea en Redis y
	// un resultado vacio guardado seguiria sirviendo productos invisibles
	// hasta que expirara el TTL.
	cacheClient := cache.NewCacheImpl(cfg.Redis.Addr, cfg.Redis.Password, 0, cfg.Redis.TLS)

	offeringRepo := repo.NewOfferingRepository(pool)
	userRepo := repo.NewUserRepository(pool)
	categoryRepo := repo.NewCategoryRepository(pool)
	fuzzyRetrieval := search.NewElasticSearchImpl(elasticSearchClient, cfg.ESClient.Index)
	cachedSearch := usecases.NewCachedSearchUseCase(usecases.NewSearchImpl(fuzzyRetrieval), cacheClient)

	uc := usecases.NewOfferingUseCase(
		offeringRepo,
		userRepo,
		categoryRepo,
		timepkg.NewClock(),
		fuzzyRetrieval,
		cachedSearch,
	)

	log.Println("reindexing offerings...")
	indexed, err := uc.ReindexAll(ctx)
	if err != nil {
		log.Fatalf("reindex incomplete: %v", err)
	}

	log.Printf("reindexed %d offerings into %q", indexed, cfg.ESClient.Index)

	if err := cachedSearch.InvalidateAll(ctx); err != nil {
		log.Printf("warning: could not clear the search cache, results may stay stale until the ttl expires: %v", err)
	}
}

