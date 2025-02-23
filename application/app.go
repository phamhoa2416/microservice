package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
	server *http.Server
}

func New() *App {
	return &App{
		router: loadRoutes(),
		rdb:    redis.NewClient(&redis.Options{}),
	}
}

func (app *App) Init(ctx context.Context) error {
	if err := app.rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis server: %w", err)
	}

	app.server = &http.Server{
		Addr:    ":3000",
		Handler: app.router,
	}

	errCh := make(chan error, 1)

	go func() {
		err := app.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("failed to start server: %w", err)
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		fmt.Println("Shutting down server...")

		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := app.server.Shutdown(timeoutCtx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
	}

	return nil
}

func (app *App) Shutdown() {
	if err := app.rdb.Close(); err != nil {
		fmt.Println("Failed to close Redis connection:", err)
	}
}
