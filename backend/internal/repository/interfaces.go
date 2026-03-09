// Package repository defines data-access interfaces and their PostgreSQL implementations.
// Interfaces enable dependency injection so services and handlers can be unit-tested
// with mock implementations that don't require a live database.
package repository

import (
	"context"

	"github.com/arunsuryan/event-booking-system/backend/internal/model"
	"github.com/google/uuid"
)

// UserRepository provides read access to pre-seeded users.
type UserRepository interface {
	ListUsers(ctx context.Context) ([]model.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
}

// EventRepository provides read access to events.
type EventRepository interface {
	ListEvents(ctx context.Context) ([]model.Event, error)
	GetEventByID(ctx context.Context, id uuid.UUID) (*model.Event, error)
}

// BookingRepository handles booking and cancellation with full transaction logic.
type BookingRepository interface {
	BookEvent(ctx context.Context, eventID, userID uuid.UUID) (*model.Booking, error)
	CancelBooking(ctx context.Context, bookingID, userID uuid.UUID) (*model.Booking, error)
	GetUserBookings(ctx context.Context, userID uuid.UUID, status string) ([]model.Booking, error)
}

// AuditFilters holds optional query parameters for filtering audit logs.
// Nil/empty values are ignored in the WHERE clause.
type AuditFilters struct {
	EventID   *uuid.UUID
	UserID    *uuid.UUID
	BookingID *uuid.UUID
	Operation string
	Outcome   string
}

// AuditRepository provides read access to the immutable audit log.
type AuditRepository interface {
	GetAuditLogs(ctx context.Context, filters AuditFilters) ([]model.AuditLog, error)
}
