package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/xtommas/challenge-hetmo/internal/models"
	"github.com/xtommas/challenge-hetmo/internal/repositories"
)

func CreateEvent(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		event := new(models.Event)
		if err := c.Bind(event); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
		}

		if err := c.Validate(event); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		eventRepo := repositories.EventRepository{DB: db}
		err := eventRepo.Insert(event)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create event"})
		}
		return c.JSON(http.StatusCreated, event)
	}
}

func GetAllEvents(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		eventRepo := repositories.EventRepository{DB: db}
		events, err := eventRepo.GetAll()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get events"})
		}
		return c.JSON(http.StatusOK, events)
	}
}

func DeleteEvent(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		}
		eventRepo := repositories.EventRepository{DB: db}
		err = eventRepo.Delete(id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete event"})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "Event deleted successfully"})
	}
}
