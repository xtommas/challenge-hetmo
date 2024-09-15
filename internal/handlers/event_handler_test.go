package handlers

import (
	"encoding/json"
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
	mock.ExpectQuery(`^INSERT INTO events`).
		WithArgs(
			"test event",
			"This is a test event",
			"Test",
			eventTime, // Use the time.Time object here
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

func TestGetAllEvents(t *testing.T) {
}
