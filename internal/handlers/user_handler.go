package handlers

import (
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
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}

		if !user.CheckPassword(input.Password) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}

		token := jwt.New(jwt.SigningMethodHS256)

		// Set claims (the info that the JWT transmits)
		claims := token.Claims.(jwt.MapClaims)
		claims["user_id"] = user.ID
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
