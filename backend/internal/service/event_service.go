// Package service contains business logic that orchestrates repository calls.
// For this scope the services are thin wrappers — the transaction logic lives
// in the repository layer. As business rules grow (e.g. booking windows,
// waitlists, per-user limits), validation would be added here.
package service

import (
	"context"

	"github.com/arunsuryan/event-booking-system/backend/internal/model"
	"github.com/arunsuryan/event-booking-system/backend/internal/repository"
	"github.com/google/uuid"
)

// EventService provides business operations on events.
type EventService struct {
	events repository.EventRepository
	users  repository.UserRepository
}

// NewEventService returns an EventService wired to the given repositories.
func NewEventService(events repository.EventRepository, users repository.UserRepository) *EventService {
	return &EventService{events: events, users: users}
}

// ListEvents returns all events ordered by date.
func (s *EventService) ListEvents(ctx context.Context) ([]model.Event, error) {
	return s.events.ListEvents(ctx)
}

// GetEvent returns a single event by ID.
func (s *EventService) GetEvent(ctx context.Context, id uuid.UUID) (*model.Event, error) {
	return s.events.GetEventByID(ctx, id)
}

// ListUsers returns all pre-seeded users.
func (s *EventService) ListUsers(ctx context.Context) ([]model.User, error) {
	return s.users.ListUsers(ctx)
}
