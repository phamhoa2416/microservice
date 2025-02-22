package application

import (
	"context"
	"fmt"
	"net/http"
)

type App struct {
	router http.Handler
}

func New() *App {
	app := &App{
		router: loadRoutes(),
	}
	return app
}

func (app *App) Init(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":3000",
		Handler: app.router,
	}

	err := server.ListenAndServe()

	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	return nil
}
