package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/arunsuryan/event-booking-system/backend/internal/model"
	"github.com/google/uuid"
)

// EventRepo is the PostgreSQL implementation of EventRepository.
type EventRepo struct {
	db *sql.DB
}

// NewEventRepo returns an EventRepo backed by the given database connection.
func NewEventRepo(db *sql.DB) *EventRepo {
	return &EventRepo{db: db}
}

// ListEvents returns all events ordered by event date.
func (r *EventRepo) ListEvents(ctx context.Context) ([]model.Event, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, description, capacity, booked_count,
		        event_date, location, created_at, updated_at
		 FROM events
		 ORDER BY event_date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(
			&e.ID, &e.Title, &e.Description, &e.Capacity, &e.BookedCount,
			&e.EventDate, &e.Location, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// GetEventByID returns a single event by primary key, or ErrEventNotFound.
func (r *EventRepo) GetEventByID(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	var e model.Event
	err := r.db.QueryRowContext(ctx,
		`SELECT id, title, description, capacity, booked_count,
		        event_date, location, created_at, updated_at
		 FROM events
		 WHERE id = $1`, id).
		Scan(&e.ID, &e.Title, &e.Description, &e.Capacity, &e.BookedCount,
			&e.EventDate, &e.Location, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrEventNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}
