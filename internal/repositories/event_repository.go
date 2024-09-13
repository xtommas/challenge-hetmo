package repositories

import (
	"database/sql"
	"errors"

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
	err := e.DB.QueryRow(query,
		event.Title,
		event.LongDescription,
		event.ShortDescription,
		event.DateAndTime,
		event.Organizer,
		event.Location,
		event.Status).Scan(&event.Id)
	return err
}

func (e EventRepository) GetAll() ([]models.Event, error) {
	query := `SELECT * FROM events`
	rows, err := e.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get all the data from the rows
	var events []models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.Id,
			&event.Title,
			&event.LongDescription,
			&event.ShortDescription,
			&event.DateAndTime,
			&event.Organizer,
			&event.Location,
			&event.Status,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	// Check if there were errors during the iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (e EventRepository) Delete(id int64) error {
	query := `DELETE FROM events WHERE id = $1`
	result, err := e.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no rows affected")
	}
	return nil
}
