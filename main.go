package main

import (
	"context"
	"log"
	"microservices/application"
)

func main() {
	app := application.New()

	err := app.Init(context.TODO())

	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
