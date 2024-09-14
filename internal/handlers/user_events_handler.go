package handlers

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/xtommas/challenge-hetmo/internal/repositories"
)

func GetUserEvents(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, ok := c.Get("user_id").(int64)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user"})
		}

		filter := c.QueryParam("filter")
		if filter != "upcoming" && filter != "past" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid filter option"})
		}

		userEventRepo := &repositories.UserEventRepository{DB: db}
		events, err := userEventRepo.GetAll(userID, filter)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get events"})
		}

		return c.JSON(http.StatusOK, events)
	}
}
