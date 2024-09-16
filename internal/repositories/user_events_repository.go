package repositories

import (
	"database/sql"
	"errors"
	"time"

	"github.com/xtommas/challenge-hetmo/internal/models"
)

type UserEventRepository struct {
	DB *sql.DB
}

func (r *UserEventRepository) CreateSignUp(userID, eventID int64) error {
	// Ensure the event is published and the date is in the future
	// 'WHERE EXISTS' makes the insert only occur if the conditions are met
	query := `
		INSERT INTO user_events (user_id, event_id)
		SELECT $1, $2
		WHERE EXISTS (
			SELECT 1 FROM events
			WHERE id = $2 AND status = 'published' AND date_and_time > $3
		)`

	result, err := r.DB.Exec(query, userID, eventID, time.Now())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("can't sign up to event")
	}

	return nil
}

func (r *UserEventRepository) GetTotalCount(userID int64, filter string) (int, error) {
	query := `
			SELECT COUNT(*) FROM events e
			JOIN user_events ue ON e.id = ue.event_id
			WHERE ue.user_id = $1
			`
	args := []interface{}{userID}

	// Apply filter
	if filter == "upcoming" {
		query += ` AND e.date_and_time > NOW()`
	} else if filter == "past" {
		query += ` AND e.date_and_time <= NOW()`
	}

	var count int
	if err := r.DB.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (r *UserEventRepository) GetAll(userID int64, filter string, limit, offset int) ([]models.Event, error) {
	query := `SELECT e.id, e.title, e.long_description, e.short_description, e.date_and_time, e.organizer, e.location, e.status
              FROM events e
              JOIN user_events ue ON e.id = ue.event_id
              WHERE ue.user_id = $1`

	if filter == "upcoming" {
		query += ` AND e.date_and_time > NOW()`
	} else if filter == "past" {
		query += ` AND e.date_and_time <= NOW()`
	}

	query += ` LIMIT $2 OFFSET $3`
	args := []interface{}{userID, limit, offset}

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
