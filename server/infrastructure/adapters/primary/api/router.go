package api

import (
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"milpa/infrastructure/adapters/primary/api/handler"
	"milpa/infrastructure/adapters/primary/api/middleware"
	"milpa/infrastructure/adapters/primary/api/ws"
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
	conversation *handler.ConversationHandler,
	message *handler.MessageHandler,
	chat *ws.Handler,
	supplyRequest *handler.SupplyRequestHandler,
	supplyOffer *handler.SupplyOfferHandler,
	inventory *handler.SupplierInventoryHandler,
	match *handler.MatchHandler,
	recommendation *handler.RecommendationHandler,
) *chi.Mux {
	r := chi.NewRouter()
	auditorMW := middleware.NewAuditorMiddleware()

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
		r.Group(func(r chi.Router) {
			r.Use(chiMiddleware.Timeout(30 * time.Second))

			r.Post("/auth/register", user.Register)
			r.Post("/auth/login", user.Login)

			r.Get("/categories", category.GetAll)

			r.Route("/users", func(r chi.Router) {
				// The read stays open to the marketplace; optional authentication is
				// what lets the owner and an admin receive the private view.
				r.With(authMW.AuthenticateOptional).Get("/{id}", user.GetByID)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/{id}", user.UpdateProfile)
			})

			r.Route("/companies", func(r chi.Router) {
				r.With(authMW.AuthenticateOptional).Get("/{id}", company.GetByID)
				r.With(authMW.AuthenticateOptional).Get("/", company.GetByOwner)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/", company.Create)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/{id}", company.Update)
			})

			r.Route("/offerings", func(r chi.Router) {
				r.Get("/{id}", offering.GetByID)
				r.Get("/", offering.GetByUserID)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/", offering.Create)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/create2/", offering.Create_v2)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/{id}/status", offering.Deactivate)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/{id}/renew", offering.Renew)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/{id}", offering.Update)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/{id}", offering.DeleteOffering)
			})

			r.Route("/reviews", func(r chi.Router) {
				r.Get("/", review.List)
				r.Get("/average", review.Average)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/", review.Create)
			})

			r.Route("/inquiries", func(r chi.Router) {
				r.Get("/company/{company_id}", inquiry.GetByCompany)
				r.Get("/{id}", inquiry.GetByID)
				r.Get("/", inquiry.GetByUser)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/", inquiry.Create)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/{id}", inquiry.Update)
			})

			r.Route("/liquidations", func(r chi.Router) {
				r.Get("/open", liquidation.GetOpen)
				r.Get("/", liquidation.GetBySupplier)
				r.Get("/{id}", liquidation.GetByID)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/", liquidation.Create)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/{id}", liquidation.Update)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Delete("/{id}", liquidation.Delete)
			})

			r.Get("/images/{filename}", image.Get)

			r.Get("/search", search.Search)

			r.Route("/reports", func(r chi.Router) {
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/", report.Create)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/", report.List)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/{id}/action", report.Resolve)
			})

			r.Route("/conversations", func(r chi.Router) {
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/", conversation.List)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/", conversation.Create)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/{id}", conversation.GetByID)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Delete("/{id}", conversation.Delete)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/{id}/messages", message.List)
			})

			r.Route("/messages", func(r chi.Router) {
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/", message.Create)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Delete("/{id}", message.Delete)
			})

			r.Route("/supply-requests", func(r chi.Router) {
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/", supplyRequest.List)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/", supplyRequest.Create)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/available", supplyRequest.ListAvailable)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/{id}", supplyRequest.GetByID)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/{id}", supplyRequest.Update)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/{id}/amounts", supplyRequest.UpdateAmounts)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/{id}/deadlines", supplyRequest.UpdateDeadlines)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/{id}/cancel", supplyRequest.Cancel)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/{id}/expire", supplyRequest.Expire)
			})

			r.Route("/supply-offers", func(r chi.Router) {
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/", supplyOffer.ListBySupplier)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/", supplyOffer.Create)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/requests/{request_id}", supplyOffer.ListByRequest)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/{id}", supplyOffer.GetByID)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/{id}", supplyOffer.Update)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/{id}/withdraw", supplyOffer.Withdraw)
			})

			r.Route("/matches", func(r chi.Router) {
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/like/{offerID}", match.Like)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/pass/{offerID}", match.Pass)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/requests/{requestID}", match.ListByRequest)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/requests/{requestID}/prioritized", match.ListPrioritized)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/{matchID}", match.GetByID)
			})

			r.Route("/inventory", func(r chi.Router) {
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/", inventory.ListBySupplier)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Put("/", inventory.Upsert)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Delete("/{id}", inventory.Delete)
			})

			r.Route("/recommendations", func(r chi.Router) {
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/availability", recommendation.Availability)
			})

			r.Route("/admin", func(r chi.Router) {
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/users/{id}/suspend", moderation.SuspendUser)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/users/{id}/role", moderation.SetUserRole)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Delete("/offerings/{id}", moderation.DeleteOffering)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/audit-logs", moderation.ListAuditLogs)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/categories", category.Create)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/categories/{id}", category.Update)
				r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Patch("/categories/{id}/status", category.SetStatus)
			})
		})

		r.Route("/ws", func(r chi.Router) {
			r.With(authMW.AuthenticateWebSocket, suspensionMW.CheckSuspension).Get("/{conversationID}", chat.WSHandler)
		})
	})

	return r
}

func RegisterTransactionRoutes(
	r chi.Router,
	transaction *handler.TransactionHandler,
	authMW *middleware.AuthMiddleware,
	suspensionMW *middleware.SuspensionMiddleware,
) {
	auditorMW := middleware.NewAuditorMiddleware()

	r.Route("/api/v1/transactions", func(r chi.Router) {
		r.Use(chiMiddleware.Timeout(30 * time.Second))

		r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/matches/{match_id}", transaction.GetByMatch)
		r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Get("/requests/{request_id}", transaction.ListByRequest)
		r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/{transaction_id}/confirm-start", transaction.ConfirmStart)
		r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/{transaction_id}/confirm-delivery", transaction.ConfirmDelivery)
		r.With(authMW.Authenticate, suspensionMW.CheckSuspension, auditorMW.CheckReadOnly).Post("/{transaction_id}/cancel", transaction.Cancel)
	})
}
