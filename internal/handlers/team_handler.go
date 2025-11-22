package handlers

import (
	"net/http"

	"github.com/forzeyy/avito-autumn/internal/models"
	"github.com/forzeyy/avito-autumn/internal/services"
	"github.com/labstack/echo/v4"
)

type TeamHandler interface {
	CreateTeam(c echo.Context) error
	GetTeam(c echo.Context) error
}

type teamHandler struct {
	teamService services.TeamService
}

func NewTeamHandler(teamService services.TeamService) TeamHandler {
	return &teamHandler{
		teamService: teamService,
	}
}

func (th *teamHandler) CreateTeam(c echo.Context) error {
	var team models.Team
	if err := c.Bind(&team); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": map[string]string{
				"code":    "INVALID_INPUT",
				"message": "please check your input",
			},
		})
	}

	err := th.teamService.CreateTeam(c.Request().Context(), &team)
	if err != nil {
		if err.Error() == "TEAM_EXISTS" {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": map[string]string{
					"code":    "TEAM_EXISTS",
					"message": "team_name already exists",
				},
			})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": "internal server error",
			},
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"team": team,
	})
}

func (th *teamHandler) GetTeam(c echo.Context) error {
	teamName := c.QueryParam("team_name")
	if teamName == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": map[string]string{
				"code":    "INVALID_INPUT",
				"message": "team_name is required",
			},
		})
	}

	team, err := th.teamService.GetTeam(c.Request().Context(), teamName)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{
			"error": map[string]string{
				"code":    "NOT_FOUND",
				"message": "team not found",
			},
		})
	}
	return c.JSON(http.StatusOK, team)
}
