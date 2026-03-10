package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/arunsuryan/event-booking-system/backend/internal/model"
	"github.com/arunsuryan/event-booking-system/backend/internal/response"
	"github.com/arunsuryan/event-booking-system/backend/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// BookingHandler handles HTTP requests for booking and cancellation.
type BookingHandler struct {
	svc *service.BookingService
}

// NewBookingHandler returns a BookingHandler wired to the given service.
func NewBookingHandler(svc *service.BookingService) *BookingHandler {
	return &BookingHandler{svc: svc}
}

// bookRequest is the JSON body for booking and cancellation endpoints.
type bookRequest struct {
	UserID string `json:"user_id"`
}

// BookEvent handles POST /api/events/{id}/book.
// Expects JSON body: {"user_id": "uuid"}.
func (h *BookingHandler) BookEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid event id", "id must be a valid UUID")
		return
	}

	var req bookRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body", "expected JSON with user_id")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user_id", "user_id must be a valid UUID")
		return
	}

	booking, err := h.svc.BookEvent(r.Context(), eventID, userID)
	if err != nil {
		switch {
		case errors.Is(err, model.ErrSoldOut):
			response.Error(w, http.StatusConflict, err.Error(), "could not complete booking")
		case errors.Is(err, model.ErrAlreadyBooked):
			response.Error(w, http.StatusConflict, err.Error(), "could not complete booking")
		default:
			response.Error(w, http.StatusInternalServerError, err.Error(), "could not complete booking")
		}
		return
	}

	response.Success(w, http.StatusCreated, booking, "booking created successfully")
}

// CancelBooking handles POST /api/bookings/{id}/cancel.
// Expects JSON body: {"user_id": "uuid"}.
func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	bookingID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid booking id", "id must be a valid UUID")
		return
	}

	var req bookRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body", "expected JSON with user_id")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user_id", "user_id must be a valid UUID")
		return
	}

	booking, err := h.svc.CancelBooking(r.Context(), bookingID, userID)
	if err != nil {
		if errors.Is(err, model.ErrBookingNotFound) {
			response.Error(w, http.StatusNotFound, err.Error(), "could not cancel booking")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error(), "could not cancel booking")
		return
	}

	response.Success(w, http.StatusOK, booking, "booking cancelled successfully")
}

// GetUserBookings handles GET /api/users/{id}/bookings.
// Optional query param: ?status=active
func (h *BookingHandler) GetUserBookings(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id", "id must be a valid UUID")
		return
	}

	status := r.URL.Query().Get("status")

	bookings, err := h.svc.GetUserBookings(r.Context(), userID, status)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), "failed to fetch bookings")
		return
	}

	response.Success(w, http.StatusOK, bookings, "bookings retrieved")
}
