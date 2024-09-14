package repositories

import (
	"database/sql"
)

type UserEventRepository struct {
	DB *sql.DB
}
