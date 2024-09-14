package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xtommas/challenge-hetmo/internal/models"
)

type EventRepository struct {
	DB *sql.DB
}

func (e EventRepository) Insert(event *models.Event) error {
	// Make the title lowercase for case-insensitive filtering
	event.Title = strings.ToLower(event.Title)
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

func (e EventRepository) Update(event *models.Event) error {
	// Make the title lowercase for case-insensitive filtering
	event.Title = strings.ToLower(event.Title)
	query := `
            UPDATE events 
            SET title = $1, long_description = $2, short_description = $3, date_and_time = $4, organizer = $5, location = $6, status = $7 
            WHERE id = $8`
	result, err := e.DB.Exec(query,
		event.Title,
		event.LongDescription,
		event.ShortDescription,
		event.DateAndTime,
		event.Organizer,
		event.Location,
		event.Status,
		event.Id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("record not found")
	}
	return nil
}

func (e EventRepository) Get(id int64) (*models.Event, error) {
	query := `SELECT * FROM events WHERE id = $1`
	row := e.DB.QueryRow(query, id)
	var event models.Event
	err := row.Scan(
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
	return &event, nil
}

func (e EventRepository) GetAll(dateStart, dateEnd time.Time, status string, title string) ([]models.Event, error) {
	// Add a condition that's always true in the query
	// so we can append other conditions based on the
	// query parameters that are provided
	query := `SELECT * FROM events WHERE 1=1`

	// Cheack for query parameters
	args := []interface{}{}
	// Counter for positional placeholder in SQL query ($1, $2, ...)
	argCounter := 1

	// Add filtering conditions if query parameters are provided
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCounter)
		args = append(args, status)
		argCounter++
	}
	if title != "" {
		query += fmt.Sprintf(" AND title LIKE $%d", argCounter)
		args = append(args, "%"+title+"%")
		argCounter++
	}
	if !dateStart.IsZero() {
		query += fmt.Sprintf(" AND date_and_time >= $%d", argCounter)
		args = append(args, dateStart)
		argCounter++
	}
	if !dateEnd.IsZero() {
		query += fmt.Sprintf(" AND date_and_time <= $%d", argCounter)
		args = append(args, dateEnd)
		argCounter++
	}

	rows, err := e.DB.Query(query, args...)
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
		return errors.New("record not found")
	}
	return nil
}
