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
	return r.DB.QueryRow(query, user.Username, user.Password, user.IsAdmin).Scan(&user.Id)
}

func (r *UserRepository) Get(username string) (*models.User, error) {
	query := `SELECT id, username, password, is_admin FROM users WHERE username = $1`
	user := &models.User{}
	err := r.DB.QueryRow(query, username).Scan(&user.Id, &user.Username, &user.Password, &user.IsAdmin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	query := `UPDATE users SET is_admin = $1 WHERE id = $2`
	result, err := r.DB.Exec(query, user.IsAdmin, user.Id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
