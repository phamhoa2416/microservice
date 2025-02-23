package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"math/rand/v2"
	"microservices/model"
	"microservices/repository/order"
	"net/http"
	"strconv"
	"time"
)

type Order struct {
	Repo *order.RedisRepository
}

func (h *Order) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerId uuid.UUID        `json:"customer_id"`
		LineItems  []model.LineItem `json:"line_items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()

	newOrder := model.Order{
		OrderID:    rand.Uint64(),
		CustomerID: body.CustomerId,
		LineItem:   body.LineItems,
		CreatedAt:  &now,
	}

	if err := h.Repo.Insert(r.Context(), newOrder); err != nil {
		fmt.Println("failed to insert order: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(newOrder)
	if err != nil {
		fmt.Println("failed to marshal order: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(res)
	if err != nil {
		fmt.Println("failed to write response: ", err)
	}
}

func (h *Order) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")

	if cursorStr == "" {
		cursorStr = "0"
	}

	const decimal = 10
	const bitSize = 64
	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	const size = 50
	res, err := h.Repo.FindAll(r.Context(), order.FindAllPage{
		Offset: cursor,
		Size:   size,
	})

	if err != nil {
		fmt.Println("failed to find all orders: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var response struct {
		Items []model.Order `json:"items"`
		Next  uint64        `json:"next,omitempty"`
	}

	response.Items = res.Orders
	response.Next = res.Cursor

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("failed to marshal response: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(data); err != nil {
		fmt.Println("failed to write response: ", err)
	}
}

func (h *Order) GetOrderById(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	const base = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	o, err := h.Repo.FindById(r.Context(), orderID)
	if errors.Is(err, order.NotExistErr) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("failed to find order: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(o); err != nil {
		fmt.Println("failed to decode order: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Order) UpdateOrderById(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		JSONResponse(w, http.StatusBadRequest, "Invalid Request Body")
		return
	}

	idParam := chi.URLParam(r, "id")

	const base = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, "invalid order id format")
		return
	}

	orderNeeded, err := h.Repo.FindById(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, order.NotExistErr) {
			JSONResponse(w, http.StatusNotFound, "Order Not Found")
			return
		} else if err != nil {
			fmt.Println("failed to find order: ", err)
			JSONResponse(w, http.StatusBadRequest, "invalid order id format")
			return
		}
	}

	const completedStatus = "completed"
	const shippedStatus = "shipped"

	now := time.Now().UTC()

	switch body.Status {
	case shippedStatus:
		if orderNeeded.ShippedAt != nil {
			JSONResponse(w, http.StatusBadRequest, "Order has already been shipped")
			return
		}
		orderNeeded.ShippedAt = &now
	case completedStatus:
		if orderNeeded.CompletedAt != nil {
			JSONResponse(w, http.StatusBadRequest, "Order has already been completed")
			return
		}

		if orderNeeded.ShippedAt == nil {
			JSONResponse(w, http.StatusBadRequest, "Order must be shipped before completed")
			return
		}
		orderNeeded.CompletedAt = &now
	default:
		JSONResponse(w, http.StatusBadRequest, "Invalid status value. Allowed values: 'shipped', 'completed'")
		return
	}

	err = h.Repo.Update(r.Context(), orderNeeded)
	if err != nil {
		fmt.Println("failed to update order: ", err)
		JSONResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orderNeeded); err != nil {
		fmt.Println("failed to write response: ", err)
		JSONResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}
}

func (h *Order) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	const base = 10
	const bitSize = 64

	orderID, err := strconv.ParseUint(idParam, base, bitSize)
	if err != nil {
		JSONResponse(w, http.StatusBadRequest, "invalid order id format")
		return
	}

	err = h.Repo.DeleteById(r.Context(), orderID)
	if errors.Is(err, order.NotExistErr) {
		JSONResponse(w, http.StatusNotFound, "Order Not Found")
		return
	} else if err != nil {
		JSONResponse(w, http.StatusInternalServerError, "Internal server error")
		return
	}
}

func JSONResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(map[string]string{"error": message})
	if err != nil {
		return
	}
}
