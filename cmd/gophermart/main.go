package main

import (
	"log"
	"loyalty-service/internal/app"
	"os"
)

func main() {
	args := os.Args[1:]

	if err := app.Run(args); err != nil {
		log.Fatal(err)
	}

}
