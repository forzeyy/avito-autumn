package handlers

import (
	"net/http"

	"github.com/forzeyy/avito-autumn/internal/models"
	"github.com/forzeyy/avito-autumn/internal/services"
	"github.com/labstack/echo/v4"
)

type PRHandler interface {
	CreatePR(c echo.Context) error
	MergePR(c echo.Context) error
	ReassignReviewer(c echo.Context) error
}

type prHandler struct {
	prService services.PRService
}

func NewPRHandler(prService services.PRService) PRHandler {
	return &prHandler{
		prService: prService,
	}
}

func (prh *prHandler) CreatePR(c echo.Context) error {
	var req models.PullRequestShort
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": map[string]string{
				"code":    "INVALID_INPUT",
				"message": "please check your input",
			},
		})
	}

	pr, err := prh.prService.CreatePR(c.Request().Context(), req.ID, req.Name, req.AuthorID)
	if err != nil {
		if err.Error() == "PR_EXISTS" {
			return c.JSON(http.StatusConflict, echo.Map{
				"error": map[string]string{
					"code":    "PR_EXISTS",
					"message": "pull request already exists",
				},
			})
		}
		if err.Error() == "NOT_FOUND" {
			return c.JSON(http.StatusNotFound, echo.Map{
				"code":    "NOT_FOUND",
				"message": "resource not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"code":    "INTERNAL_ERROR",
			"message": "internal server error",
		})
	}
	return c.JSON(http.StatusCreated, echo.Map{
		"pr": pr,
	})
}

func (prh *prHandler) MergePR(c echo.Context) error {
	var req models.PullRequestShort
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": map[string]string{
				"code":    "INVALID_INPUT",
				"message": "please check your input",
			},
		})
	}

	pr, err := prh.prService.MergePR(c.Request().Context(), req.ID)
	if err != nil {
		if err.Error() == "NOT_FOUND" {
			return c.JSON(http.StatusNotFound, echo.Map{
				"code":    "NOT_FOUND",
				"message": "resource not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "internal server error",
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"pr": pr,
	})
}

func (prh *prHandler) ReassignReviewer(c echo.Context) error {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": map[string]string{
				"code":    "INVALID_INPUT",
				"message": "please check your input",
			},
		})
	}

	pr, newID, err := prh.prService.ReassignReviewer(
		c.Request().Context(),
		req.PullRequestID,
		req.OldUserID,
	)
	if err != nil {
		errCode := err.Error()
		var msg string
		switch errCode {
		case "NOT_FOUND":
			msg = "PR or user not found"
		case "PR_MERGED":
			msg = "cannot reassign on merged PR"
		case "NOT_ASSIGNED":
			msg = "reviewer is not assigned to this PR"
		case "NO_CANDIDATE":
			msg = "no active replacement candidate in team"
		default:
			msg = err.Error()
		}
		return c.JSON(http.StatusConflict, echo.Map{
			"error": map[string]string{
				"code":    errCode,
				"message": msg,
			},
		})
	}
	return c.JSON(http.StatusOK, echo.Map{
		"pr":          pr,
		"replaced_by": newID,
	})
}
