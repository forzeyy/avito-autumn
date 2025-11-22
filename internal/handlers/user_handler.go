package handlers

import (
	"net/http"

	"github.com/forzeyy/avito-autumn/internal/services"
	"github.com/labstack/echo/v4"
)

type UserHandler interface {
	SetUserActive(c echo.Context) error
	GetPRsByReviewer(c echo.Context) error
}

type userHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) UserHandler {
	return &userHandler{
		userService: userService,
	}
}

func (uh *userHandler) SetUserActive(c echo.Context) error {
	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": map[string]string{
				"code":    "INVALID_INPUT",
				"message": "please check your input",
			},
		})
	}

	user, err := uh.userService.SetUserActive(c.Request().Context(), req.UserID, req.IsActive)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			return c.JSON(http.StatusNotFound, echo.Map{
				"error": map[string]string{
					"code":    "NOT_FOUND",
					"message": "user not found",
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

	return c.JSON(http.StatusOK, echo.Map{"user": user})
}

func (uh *userHandler) GetPRsByReviewer(c echo.Context) error {
	userID := c.QueryParam("user_id")

	prs, err := uh.userService.GetPRsByReviewer(c.Request().Context(), userID)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			return c.JSON(http.StatusNotFound, echo.Map{
				"error": map[string]string{
					"code":    "NOT_FOUND",
					"message": "user not found",
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
		"user_id":       userID,
		"pull_requests": prs,
	})
}
