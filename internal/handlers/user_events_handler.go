package handlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/xtommas/challenge-hetmo/internal/repositories"
)

func SignUpForEvent(userEventRepo *repositories.UserEventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := c.Get("user_id").(int64)
		eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid event ID"})
		}

		err = userEventRepo.CreateSignUp(userID, eventID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "Successfully signed up for the event"})
	}
}

func GetUserEvents(userEventRepo *repositories.UserEventRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, ok := c.Get("user_id").(int64)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid user"})
		}

		// Pagination
		pageParam := c.QueryParam("page")
		limitParam := c.QueryParam("limit")

		filter := c.QueryParam("filter")
		if filter != "upcoming" && filter != "past" && filter != "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid filter option"})
		}

		// Default page and limit
		page := 1
		limit := 10

		var err error

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

		events, err := userEventRepo.GetAll(userID, filter, limit, offset)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get events"})
		}

		total, err := userEventRepo.GetTotalCount(userID, filter)
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
