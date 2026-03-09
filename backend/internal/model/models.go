// Package model defines the domain types and sentinel errors shared across layers.
package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// User represents a pre-seeded application user.
type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Event represents a bookable event with a fixed capacity.
type Event struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Capacity    int       `json:"capacity"`
	BookedCount int       `json:"booked_count"`
	EventDate   time.Time `json:"event_date"`
	Location    string    `json:"location"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Booking represents a user's reservation for an event.
// EventTitle is populated via JOIN in some queries for frontend convenience.
type Booking struct {
	ID         uuid.UUID `json:"id"`
	EventID    uuid.UUID `json:"event_id"`
	UserID     uuid.UUID `json:"user_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	EventTitle string    `json:"event_title,omitempty"`
}

// AuditLog records every booking-changing operation for traceability.
// BookingID and FailureReason are nullable — pointers encode SQL NULL cleanly.
// EventTitle and UserName are populated via JOINs for display.
type AuditLog struct {
	ID            uuid.UUID  `json:"id"`
	Operation     string     `json:"operation"`
	EventID       uuid.UUID  `json:"event_id"`
	UserID        uuid.UUID  `json:"user_id"`
	BookingID     *uuid.UUID `json:"booking_id"`
	Outcome       string     `json:"outcome"`
	FailureReason *string    `json:"failure_reason"`
	CreatedAt     time.Time  `json:"created_at"`
	EventTitle    string     `json:"event_title,omitempty"`
	UserName      string     `json:"user_name,omitempty"`
}

// Sentinel errors returned by the repository layer.
// Handlers map these to appropriate HTTP status codes.
var (
	ErrSoldOut         = errors.New("event is sold out")
	ErrAlreadyBooked   = errors.New("user already has an active booking for this event")
	ErrBookingNotFound = errors.New("booking not found or already cancelled")
	ErrEventNotFound   = errors.New("event not found")
	ErrUserNotFound    = errors.New("user not found")
)
