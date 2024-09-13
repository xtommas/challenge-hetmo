package repositories

import (
	"database/sql"

	"github.com/xtommas/challenge-hetmo/internal/models"
)

type EventRepository struct {
	DB *sql.DB
}

func (e EventRepository) Insert(event *models.Event) error {
	query := `
            INSERT INTO events (title, long_description, short_description, date_and_time, organizer, location, status) 
            VALUES ($1, $2, $3, $4, $5, $6, $7) 
            RETURNING id`
	return e.DB.QueryRow(query,
		event.Title,
		event.LongDescription,
		event.ShortDescription,
		event.DateAndTime,
		event.Organizer,
		event.Location,
		event.Status).Scan(&event.Id)
}
