package main

import (
	"log"

	"github.com/chr-hen/mtg-tui/internal/tui"
)

func main() {
	app := tui.NewApp()
	if err := app.Run(); err != nil {
		log.Fatalf("Error running app: %v", err)
	}
}
