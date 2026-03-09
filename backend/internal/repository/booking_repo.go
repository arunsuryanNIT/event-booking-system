package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/arunsuryan/event-booking-system/backend/internal/logger"
	"github.com/arunsuryan/event-booking-system/backend/internal/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// BookingRepo is the PostgreSQL implementation of BookingRepository.
// It owns the transaction logic for booking and cancellation.
type BookingRepo struct {
	db  *sql.DB
	log *logger.Logger
}

// NewBookingRepo returns a BookingRepo backed by the given database connection.
func NewBookingRepo(db *sql.DB, log *logger.Logger) *BookingRepo {
	return &BookingRepo{db: db, log: log}
}

// BookEvent atomically reserves a spot for a user on an event.
//
// Transaction flow:
//  1. Atomic conditional UPDATE on events — increments booked_count only if
//     booked_count < capacity. Postgres acquires a row-level lock during the
//     UPDATE, so concurrent requests serialise at the row automatically.
//  2. INSERT booking row — the partial unique index (event_id, user_id WHERE
//     status='active') prevents duplicate active bookings at the DB level.
//  3. INSERT success audit log inside the same transaction.
//
// On failure (sold out or duplicate), the transaction is rolled back and a
// failure audit log is written in a separate connection so it persists.
func (r *BookingRepo) BookEvent(ctx context.Context, eventID, userID uuid.UUID) (*model.Booking, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // no-op if already committed

	// Step 1: Atomic capacity check + increment.
	// If 0 rows affected, the event is either full or doesn't exist.
	result, err := tx.ExecContext(ctx,
		`UPDATE events SET booked_count = booked_count + 1, updated_at = NOW()
		 WHERE id = $1 AND booked_count < capacity`, eventID)
	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback()
		r.logAuditFailure(ctx, "book", eventID, userID, nil, "sold_out")
		return nil, model.ErrSoldOut
	}

	// Step 2: Create the booking row.
	var booking model.Booking
	err = tx.QueryRowContext(ctx,
		`INSERT INTO bookings (event_id, user_id, status)
		 VALUES ($1, $2, 'active')
		 RETURNING id, event_id, user_id, status, created_at, updated_at`,
		eventID, userID).
		Scan(&booking.ID, &booking.EventID, &booking.UserID,
			&booking.Status, &booking.CreatedAt, &booking.UpdatedAt)
	if err != nil {
		tx.Rollback()
		if isUniqueViolation(err) {
			r.logAuditFailure(ctx, "book", eventID, userID, nil, "already_booked")
			return nil, model.ErrAlreadyBooked
		}
		return nil, err
	}

	// Step 3: Success audit log — same transaction, so it's atomic with the booking.
	_, err = tx.ExecContext(ctx,
		`INSERT INTO audit_logs (operation, event_id, user_id, booking_id, outcome)
		 VALUES ('book', $1, $2, $3, 'success')`,
		eventID, userID, booking.ID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	r.log.Info("booking created", "booking_id", booking.ID, "event_id", eventID, "user_id", userID)
	return &booking, nil
}

// CancelBooking atomically cancels an active booking and returns the spot to capacity.
//
// Transaction flow:
//  1. UPDATE booking status to 'cancelled' — the WHERE clause includes user_id
//     to prevent a user from cancelling someone else's booking, and status='active'
//     to prevent double-cancel.
//  2. Decrement booked_count on the event (the CHECK constraint booked_count >= 0
//     is a database-level safety net against underflow).
//  3. INSERT success audit log inside the same transaction.
func (r *BookingRepo) CancelBooking(ctx context.Context, bookingID, userID uuid.UUID) (*model.Booking, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Step 1: Mark as cancelled (atomic — prevents double-cancel).
	var booking model.Booking
	err = tx.QueryRowContext(ctx,
		`UPDATE bookings SET status = 'cancelled', updated_at = NOW()
		 WHERE id = $1 AND user_id = $2 AND status = 'active'
		 RETURNING id, event_id, user_id, status, created_at, updated_at`,
		bookingID, userID).
		Scan(&booking.ID, &booking.EventID, &booking.UserID,
			&booking.Status, &booking.CreatedAt, &booking.UpdatedAt)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			r.logAuditFailure(ctx, "cancel", uuid.Nil, userID, &bookingID, "not_found_or_already_cancelled")
			return nil, model.ErrBookingNotFound
		}
		return nil, err
	}

	// Step 2: Return the spot to capacity.
	_, err = tx.ExecContext(ctx,
		`UPDATE events SET booked_count = booked_count - 1, updated_at = NOW()
		 WHERE id = $1`, booking.EventID)
	if err != nil {
		return nil, err
	}

	// Step 3: Success audit log — same transaction.
	_, err = tx.ExecContext(ctx,
		`INSERT INTO audit_logs (operation, event_id, user_id, booking_id, outcome)
		 VALUES ('cancel', $1, $2, $3, 'success')`,
		booking.EventID, userID, bookingID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	r.log.Info("booking cancelled", "booking_id", bookingID, "event_id", booking.EventID, "user_id", userID)
	return &booking, nil
}

// GetUserBookings returns all bookings for a user, optionally filtered by status.
// Results include the event title via JOIN for frontend display.
func (r *BookingRepo) GetUserBookings(ctx context.Context, userID uuid.UUID, status string) ([]model.Booking, error) {
	query := `SELECT b.id, b.event_id, b.user_id, b.status, b.created_at, b.updated_at,
	                 e.title AS event_title
	          FROM bookings b
	          JOIN events e ON e.id = b.event_id
	          WHERE b.user_id = $1
	            AND ($2::text = '' OR b.status = $2)
	          ORDER BY b.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []model.Booking
	for rows.Next() {
		var b model.Booking
		if err := rows.Scan(
			&b.ID, &b.EventID, &b.UserID, &b.Status,
			&b.CreatedAt, &b.UpdatedAt, &b.EventTitle,
		); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, rows.Err()
}

// logAuditFailure writes a failure audit record using a separate connection
// (not the rolled-back transaction) so the failure is always persisted.
func (r *BookingRepo) logAuditFailure(ctx context.Context, operation string, eventID, userID uuid.UUID, bookingID *uuid.UUID, reason string) {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO audit_logs (operation, event_id, user_id, booking_id, outcome, failure_reason)
		 VALUES ($1, $2, $3, $4, 'failure', $5)`,
		operation, eventID, userID, bookingID, reason)
	if err != nil {
		r.log.Error("failed to write failure audit log", "error", err,
			"operation", operation, "event_id", eventID, "user_id", userID)
	}
}

// isUniqueViolation checks whether the error is a Postgres unique constraint violation (23505).
func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	ok := errors.As(err, &pqErr)
	return ok && pqErr.Code == "23505"
}
