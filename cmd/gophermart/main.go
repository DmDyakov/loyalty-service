package main

import (
	"loyalty-service/internal/app"
	"os"
)

func main() {
	args := os.Args[1:]

	app.Run(args)
}
