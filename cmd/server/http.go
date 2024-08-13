package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/theyudiriski/billing-service/config"
	"github.com/theyudiriski/billing-service/internal/postgres"
	billing "github.com/theyudiriski/billing-service/internal/service"
)

const (
	GracefulShutdownTimeout = 30 * time.Second
	ServerAPITimeout        = 30 * time.Second
)

type Server struct {
	logger billing.Logger
	server *http.Server
}

func NewServer() *Server {
	conf := config.LoadAPI()
	logger := billing.NewLogger()

	db, err := postgres.NewClient(conf.Database)
	if err != nil {
		panic(err)
	}

	loanStore := postgres.NewLoanStore(db)

	loanService := billing.NewLoanService(logger, loanStore)

	router := NewRouter(
		logger,
		db,

		loanService,
	)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", conf.HTTP.Port),
		ReadTimeout:  conf.HTTP.ReadTimeout,
		WriteTimeout: conf.HTTP.WriteTimeout,
		Handler:      router,
	}
	return &Server{
		logger: logger,
		server: server,
	}
}

func NewRouter(
	logger billing.Logger,
	db *postgres.Client,
	loanService billing.LoanService,
) *chi.Mux {
	r := chi.NewRouter()
	h := &routerHandler{
		router: r,

		logger: logger,
		db:     db,

		loanService: loanService,
	}

	h.router.Use(chiMiddleware.Recoverer)
	h.router.Use(chiMiddleware.Timeout(ServerAPITimeout))

	h.router.Mount("/api", h.registerAPIRoutes())

	return h.router
}

type routerHandler struct {
	router *chi.Mux

	logger billing.Logger
	db     *postgres.Client

	loanService billing.LoanService
}

func (s *Server) Run() error {
	s.logger.Info(fmt.Sprintf("api server running on %v with PID %v", s.server.Addr, os.Getpid()))
	return s.server.ListenAndServe()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), GracefulShutdownTimeout)
	defer cancel()

	s.logger.Info("gracefully shutdown HTTP server")
	return s.server.Shutdown(ctx)
}

func (h *routerHandler) registerAPIRoutes() http.Handler {
	r := chi.NewRouter()

	r.Route("/loans", func(r chi.Router) {
		r.Post("/", CreateLoan(h.logger, h.loanService))

		r.Get("/{id}/outstanding", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			GetOutstandingLoan(h.logger, h.loanService, id)(w, r)
		})

		r.Get("/{id}/pending", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			GetPendingLoan(h.logger, h.loanService, id)(w, r)
		})

		r.Get("/{id}/delinquency", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			IsDelinquent(h.logger, h.loanService, id)(w, r)
		})

		r.Post("/pay", PayLoan(h.logger, h.loanService))
	})

	return r
}
