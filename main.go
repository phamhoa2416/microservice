package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"microservices/application"
)

func main() {
	app := application.New(application.LoadConfig())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err := app.Init(ctx)

	if err != nil {

		log.Fatalf("failed to start server: %v", err)
	}
}
