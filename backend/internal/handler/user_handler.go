package handler

import (
	"net/http"

	"github.com/arunsuryan/event-booking-system/backend/internal/response"
	"github.com/arunsuryan/event-booking-system/backend/internal/service"
)

// UserHandler handles HTTP requests for user resources.
type UserHandler struct {
	svc *service.EventService
}

// NewUserHandler returns a UserHandler wired to the given service.
func NewUserHandler(svc *service.EventService) *UserHandler {
	return &UserHandler{svc: svc}
}

// ListUsers handles GET /api/users — returns all pre-seeded users
// for the frontend user-picker dropdown.
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.ListUsers(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), "failed to fetch users")
		return
	}
	response.Success(w, http.StatusOK, users, "users retrieved")
}
