package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/xtommas/challenge-hetmo/internal/models"
	"github.com/xtommas/challenge-hetmo/internal/repositories"
)

func Register(userRepo *repositories.UserRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		var input struct {
			Username string `json:"username" validate:"required,min=3,max=50"`
			Password string `json:"password" validate:"required,min=5"`
		}
		if err := c.Bind(&input); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		if err := c.Validate(input); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		user := &models.User{
			Username: input.Username,
			IsAdmin:  false,
		}
		if err := user.SetPassword(input.Password); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to set password"})
		}

		err := userRepo.Create(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to register user"})
		}
		return c.JSON(http.StatusCreated, user)
	}
}

func Login(userRepo *repositories.UserRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.Bind(&input); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
		}

		user, err := userRepo.Get(input.Username)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "An unexpected error occurred"})
		}

		if !user.CheckPassword(input.Password) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}

		token := jwt.New(jwt.SigningMethodHS256)

		// Set claims (the info that the JWT transmits)
		claims := token.Claims.(jwt.MapClaims)
		claims["user_id"] = user.Id
		claims["username"] = user.Username
		claims["is_admin"] = user.IsAdmin
		claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

		// Generate encoded token
		t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]string{
			"token": t,
		})

	}
}

func PromoteUserToAdmin(userRepo *repositories.UserRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the username from the URL
		username := c.Param("username")

		// Get the user from the repository
		user, err := userRepo.Get(username)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to promote user"})
		}

		// Check if the user is already an admin
		if user.IsAdmin {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "User is already an admin"})
		}

		// Promote the user to admin
		user.IsAdmin = true
		err = userRepo.Update(user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to promote user"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "User promoted to admin successfully"})
	}
}
