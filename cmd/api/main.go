package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	router "pocketly/internal/delivery/http"
	"pocketly/internal/delivery/http/middleware"
	"pocketly/internal/infrastructure/config"
	persistence "pocketly/internal/infrastructure/persistence/gorm"
)

/*
shutdownTimeout is how long in-flight requests get to finish after a
shutdown signal before the server is forced to stop. It is sized for
the slowest endpoint, /transactions/scan, which waits on the vision API.
*/
const shutdownTimeout = 30 * time.Second

/*
main loads configuration, opens the database connection, delegates
all dependency wiring to Bootstrap, and starts the HTTP server. It
intentionally contains no construction logic itself — see
bootstrap.go for that.

On SIGINT (Ctrl+C) or SIGTERM (sent by Docker/Kubernetes on deploy),
the server stops accepting new connections, waits up to
shutdownTimeout for in-flight requests to finish, then closes the
database before exiting.
*/
func main() {
	cfg := config.Load()

	db, err := persistence.Connect(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := persistence.Close(db); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	if err := persistence.CheckSchema(db); err != nil {
		log.Fatal(err)
	}

	app, err := Bootstrap(cfg, db)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.Use(middleware.CORS())
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		log.Fatal(err)
	}

	router.SetupRoutes(r, app.AuthHandler, app.TransactionHandler, app.TokenVerifier, cfg.VisionConfig.MaxFileSize)

	/*
		ReadHeaderTimeout stops slow clients from holding connections
		open by trickling headers. Read/write timeouts are left unset
		so large receipt uploads and slow vision API calls are not cut
		off mid-request.
	*/
	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("server listening on %s", srv.Addr)
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server error: %v", err)
			return
		}
	case <-ctx.Done():
		log.Println("shutdown signal received, draining in-flight requests")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		return
	}

	log.Println("server stopped")
}
