package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/madcarpet/gophermart/internal/authorization"
	"github.com/madcarpet/gophermart/internal/logger"
	"github.com/madcarpet/gophermart/internal/middlewares"
	"github.com/madcarpet/gophermart/internal/models"
	"github.com/madcarpet/gophermart/internal/storage"
)

type HTTPRouter struct {
	mux           *chi.Mux
	storage       storage.Storage
	authorizer    authorization.Authorizer
	ordersChannel chan<- *models.Order
}

func NewHTTPRouter(s storage.Storage, a authorization.Authorizer, ch chan<- *models.Order) *HTTPRouter {
	r := chi.NewRouter()
	return &HTTPRouter{mux: r, storage: s, authorizer: a, ordersChannel: ch}
}

func (r *HTTPRouter) RouterInit(ctx context.Context) error {
	storage := r.storage
	authorizer := r.authorizer
	ordersChannel := r.ordersChannel
	r.mux.Use(middleware.Logger)
	r.mux.Use(middleware.Compress(5))
	r.mux.Route("/api/user", func(r chi.Router) {
		r.Post("/register", RegisterPostHandler(ctx, storage, authorizer))
		r.Post("/login", LoginPostHandler(ctx, storage, authorizer))

		r.Route("/orders", func(r chi.Router) {
			r.Get("/", middlewares.Authorize(authorizer, OrdersGetHandler(ctx, storage)))
			r.Post("/", middlewares.Authorize(authorizer, OrdersPostHandler(ctx, storage, ordersChannel)))
		})

		r.Route("/balance", func(r chi.Router) {
			r.Get("/", middlewares.Authorize(authorizer, BalanceGetHandler(ctx, storage)))
			r.Post("/withdraw", middlewares.Authorize(authorizer, WithdrawPostHandler(ctx, storage)))
		})

		r.Get("/withdrawals", middlewares.Authorize(authorizer, WithdrawlsGetHandler(ctx, storage)))
	})

	// Set NotFound handler
	r.mux.NotFound(NotFoundHandler())
	return nil
}

func (r *HTTPRouter) StartRouter(ra string) error {
	logger.Log.Info("Http Router starting")
	err := http.ListenAndServe(ra, r.mux)
	if err != nil {
		return err
	}
	return nil
}
