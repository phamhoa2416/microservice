package handler

import (
	"fmt"
	"net/http"
)

type Order struct {
}

func (order *Order) CreateOrder(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CreateOrder")
}

func (order *Order) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetAllOrders")
}

func (order *Order) GetOrderById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetOrderById")
}

func (order *Order) UpdateOrder(w http.ResponseWriter, r *http.Request) {

}

func (order *Order) DeleteOrder(w http.ResponseWriter, r *http.Request) {

}
