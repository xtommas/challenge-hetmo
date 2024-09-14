package repositories

import (
	"database/sql"

	"github.com/xtommas/challenge-hetmo/internal/models"
)

type UserEventRepository struct {
	DB *sql.DB
}

func (r *UserEventRepository) GetUserEvents(userID int64, filter string) ([]models.Event, error) {
	var query string
	var args []interface{}

	switch filter {
	case "upcoming":
		query = `
			SELECT e.id, e.title, e.long_description, e.short_description, e.date_and_time, e.organizer, e.location, e.status
			FROM events e
			JOIN user_events ue ON e.id = ue.event_id
			WHERE ue.user_id = $1 AND e.date_and_time > NOW()
		`
		args = append(args, userID)
	case "past":
		query = `
			SELECT e.id, e.title, e.long_description, e.short_description, e.date_and_time, e.organizer, e.location, e.status
			FROM events e
			JOIN user_events ue ON e.id = ue.event_id
			WHERE ue.user_id = $1 AND e.date_and_time <= NOW()
		`
		args = append(args, userID)
	default:
		query = `
			SELECT e.id, e.title, e.long_description, e.short_description, e.date_and_time, e.organizer, e.location, e.status
			FROM events e
			JOIN user_events ue ON e.id = ue.event_id
			WHERE ue.user_id = $1
		`
		args = append(args, userID)
	}

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
