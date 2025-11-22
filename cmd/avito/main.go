package main

import (
	"log"

	"github.com/forzeyy/avito-autumn/internal/app"
	"github.com/forzeyy/avito-autumn/internal/config"
)

func main() {
	cfg := config.LoadConfig()

	if err := app.Run(cfg); err != nil {
		log.Fatalf("ошибка при запуске приложения: %v", err)
	}
}
