package application

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"microservices/handler"
	"net/http"
)

func loadRoutes() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/orders", loadOrders)

	return router
}

func loadOrders(router chi.Router) {
	orderHandler := &handler.Order{}

	router.Post("/", orderHandler.CreateOrder)
	router.Get("/", orderHandler.GetAllOrders)
	router.Get("/{id}", orderHandler.GetOrderById)
	router.Put("/{id}", orderHandler.UpdateOrder)
	router.Delete("/{id}", orderHandler.DeleteOrder)
}
