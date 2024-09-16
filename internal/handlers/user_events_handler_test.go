package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/xtommas/challenge-hetmo/internal/models"
	"github.com/xtommas/challenge-hetmo/internal/repositories"
)

func TestSignUpForEvent(t *testing.T) {
	// Setup
	e := echo.New()

	// Test cases
	testCases := []struct {
		name            string
		userID          int64
		eventID         string
		expectedStatus  int
		expectedMessage string
		mockBehavior    func(mock sqlmock.Sqlmock)
	}{
		{
			name:            "Successful sign up",
			userID:          1,
			eventID:         "2",
			expectedStatus:  http.StatusOK,
			expectedMessage: "Successfully signed up for the event",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO user_events").
					WithArgs(1, 2, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name:            "Invalid event ID",
			userID:          1,
			eventID:         "invalid",
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Invalid event ID",
			mockBehavior:    func(mock sqlmock.Sqlmock) {},
		},
		{
			name:            "Can't sign up to event",
			userID:          1,
			eventID:         "2",
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "can't sign up to event",
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO user_events").
					WithArgs(1, 2, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/events/"+tc.eventID+"/signup", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.eventID)
			c.Set("user_id", tc.userID)

			// Mock database
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			tc.mockBehavior(mock)

			// Create repository with mock db
			repo := &repositories.UserEventRepository{DB: db}

			// Call the handler
			handler := SignUpForEvent(repo)
			err = handler(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			var response map[string]string
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tc.expectedStatus == http.StatusOK {
				assert.Equal(t, tc.expectedMessage, response["message"])
			} else {
				assert.Equal(t, tc.expectedMessage, response["error"])
			}

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetUserEvents(t *testing.T) {
	// Setup
	e := echo.New()

	// Test cases
	testCases := []struct {
		name           string
		userID         int64
		filter         string
		expectedStatus int
		expectedEvents []models.Event
		mockBehavior   func(mock sqlmock.Sqlmock)
	}{
		{
			name:           "Get all user events",
			userID:         1,
			filter:         "",
			expectedStatus: http.StatusOK,
			expectedEvents: []models.Event{
				{Id: 1, Title: "Event 1", DateAndTime: time.Now().Add(24 * time.Hour)},
				{Id: 2, Title: "Event 2", DateAndTime: time.Now().Add(-24 * time.Hour)},
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "title", "long_description", "short_description", "date_and_time", "organizer", "location", "status"}).
					AddRow(1, "Event 1", "Long desc 1", "Short desc 1", time.Now().Add(24*time.Hour), "Org 1", "Loc 1", "published").
					AddRow(2, "Event 2", "Long desc 2", "Short desc 2", time.Now().Add(-24*time.Hour), "Org 2", "Loc 2", "published")
				mock.ExpectQuery("SELECT e.id, e.title, e.long_description, e.short_description, e.date_and_time, e.organizer, e.location, e.status FROM events e JOIN user_events ue ON e.id = ue.event_id WHERE ue.user_id = ?").
					WithArgs(1).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Get upcoming user events",
			userID:         1,
			filter:         "upcoming",
			expectedStatus: http.StatusOK,
			expectedEvents: []models.Event{
				{Id: 1, Title: "Event 1", DateAndTime: time.Now().Add(24 * time.Hour)},
			},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "title", "long_description", "short_description", "date_and_time", "organizer", "location", "status"}).
					AddRow(1, "Event 1", "Long desc 1", "Short desc 1", time.Now().Add(24*time.Hour), "Org 1", "Loc 1", "published")
				mock.ExpectQuery("SELECT e.id, e.title, e.long_description, e.short_description, e.date_and_time, e.organizer, e.location, e.status FROM events e JOIN user_events ue ON e.id = ue.event_id WHERE ue.user_id = \\$1 AND e.date_and_time > NOW()").
					WithArgs(1).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Invalid filter",
			userID:         1,
			filter:         "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedEvents: nil,
			mockBehavior:   func(mock sqlmock.Sqlmock) {},
		},
		{
			name:           "Database error",
			userID:         1,
			filter:         "",
			expectedStatus: http.StatusInternalServerError,
			expectedEvents: nil,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT e.id, e.title, e.long_description, e.short_description, e.date_and_time, e.organizer, e.location, e.status FROM events e JOIN user_events ue ON e.id = ue.event_id WHERE ue.user_id = ?").
					WithArgs(1).
					WillReturnError(sqlmock.ErrCancelled)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/user/events?filter="+tc.filter, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", tc.userID)

			// Mock database
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			tc.mockBehavior(mock)

			// Create repository with mock db
			repo := &repositories.UserEventRepository{DB: db}

			// Call the handler
			handler := GetUserEvents(repo)
			err = handler(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var responseEvents []models.Event
				err = json.Unmarshal(rec.Body.Bytes(), &responseEvents)
				assert.NoError(t, err)
				assert.Equal(t, len(tc.expectedEvents), len(responseEvents))
				for i, event := range tc.expectedEvents {
					assert.Equal(t, event.Id, responseEvents[i].Id)
					assert.Equal(t, event.Title, responseEvents[i].Title)
					assert.WithinDuration(t, event.DateAndTime, responseEvents[i].DateAndTime, time.Second)
				}
			} else {
				var response map[string]string
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "error")
				assert.NotEmpty(t, response["error"])
			}

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
