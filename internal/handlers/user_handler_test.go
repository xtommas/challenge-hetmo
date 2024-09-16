package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/xtommas/challenge-hetmo/internal/repositories"
	"github.com/xtommas/challenge-hetmo/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	// Setup
	e := echo.New()
	e.Validator = validator.NewCustomValidator()

	// Test cases
	testCases := []struct {
		name           string
		reqBody        string
		expectedStatus int
		mockBehavior   func(mock sqlmock.Sqlmock)
	}{
		{
			name: "Successful registration",
			reqBody: `{
				"username": "newuser",
				"password": "password123"
			}`,
			expectedStatus: http.StatusCreated,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("newuser", sqlmock.AnyArg(), false).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
		},
		{
			name: "Invalid input - short username",
			reqBody: `{
				"username": "nu",
				"password": "password123"
			}`,
			expectedStatus: http.StatusBadRequest,
			mockBehavior:   func(mock sqlmock.Sqlmock) {},
		},
		{
			name: "Invalid input - short password",
			reqBody: `{
				"username": "newuser",
				"password": "pass"
			}`,
			expectedStatus: http.StatusBadRequest,
			mockBehavior:   func(mock sqlmock.Sqlmock) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tc.reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Mock database
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			tc.mockBehavior(mock)

			// Create repository with mock db
			repo := &repositories.UserRepository{DB: db}

			// Call the handler
			handler := Register(repo)
			err = handler(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLogin(t *testing.T) {
	// Setup
	e := echo.New()

	// Create a hashed password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	// Test cases
	testCases := []struct {
		name           string
		reqBody        string
		expectedStatus int
		mockBehavior   func(mock sqlmock.Sqlmock)
	}{
		{
			name: "Successful login",
			reqBody: `{
				"username": "existinguser",
				"password": "correctpassword"
			}`,
			expectedStatus: http.StatusOK,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "password", "is_admin"}).
					AddRow(1, "existinguser", string(hashedPassword), false)
				mock.ExpectQuery("SELECT (.+) FROM users WHERE username = ?").
					WithArgs("existinguser").
					WillReturnRows(rows)
			},
		},
		{
			name: "Invalid credentials",
			reqBody: `{
				"username": "nonexistentuser",
				"password": "wrongpassword"
			}`,
			expectedStatus: http.StatusUnauthorized,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE username = ?").
					WithArgs("nonexistentuser").
					WillReturnError(sql.ErrNoRows)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tc.reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Mock database
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			tc.mockBehavior(mock)

			// Create repository with mock db
			repo := &repositories.UserRepository{DB: db}

			// Call the handler
			handler := Login(repo)
			err = handler(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var response map[string]string
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "token")
			}

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPromoteUserToAdmin(t *testing.T) {
	// Setup
	e := echo.New()

	// Test cases
	testCases := []struct {
		name           string
		username       string
		expectedStatus int
		mockBehavior   func(mock sqlmock.Sqlmock)
	}{
		{
			name:           "Successfully promote user to admin",
			username:       "regularuser",
			expectedStatus: http.StatusOK,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "password", "is_admin"}).
					AddRow(1, "regularuser", "hashedpassword", false)
				mock.ExpectQuery("SELECT (.+) FROM users WHERE username = ?").
					WithArgs("regularuser").
					WillReturnRows(rows)
				mock.ExpectExec("UPDATE users SET").
					WithArgs(true, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name:           "User not found",
			username:       "nonexistentuser",
			expectedStatus: http.StatusNotFound,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM users WHERE username = ?").
					WithArgs("nonexistentuser").
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "User already an admin",
			username:       "adminuser",
			expectedStatus: http.StatusBadRequest,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "password", "is_admin"}).
					AddRow(2, "adminuser", "hashedpassword", true)
				mock.ExpectQuery("SELECT (.+) FROM users WHERE username = ?").
					WithArgs("adminuser").
					WillReturnRows(rows)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/admin/users/"+tc.username+"/promote", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("username")
			c.SetParamValues(tc.username)

			// Mock database
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			tc.mockBehavior(mock)

			// Create repository with mock db
			repo := &repositories.UserRepository{DB: db}

			// Call the handler
			handler := PromoteUserToAdmin(repo)
			err = handler(c)

			// Assertions
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, rec.Code)

			var response map[string]string
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tc.expectedStatus == http.StatusOK {
				assert.Equal(t, "User promoted to admin successfully", response["message"])
			} else {
				assert.Contains(t, response, "error")
			}

			// Ensure all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
