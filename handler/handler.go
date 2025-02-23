package handler

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"math/rand/v2"
	"microservices/model"
	"microservices/repository/order"
	"net/http"
	"strconv"
	time2 "time"
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

	time := time2.Now().UTC()

	newOrder := model.Order{
		OrderID:    rand.Uint64(),
		CustomerID: body.CustomerId,
		LineItem:   body.LineItems,
		CreatedAt:  &time,
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
	fmt.Println("GetOrderById")
}

func (h *Order) UpdateOrder(w http.ResponseWriter, r *http.Request) {

}

func (h *Order) DeleteOrder(w http.ResponseWriter, r *http.Request) {

}
