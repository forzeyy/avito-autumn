package app

import (
	"fmt"

	"github.com/forzeyy/avito-autumn/internal/config"
	"github.com/forzeyy/avito-autumn/internal/database"
	"github.com/forzeyy/avito-autumn/internal/routes"
	"github.com/labstack/echo/v4"
)

func Run(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)
	conn, err := database.ConnectDatabase(dsn)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к бд: %v", err)
	}
	defer conn.Close()

	e := echo.New()
	routes.InitRoutes(e, conn)
	e.Logger.Fatal(e.Start(":8080"))

	return nil
}
