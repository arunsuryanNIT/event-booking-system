// Package handler contains HTTP handlers that parse requests, call services,
// and write responses. No business logic or SQL lives here.
package handler

import (
	"net/http"

	"github.com/arunsuryan/event-booking-system/backend/internal/model"
	"github.com/arunsuryan/event-booking-system/backend/internal/response"
	"github.com/arunsuryan/event-booking-system/backend/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// EventHandler handles HTTP requests for event resources.
type EventHandler struct {
	svc *service.EventService
}

// NewEventHandler returns an EventHandler wired to the given service.
func NewEventHandler(svc *service.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

// ListEvents handles GET /api/events — returns all events.
func (h *EventHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.svc.ListEvents(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), "failed to fetch events")
		return
	}
	response.Success(w, http.StatusOK, events, "events retrieved")
}

// GetEvent handles GET /api/events/{id} — returns a single event.
func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid event id", "id must be a valid UUID")
		return
	}

	event, err := h.svc.GetEvent(r.Context(), id)
	if err != nil {
		if err == model.ErrEventNotFound {
			response.Error(w, http.StatusNotFound, err.Error(), "event not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error(), "failed to fetch event")
		return
	}
	response.Success(w, http.StatusOK, event, "event retrieved")
}
