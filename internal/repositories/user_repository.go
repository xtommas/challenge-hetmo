package repositories

import (
	"database/sql"

	"github.com/xtommas/challenge-hetmo/internal/models"
)

type UserRepository struct {
	DB *sql.DB
}

func (r *UserRepository) Create(user *models.User) error {
	query := `INSERT INTO users (username, password, is_admin) VALUES ($1, $2, $3) RETURNING id`
	return r.DB.QueryRow(query, user.Username, user.Password, user.IsAdmin).Scan(&user.ID)
}

func (r *UserRepository) Get(username string) (*models.User, error) {
	query := `SELECT id, username, password, is_admin FROM users WHERE username = $1`
	user := &models.User{}
	err := r.DB.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password, &user.IsAdmin)
	if err != nil {
		return nil, err
	}
	return user, nil
}
