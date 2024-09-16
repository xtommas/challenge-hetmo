package handlers

import (
	"database/sql"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/xtommas/challenge-hetmo/internal/models"
	"github.com/xtommas/challenge-hetmo/internal/repositories"
)

func CreateEvent(eventRepo *repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		event := new(models.Event)
		if err := c.Bind(event); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		if err := c.Validate(event); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		err := eventRepo.Create(event)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create event"})
		}
		return c.JSON(http.StatusCreated, event)
	}
}

func GetAllEvents(eventRepo *repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		isAdmin := c.Get("is_admin").(bool)

		dateStartStr := c.QueryParam("date_start")
		dateEndStr := c.QueryParam("date_end")
		status := strings.ToLower(c.QueryParam("status"))
		title := strings.ToLower(c.QueryParam("title"))
		// Pagination
		pageParam := c.QueryParam("page")
		limitParam := c.QueryParam("limit")

		// Parse the dates into a time.Time
		var dateStart, dateEnd time.Time
		var err error

		// Parse date_start if provided
		if dateStartStr != "" {
			dateStart, err = time.Parse("2006-01-02", dateStartStr)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid date_start format. Use YYYY-MM-DD."})
			}
		}

		// Parse date_end if provided
		if dateEndStr != "" {
			dateEnd, err = time.Parse("2006-01-02", dateEndStr)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid date_end format. Use YYYY-MM-DD."})
			}
			// Set end time to the end of the date_end day (23:59:59)
			dateEnd = dateEnd.Add(24 * time.Hour).Add(-time.Second)
		}

		// Admins can filter by status, but validate the status parameter
		if isAdmin {
			if status != "" && status != "draft" && status != "published" {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid status."})
			}
		} else {
			// Non-admins can only see published events
			if status != "" && status != "published" {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Access denied"})
			}
			status = "published"
		}

		// Default page and limit
		page := 1
		limit := 10

		if pageParam != "" {
			page, err = strconv.Atoi(pageParam)
			if err != nil || page < 1 {
				page = 1
			}
		}

		if limitParam != "" {
			limit, err = strconv.Atoi(limitParam)
			if err != nil || limit < 1 {
				limit = 10
			}
		}

		// Calculate offset
		offset := (page - 1) * limit

		events, err := eventRepo.GetAll(dateStart, dateEnd, status, title, limit, offset)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get events"})
		}

		total, err := eventRepo.GetTotalCount(status, title, dateStart, dateEnd)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get total count"})
		}

		totalPages := int(math.Ceil(float64(total) / float64(limit)))

		response := map[string]interface{}{
			"events": events,
			"page":   page,
			"limit":  limit,
			"total":  total,
			"pages":  totalPages,
		}

		return c.JSON(http.StatusOK, response)
	}
}

func GetEvent(eventRepo *repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		isAdmin := c.Get("is_admin").(bool)
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		}
		event, err := eventRepo.Get(id)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Event not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get event"})
		}

		if event.Status == "draft" && !isAdmin {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Event not found"})
		}

		return c.JSON(http.StatusOK, event)
	}
}

func DeleteEvent(eventRepo *repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		}
		err = eventRepo.Delete(id)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Event not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete event"})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "Event deleted successfully"})
	}
}

func UpdateEvent(eventRepo *repositories.EventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		}
		event, err := eventRepo.Get(id)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Event not found"})
			}
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
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Event not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update event"})
		}
		return c.JSON(http.StatusOK, event)
	}
}
