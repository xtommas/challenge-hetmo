package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/xtommas/challenge-hetmo/internal/models"
	"github.com/xtommas/challenge-hetmo/internal/repositories"
)

func CreateEvent(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		event := new(models.Event)
		if err := c.Bind(event); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
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
		date := c.QueryParam("date")
		status := c.QueryParam("status")
		title := c.QueryParam("title")

		eventRepo := repositories.EventRepository{DB: db}
		events, err := eventRepo.GetAll(date, status, title)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get events"})
		}
		return c.JSON(http.StatusOK, events)
	}
}

func GetEvent(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		}
		eventRepo := repositories.EventRepository{DB: db}
		event, err := eventRepo.Get(id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get event"})
		}
		return c.JSON(http.StatusOK, event)
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

func UpdateEvent(db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		}
		eventRepo := repositories.EventRepository{DB: db}
		event, err := eventRepo.Get(id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get event"})
		}

		var input struct {
			Title            *string    `json:"title"`
			LongDescription  *string    `json:"long_description"`
			ShortDescription *string    `json:"short_description"`
			DateAndTime      *time.Time `json:"date_and_time"`
			Organizer        *string    `json:"organizer"`
			Location         *string    `json:"location"`
			Status           *string    `json:"status"`
		}

		if err := c.Bind(&input); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		if input.Title != nil {
			event.Title = *input.Title
		}
		if input.LongDescription != nil {
			event.LongDescription = *input.LongDescription
		}
		if input.ShortDescription != nil {
			event.ShortDescription = *input.ShortDescription
		}
		if input.DateAndTime != nil {
			event.DateAndTime = *input.DateAndTime
		}
		if input.Organizer != nil {
			event.Organizer = *input.Organizer
		}
		if input.Location != nil {
			event.Location = *input.Location
		}
		if input.Status != nil {
			event.Status = *input.Status
		}

		if err := c.Validate(event); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		err = eventRepo.Update(event)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update event"})
		}
		return c.JSON(http.StatusOK, event)
	}
}
