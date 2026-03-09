package service

import (
	"context"

	"github.com/arunsuryan/event-booking-system/backend/internal/model"
	"github.com/arunsuryan/event-booking-system/backend/internal/repository"
	"github.com/google/uuid"
)

// BookingService provides business operations for booking and cancellation.
type BookingService struct {
	bookings repository.BookingRepository
	audits   repository.AuditRepository
}

// NewBookingService returns a BookingService wired to the given repositories.
func NewBookingService(bookings repository.BookingRepository, audits repository.AuditRepository) *BookingService {
	return &BookingService{bookings: bookings, audits: audits}
}

// BookEvent reserves a spot for a user on an event.
func (s *BookingService) BookEvent(ctx context.Context, eventID, userID uuid.UUID) (*model.Booking, error) {
	return s.bookings.BookEvent(ctx, eventID, userID)
}

// CancelBooking cancels an active booking and returns the spot.
func (s *BookingService) CancelBooking(ctx context.Context, bookingID, userID uuid.UUID) (*model.Booking, error) {
	return s.bookings.CancelBooking(ctx, bookingID, userID)
}

// GetUserBookings returns all bookings for a user, optionally filtered by status.
func (s *BookingService) GetUserBookings(ctx context.Context, userID uuid.UUID, status string) ([]model.Booking, error) {
	return s.bookings.GetUserBookings(ctx, userID, status)
}

// GetAuditLogs returns filtered audit log entries.
func (s *BookingService) GetAuditLogs(ctx context.Context, filters repository.AuditFilters) ([]model.AuditLog, error) {
	return s.audits.GetAuditLogs(ctx, filters)
}
