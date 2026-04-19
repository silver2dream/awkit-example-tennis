package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/silver2dream/awkit-example-tennis/backend/internal/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := server.Run(ctx, os.Getenv("ADDR")); err != nil {
		log.Fatal(err)
	}
}
