package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/xtommas/challenge-hetmo/internal/repositories"
)

func SignUpForEvent(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(int64)
		eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid event ID"})
		}

		userEventRepo := repositories.UserEventRepository{DB: db}
		err = userEventRepo.CreateSignUp(userID, eventID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "Successfully signed up for the event"})
	}
}

func GetUserEvents(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, ok := c.Get("user_id").(int64)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user"})
		}

		filter := c.QueryParam("filter")
		if filter != "upcoming" && filter != "past" && filter != "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid filter option"})
		}

		userEventRepo := repositories.UserEventRepository{DB: db}
		events, err := userEventRepo.GetAll(userID, filter)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get events"})
		}

		return c.JSON(http.StatusOK, events)
	}
}
