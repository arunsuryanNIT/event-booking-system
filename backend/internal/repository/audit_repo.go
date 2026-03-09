package repository

import (
	"context"
	"database/sql"

	"github.com/arunsuryan/event-booking-system/backend/internal/model"
)

// AuditRepo is the PostgreSQL implementation of AuditRepository.
type AuditRepo struct {
	db *sql.DB
}

// NewAuditRepo returns an AuditRepo backed by the given database connection.
func NewAuditRepo(db *sql.DB) *AuditRepo {
	return &AuditRepo{db: db}
}

// GetAuditLogs returns the most recent 100 audit log entries, newest first.
// All fields in AuditFilters are optional — nil/empty values are ignored.
// The "$N::type IS NULL OR" pattern lets Postgres skip the condition when the
// parameter is NULL, avoiding dynamic SQL string building and SQL injection risk.
func (r *AuditRepo) GetAuditLogs(ctx context.Context, filters AuditFilters) ([]model.AuditLog, error) {
	query := `
		SELECT
			a.id,
			a.operation,
			a.outcome,
			a.failure_reason,
			a.booking_id,
			a.created_at,
			a.event_id,
			a.user_id,
			e.title  AS event_title,
			u.name   AS user_name
		FROM audit_logs a
		JOIN events e ON e.id = a.event_id
		JOIN users  u ON u.id = a.user_id
		WHERE ($1::uuid IS NULL OR a.event_id   = $1)
		  AND ($2::uuid IS NULL OR a.user_id    = $2)
		  AND ($3::uuid IS NULL OR a.booking_id = $3)
		  AND ($4::text IS NULL OR a.operation   = $4)
		  AND ($5::text IS NULL OR a.outcome     = $5)
		ORDER BY a.created_at DESC
		LIMIT 100`

	// Convert empty strings to nil so Postgres treats them as NULL.
	var operation, outcome interface{}
	if filters.Operation != "" {
		operation = filters.Operation
	}
	if filters.Outcome != "" {
		outcome = filters.Outcome
	}

	rows, err := r.db.QueryContext(ctx, query,
		filters.EventID, filters.UserID, filters.BookingID,
		operation, outcome)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		if err = rows.Scan(
			&l.ID, &l.Operation, &l.Outcome, &l.FailureReason,
			&l.BookingID, &l.CreatedAt, &l.EventID, &l.UserID,
			&l.EventTitle, &l.UserName,
		); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
