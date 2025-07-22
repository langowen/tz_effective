package public

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
	"time"
	"tz_effective/deploy/config"
	_ "tz_effective/internal/ports/http/public/docs"
	mwLogger "tz_effective/internal/ports/http/public/middleware/logger"
	"tz_effective/internal/service"
)

type Server struct {
	Server  *http.Server
	cfg     *config.Config
	Service Service
}

func NewServer(server *http.Server, cfg *config.Config, service2 *service.Service) *Server {
	return &Server{
		Server:  server,
		cfg:     cfg,
		Service: service2,
	}
}

func StartServer(ctx context.Context, service *service.Service, cfg *config.Config) <-chan struct{} {

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mwLogger.New())
	r.Use(middleware.Recoverer)

	serverConfig := &http.Server{
		Addr:         ":" + cfg.HTTPServer.Port,
		Handler:      r,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	server := NewServer(serverConfig, cfg, service)

	doneChan := make(chan struct{})

	go func() {
		if err := server.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Http server error", "error", err)
		}
	}()

	r.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", server.CreateSubscription)
		r.Get("/{id}", server.GetSubscription)
		r.Put("/{id}", server.UpdateSubscription)
		r.Delete("/{id}", server.DeleteSubscription)
		r.Get("/", server.ListSubscriptions)
		r.Get("/cost", server.CalculateTotalCost)
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:"+cfg.HTTPServer.Port+"/swagger/doc.json"), // The url pointing to API definition
	))

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Server.Shutdown(shutdownCtx); err != nil {
			slog.Error("Failed to stop server", "error", err)
		}

		close(doneChan)
	}()

	return doneChan
}

func RespondWithJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to encode response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RespondWithError(w http.ResponseWriter, code int, message string, details ...string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)

	errorText := message
	if len(details) > 0 {
		errorText += "\nDetails: " + details[0]
	}

	if _, err := w.Write([]byte(errorText)); err != nil {
		slog.Error("Failed to write error response", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
