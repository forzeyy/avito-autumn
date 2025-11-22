package routes

import (
	"github.com/forzeyy/avito-autumn/internal/database"
	"github.com/forzeyy/avito-autumn/internal/handlers"
	"github.com/forzeyy/avito-autumn/internal/repos"
	"github.com/forzeyy/avito-autumn/internal/services"
	"github.com/labstack/echo/v4"
)

func InitRoutes(e *echo.Echo, db *database.DB) {
	userRepo := repos.NewUserRepo(db)
	prRepo := repos.NewPRRepo(db)
	teamRepo := repos.NewTeamRepo(db)

	userService := services.NewUserService(userRepo, prRepo)
	prService := services.NewPRService(prRepo, userRepo)
	teamService := services.NewTeamService(teamRepo, userRepo)
	statsService := services.NewStatsService(prRepo, userRepo)

	userHandler := handlers.NewUserHandler(userService)
	prHandler := handlers.NewPRHandler(prService)
	teamHandler := handlers.NewTeamHandler(teamService)
	statsHandler := handlers.NewStatsHandler(statsService)

	// users
	e.POST("/users/setIsActive", userHandler.SetUserActive)
	e.GET("/users/getReview", userHandler.GetPRsByReviewer)

	// pull requests
	e.POST("/pullRequest/create", prHandler.CreatePR)
	e.POST("/pullRequest/merge", prHandler.MergePR)
	e.POST("/pullRequest/reassign", prHandler.ReassignReviewer)

	// teams
	e.POST("/team/add", teamHandler.CreateTeam)
	e.GET("/team/get", teamHandler.GetTeam)

	// stats
	e.GET("/stats", statsHandler.GetStats)
}
