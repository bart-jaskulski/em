package main

import (
	"log"

	"github.com/bart-jaskulski/em/internal/ui"
)

func main() {
	if err := ui.Start(); err != nil {
		log.Fatal(err)
	}
}
