package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/xtommas/challenge-hetmo/internal/models"
	"github.com/xtommas/challenge-hetmo/internal/repositories"
	"github.com/xtommas/challenge-hetmo/internal/validator"
)

func TestCreateEvent(t *testing.T) {
	// Setup
	e := echo.New()
	e.Validator = validator.NewCustomValidator()

	eventTime, _ := time.Parse(time.RFC3339, "2025-05-01T15:00:00Z")

	// Create a request body
	reqBody := `{
		"title": "Test Event",
		"long_description": "This is a test event",
		"short_description": "Test",
		"date_and_time": "2025-05-01T15:00:00Z",
		"organizer": "Test Org",
		"location": "Test Location",
		"status": "draft"
	}`

	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Set up the expected query with lowercase title, organizer, and location
	mock.ExpectQuery(`INSERT INTO events`).
		WithArgs(
			"test event",
			"This is a test event",
			"Test",
			eventTime,
			"test org",
			"test location",
			"draft",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Create a repository with the mock db
	repo := &repositories.EventRepository{DB: db}

	// Call the handler
	handler := CreateEvent(repo)
	err = handler(c)

	// Assertions
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, rec.Code)

		// Check the response body
		var responseEvent models.Event
		err = json.Unmarshal(rec.Body.Bytes(), &responseEvent)
		assert.NoError(t, err)

		// Check the response values
		assert.Equal(t, int64(1), responseEvent.Id)
		assert.Equal(t, "test event", responseEvent.Title)
		assert.Equal(t, "This is a test event", responseEvent.LongDescription)
		assert.Equal(t, "Test", responseEvent.ShortDescription)
		assert.Equal(t, "test org", responseEvent.Organizer)
		assert.Equal(t, "test location", responseEvent.Location)
		assert.Equal(t, "draft", responseEvent.Status)

		// Parse and check the date
		expectedTime, _ := time.Parse(time.RFC3339, "2025-05-01T15:00:00Z")
		assert.Equal(t, expectedTime, responseEvent.DateAndTime)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCreateEventInvalidInput(t *testing.T) {
	// Set up
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(`{"title": ""}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create a mock validator
	e.Validator = validator.NewCustomValidator()

	// Create a mock repository (we don't need a real DB for this test)
	repo := &repositories.EventRepository{}

	// Call the handler
	handler := CreateEvent(repo)
	err := handler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check that the error message contains validation errors for all required fields
	assert.Contains(t, response["error"], "Field validation for 'Title' failed")
	assert.Contains(t, response["error"], "Field validation for 'LongDescription' failed")
	assert.Contains(t, response["error"], "Field validation for 'ShortDescription' failed")
	assert.Contains(t, response["error"], "Field validation for 'DateAndTime' failed")
	assert.Contains(t, response["error"], "Field validation for 'Organizer' failed")
	assert.Contains(t, response["error"], "Field validation for 'Location' failed")
	assert.Contains(t, response["error"], "Field validation for 'Status' failed")
}

func TestCreateEventDatabaseError(t *testing.T) {
	// Set up
	e := echo.New()
	eventTime := time.Date(2025, 5, 1, 15, 0, 0, 0, time.UTC)
	reqBody := fmt.Sprintf(`{
		"title": "Test Event",
		"long_description": "This is a test event",
		"short_description": "Test",
		"date_and_time": "%s",
		"organizer": "Test Org",
		"location": "Test Location",
		"status": "draft"
	}`, eventTime.Format(time.RFC3339))
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create a mock validator
	e.Validator = validator.NewCustomValidator()

	// Create a mock DB
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`INSERT INTO events`).WillReturnError(fmt.Errorf("database error"))

	// Create a mock repository
	repo := &repositories.EventRepository{DB: db}

	// Call the handler
	handler := CreateEvent(repo)
	err = handler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to create event", response["error"])

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllEvents(t *testing.T) {
	// Setup
	e := echo.New()
	e.Validator = validator.NewCustomValidator()

	eventTime := time.Date(2023, time.June, 1, 0, 0, 0, 0, time.UTC)

	// Test cases
	testCases := []struct {
		name           string
		isAdmin        bool
		queryParams    string
		expectedStatus int
		expectedEvents []models.Event
		expectError    bool
	}{
		{
			name:           "Admin gets all events",
			isAdmin:        true,
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedEvents: []models.Event{
				{Id: 1, Title: "event 1", Status: "draft", DateAndTime: eventTime},
				{Id: 2, Title: "event 2", Status: "published", DateAndTime: eventTime},
			},
			expectError: false,
		},
		{
			name:           "Invalid date format",
			isAdmin:        true,
			queryParams:    "date_start=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedEvents: nil,
			expectError:    true,
		},
		{
			name:           "Non-admin tries to filter by draft status",
			isAdmin:        false,
			queryParams:    "status=draft",
			expectedStatus: http.StatusForbidden,
			expectedEvents: nil,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/events?"+tc.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("is_admin", tc.isAdmin)

			// Mock database
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			// Set up mock expectations for successful test cases
			if tc.expectedStatus == http.StatusOK && !tc.expectError {
				rows := sqlmock.NewRows([]string{"id", "title", "long_description", "short_description", "date_and_time", "organizer", "location", "status"})
				for _, event := range tc.expectedEvents {
					rows.AddRow(event.Id, event.Title, event.LongDescription, event.ShortDescription, event.DateAndTime, event.Organizer, event.Location, event.Status)
				}
				mock.ExpectQuery("SELECT (.+) FROM events").WillReturnRows(rows)
			}

			// Create repository with mock db
			repo := &repositories.EventRepository{DB: db}

			// Call the handler
			handler := GetAllEvents(repo)
			err = handler(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK && !tc.expectError {
				// For successful responses, unmarshal the result and compare events
				var responseEvents []models.Event
				err = json.Unmarshal(rec.Body.Bytes(), &responseEvents)
				assert.NoError(t, err)
				assert.Len(t, responseEvents, len(tc.expectedEvents))

				for i, expectedEvent := range tc.expectedEvents {
					assert.Equal(t, expectedEvent.Id, responseEvents[i].Id)
					assert.Equal(t, expectedEvent.Title, responseEvents[i].Title)
					assert.Equal(t, expectedEvent.Status, responseEvents[i].Status)
					assert.Equal(t, expectedEvent.DateAndTime.Format(time.RFC3339), responseEvents[i].DateAndTime.Format(time.RFC3339))
				}
			} else if tc.expectError {
				// For error responses, unmarshal the error message
				var errorResponse map[string]string
				err = json.Unmarshal(rec.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse, "error")
			}

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetEvent(t *testing.T) {
	// Setup
	e := echo.New()
	e.Validator = validator.NewCustomValidator()

	// Test cases
	testCases := []struct {
		name           string
		isAdmin        bool
		eventID        string
		expectedStatus int
		expectedEvent  *models.Event
	}{
		{
			name:           "Admin gets draft event",
			isAdmin:        true,
			eventID:        "1",
			expectedStatus: http.StatusOK,
			expectedEvent:  &models.Event{Id: 1, Title: "Draft Event", Status: "draft"},
		},
		{
			name:           "Non-admin gets published event",
			isAdmin:        false,
			eventID:        "2",
			expectedStatus: http.StatusOK,
			expectedEvent:  &models.Event{Id: 2, Title: "Published Event", Status: "published"},
		},
		{
			name:           "Non-admin tries to get draft event",
			isAdmin:        false,
			eventID:        "1",
			expectedStatus: http.StatusNotFound,
			expectedEvent:  nil,
		},
		{
			name:           "Event not found",
			isAdmin:        true,
			eventID:        "999",
			expectedStatus: http.StatusNotFound,
			expectedEvent:  nil,
		},
		{
			name:           "Invalid event ID",
			isAdmin:        true,
			eventID:        "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedEvent:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/events/"+tc.eventID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.eventID)
			c.Set("is_admin", tc.isAdmin)

			// Mock database
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			// Set up mock expectations
			if tc.expectedEvent != nil {
				rows := sqlmock.NewRows([]string{"id", "title", "long_description", "short_description", "date_and_time", "organizer", "location", "status"}).
					AddRow(tc.expectedEvent.Id, tc.expectedEvent.Title, tc.expectedEvent.LongDescription, tc.expectedEvent.ShortDescription, tc.expectedEvent.DateAndTime, tc.expectedEvent.Organizer, tc.expectedEvent.Location, tc.expectedEvent.Status)
				mock.ExpectQuery("SELECT (.+) FROM events WHERE id = ?").
					WithArgs(tc.expectedEvent.Id).
					WillReturnRows(rows)
			} else if tc.expectedStatus == http.StatusNotFound {
				mock.ExpectQuery("SELECT (.+) FROM events WHERE id = ?").
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "long_description", "short_description", "date_and_time", "organizer", "location", "status"}))
			}

			// Create repository with mock db
			repo := &repositories.EventRepository{DB: db}

			// Call the handler
			handler := GetEvent(repo)
			err = handler(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedEvent != nil {
				var responseEvent models.Event
				err = json.Unmarshal(rec.Body.Bytes(), &responseEvent)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedEvent.Id, responseEvent.Id)
				assert.Equal(t, tc.expectedEvent.Title, responseEvent.Title)
				assert.Equal(t, tc.expectedEvent.Status, responseEvent.Status)
			} else if tc.expectedStatus != http.StatusOK {
				var errorResponse map[string]string
				err = json.Unmarshal(rec.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Contains(t, errorResponse, "error")
			}

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateEvent(t *testing.T) {
	// Setup
	e := echo.New()
	e.Validator = validator.NewCustomValidator()

	// Test cases
	testCases := []struct {
		name           string
		eventID        string
		reqBody        string
		mockSetup      func(mock sqlmock.Sqlmock)
		expectedStatus int
		expectedEvent  *models.Event
	}{
		{
			name:    "Update event successfully",
			eventID: "1",
			reqBody: `{
				"title": "Updated Event",
				"long_description": "This is an updated event",
				"short_description": "Updated",
				"date_and_time": "2025-06-01T15:00:00Z",
				"organizer": "Updated Org",
				"location": "Updated Location",
				"status": "published"
			}`,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM events WHERE id = ?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "long_description", "short_description", "date_and_time", "organizer", "location", "status"}).
						AddRow(1, "Old Title", "Old Description", "Old Short", time.Now(), "Old Org", "Old Location", "draft"))
				mock.ExpectExec("UPDATE events SET").
					WithArgs(
						"updated event",
						"This is an updated event",
						"Updated",
						time.Date(2025, 6, 1, 15, 0, 0, 0, time.UTC),
						"updated org",
						"updated location",
						"published",
						1,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedStatus: http.StatusOK,
			expectedEvent: &models.Event{
				Id:               1,
				Title:            "updated event",
				LongDescription:  "This is an updated event",
				ShortDescription: "Updated",
				DateAndTime:      time.Date(2025, 6, 1, 15, 0, 0, 0, time.UTC),
				Organizer:        "updated org",
				Location:         "updated location",
				Status:           "published",
			},
		},
		{
			name:    "Event not found",
			eventID: "999",
			reqBody: `{"title": "Non-existent Event"}`,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM events WHERE id = ?").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
			expectedEvent:  nil,
		},
		{
			name:    "Invalid request body",
			eventID: "1",
			reqBody: `{"title": ""}`,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM events WHERE id = ?").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "long_description", "short_description", "date_and_time", "organizer", "location", "status"}).
						AddRow(1, "Old Title", "Old Description", "Old Short", time.Now(), "Old Org", "Old Location", "draft"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedEvent:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodPut, "/events/"+tc.eventID, strings.NewReader(tc.reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.eventID)

			// Mock database
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			// Set up mock expectations
			tc.mockSetup(mock)

			// Create repository with mock db
			repo := &repositories.EventRepository{DB: db}

			// Call the handler
			handler := UpdateEvent(repo)
			err = handler(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedEvent != nil {
				var responseEvent models.Event
				err = json.Unmarshal(rec.Body.Bytes(), &responseEvent)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedEvent.Id, responseEvent.Id)
				assert.Equal(t, tc.expectedEvent.Title, responseEvent.Title)
				assert.Equal(t, tc.expectedEvent.LongDescription, responseEvent.LongDescription)
				assert.Equal(t, tc.expectedEvent.ShortDescription, responseEvent.ShortDescription)
				assert.Equal(t, tc.expectedEvent.DateAndTime.Format(time.RFC3339), responseEvent.DateAndTime.Format(time.RFC3339))
				assert.Equal(t, tc.expectedEvent.Organizer, responseEvent.Organizer)
				assert.Equal(t, tc.expectedEvent.Location, responseEvent.Location)
				assert.Equal(t, tc.expectedEvent.Status, responseEvent.Status)
			}

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDeleteEvent(t *testing.T) {
	// Setup
	e := echo.New()

	// Test cases
	testCases := []struct {
		name           string
		eventID        string
		expectedStatus int
	}{
		{
			name:           "Delete event successfully",
			eventID:        "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Event not found",
			eventID:        "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid event ID",
			eventID:        "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/events/"+tc.eventID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tc.eventID)

			// Mock database
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			// Set up mock expectations
			if tc.expectedStatus == http.StatusOK {
				mock.ExpectExec("DELETE FROM events").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			} else if tc.expectedStatus == http.StatusNotFound {
				mock.ExpectExec("DELETE FROM events").
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			}

			// Create repository with mock db
			repo := &repositories.EventRepository{DB: db}

			// Call the handler
			handler := DeleteEvent(repo)
			err = handler(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Ensure all expectations were met
			if tc.expectedStatus != http.StatusBadRequest {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}
