package api

import (
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"milpa/infrastructure/adapters/primary/api/handler"
	"milpa/infrastructure/adapters/primary/api/middleware"
)

func NewRouter(
	user *handler.UserHandler,
	company *handler.CompanyHandler,
	offering *handler.OfferingHandler,
	review *handler.ReviewHandler,
	category *handler.CategoryHandler,
	inquiry *handler.InquiryHandler,
	authMW *middleware.AuthMiddleware,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RequestID)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", user.Register)
		r.Post("/auth/login", user.Login)

		r.Get("/categories", category.GetAll)

		r.Route("/users", func(r chi.Router) {
			r.Get("/{id}", user.GetByID)
			r.With(authMW.Authenticate).Patch("/{id}", user.UpdateProfile)
		})

		r.Route("/companies", func(r chi.Router) {
			r.Get("/{id}", company.GetByID)
			r.Get("/", company.GetByOwner)
			r.With(authMW.Authenticate).Post("/", company.Create)
			r.With(authMW.Authenticate).Patch("/{id}", company.Update)
		})

		r.Route("/offerings", func(r chi.Router) {
			r.Get("/{id}", offering.GetByID)
			r.Get("/", offering.GetByCompany)
			r.With(authMW.Authenticate).Post("/", offering.Create)
			r.With(authMW.Authenticate).Patch("/{id}", offering.Update)
		})

		r.Route("/reviews", func(r chi.Router) {
			r.Get("/", review.List)
			r.With(authMW.Authenticate).Post("/", review.Create)
		})

		r.Route("/inquiries", func(r chi.Router) {
			r.Get("/{id}", inquiry.GetByID)
			r.Get("/", inquiry.GetByUser)
			r.With(authMW.Authenticate).Post("/", inquiry.Create)
			r.With(authMW.Authenticate).Patch("/{id}", inquiry.Update)
		})
	})

	return r
}
