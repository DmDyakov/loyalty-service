package main

import (
	"context"
	"log"
	"loyalty-service/internal/app"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	args := os.Args[1:]

	if err := app.Run(ctx, args); err != nil {
		log.Fatal(err)
	}

}
