package application

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"microservices/handler"
	"microservices/repository/order"
	"net/http"
)

func (app *App) loadRoutes() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/orders", app.loadOrders)

	app.router = router
}

func (app *App) loadOrders(router chi.Router) {
	orderHandler := &handler.Order{
		Repo: &order.RedisRepository{
			Client: app.rdb,
		},
	}

	router.Post("/", orderHandler.CreateOrder)
	router.Get("/", orderHandler.GetAllOrders)
	router.Get("/{id}", orderHandler.GetOrderById)
	router.Put("/{id}", orderHandler.UpdateOrder)
	router.Delete("/{id}", orderHandler.DeleteOrder)
}
