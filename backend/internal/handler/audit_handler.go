package handler

import (
	"net/http"

	"github.com/arunsuryan/event-booking-system/backend/internal/repository"
	"github.com/arunsuryan/event-booking-system/backend/internal/response"
	"github.com/arunsuryan/event-booking-system/backend/internal/service"
	"github.com/google/uuid"
)

// AuditHandler handles HTTP requests for the audit log.
type AuditHandler struct {
	svc *service.BookingService
}

// NewAuditHandler returns an AuditHandler wired to the given service.
func NewAuditHandler(svc *service.BookingService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

// GetAuditLogs handles GET /api/audit.
// All query params are optional and combined with AND:
//
//	?event_id=uuid &user_id=uuid &booking_id=uuid &operation=book|cancel &outcome=success|failure
func (h *AuditHandler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filters := repository.AuditFilters{
		Operation: q.Get("operation"),
		Outcome:   q.Get("outcome"),
	}

	if v := q.Get("event_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid event_id", "must be a valid UUID")
			return
		}
		filters.EventID = &id
	}

	if v := q.Get("user_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid user_id", "must be a valid UUID")
			return
		}
		filters.UserID = &id
	}

	if v := q.Get("booking_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "invalid booking_id", "must be a valid UUID")
			return
		}
		filters.BookingID = &id
	}

	logs, err := h.svc.GetAuditLogs(r.Context(), filters)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(), "failed to fetch audit logs")
		return
	}

	response.Success(w, http.StatusOK, logs, "audit logs retrieved")
}
