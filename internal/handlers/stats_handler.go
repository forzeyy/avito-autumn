package handlers

import (
	"net/http"

	"github.com/forzeyy/avito-autumn/internal/services"
	"github.com/labstack/echo/v4"
)

type StatsHandler interface {
	GetStats(c echo.Context) error
}

type statsHandler struct {
	statsService services.StatsService
}

func NewStatsHandler(statsService services.StatsService) StatsHandler {
	return &statsHandler{
		statsService: statsService,
	}
}

func (sh *statsHandler) GetStats(c echo.Context) error {
	stats, err := sh.statsService.GetStats(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": "failed to retrieve statistics",
			},
		})
	}

	return c.JSON(http.StatusOK, stats)
}
