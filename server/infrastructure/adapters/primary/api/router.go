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
	liquidation *handler.LiquidationHandler,
	authMW *middleware.AuthMiddleware,
	suspensionMW *middleware.SuspensionMiddleware,
	image *handler.ImageHandler,
	search *handler.SearchHandler,
	report *handler.ReportHandler,
	moderation *handler.ModerationHandler,
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
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Patch("/{id}", user.UpdateProfile)
		})

		r.Route("/companies", func(r chi.Router) {
			r.Get("/{id}", company.GetByID)
			r.Get("/", company.GetByOwner)
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Post("/", company.Create)
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Patch("/{id}", company.Update)
		})

		r.Route("/offerings", func(r chi.Router) {
			r.Get("/{id}", offering.GetByID)
			r.Get("/", offering.GetByUserID)
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Post("/", offering.Create)
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Post("/create2/", offering.Create_v2)
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Patch("/{id}", offering.Update)
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Patch("/{id}", offering.DeleteOffering)
		})

		r.Route("/reviews", func(r chi.Router) {
			r.Get("/", review.List)
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Post("/", review.Create)
		})

		r.Route("/inquiries", func(r chi.Router) {
			r.Get("/company/{company_id}", inquiry.GetByCompany)
			r.Get("/{id}", inquiry.GetByID)
			r.Get("/", inquiry.GetByUser)
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Post("/", inquiry.Create)
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Patch("/{id}", inquiry.Update)
		})

		r.Route("/liquidations", func(r chi.Router) {
			r.Get("/open", liquidation.GetOpen)
			r.Get("/", liquidation.GetBySupplier)
			r.Get("/{id}", liquidation.GetByID)
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Post("/", liquidation.Create)
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Patch("/{id}", liquidation.Update)
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Delete("/{id}", liquidation.Delete)
		})

		r.Get("/images/{filename}", image.Get)

		r.Get("/search", search.Search)

		r.Route("/reports", func(r chi.Router) {
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Post("/", report.Create)
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Get("/", report.List)
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Patch("/{id}/action", report.Resolve)
		})

		r.Route("/admin", func(r chi.Router) {
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Patch("/users/{id}/suspend", moderation.SuspendUser)
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Delete("/offerings/{id}", moderation.DeleteOffering)
			r.With(authMW.Authenticate, suspensionMW.CheckSuspension).Get("/audit-logs", moderation.ListAuditLogs)
		})
	})

	return r
}
