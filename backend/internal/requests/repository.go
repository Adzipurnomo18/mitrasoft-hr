package requests

import (
	"context"
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type Request struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Type            string    `json:"type"` // LEAVE, OVERTIME
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	Reason          string    `json:"reason"`
	Status          string    `json:"status"` // PENDING, APPROVED, REJECTED
	ApproverID      *int64    `json:"approver_id,omitempty"`
	RejectionReason *string   `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Joins
	UserName     string `json:"user_name,omitempty"`
	ApproverName string `json:"approver_name,omitempty"`
}

func (r *Repository) Create(ctx context.Context, req *Request) error {
	q := `
		INSERT INTO requests (user_id, type, start_date, end_date, reason, status)
		VALUES ($1, $2, $3, $4, $5, 'PENDING')
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, q,
		req.UserID, req.Type, req.StartDate, req.EndDate, req.Reason,
	).Scan(&req.ID, &req.CreatedAt, &req.UpdatedAt)
}

func (r *Repository) FindByUserID(ctx context.Context, userID int64) ([]*Request, error) {
	q := `
		SELECT 
			r.id, r.user_id, r.type, r.start_date, r.end_date, r.reason, r.status, 
			r.approver_id, r.rejection_reason, r.created_at, r.updated_at,
			u.name as user_name
		FROM requests r
		JOIN users u ON r.user_id = u.id
		WHERE r.user_id = $1
		ORDER BY r.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*Request
	for rows.Next() {
		var req Request
		// Handle potential NULLs for approver/rejection
		var approverID sql.NullInt64
		var rejectionReason sql.NullString

		if err := rows.Scan(
			&req.ID, &req.UserID, &req.Type, &req.StartDate, &req.EndDate, &req.Reason, &req.Status,
			&approverID, &rejectionReason, &req.CreatedAt, &req.UpdatedAt,
			&req.UserName,
		); err != nil {
			return nil, err
		}
		if approverID.Valid {
			aid := approverID.Int64
			req.ApproverID = &aid
		}
		if rejectionReason.Valid {
			rr := rejectionReason.String
			req.RejectionReason = &rr
		}
		requests = append(requests, &req)
	}
	return requests, nil
}

// FindPending returns all pending requests (for admins/HR)
// In a real app, this might be filtered by department or manager hierarchy
func (r *Repository) FindPending(ctx context.Context) ([]*Request, error) {
	q := `
		SELECT 
			r.id, r.user_id, r.type, r.start_date, r.end_date, r.reason, r.status, 
			r.approver_id, r.rejection_reason, r.created_at, r.updated_at,
			u.name as user_name
		FROM requests r
		JOIN users u ON r.user_id = u.id
		WHERE r.status = 'PENDING'
		ORDER BY r.created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*Request
	for rows.Next() {
		var req Request
		var approverID sql.NullInt64
		var rejectionReason sql.NullString

		if err := rows.Scan(
			&req.ID, &req.UserID, &req.Type, &req.StartDate, &req.EndDate, &req.Reason, &req.Status,
			&approverID, &rejectionReason, &req.CreatedAt, &req.UpdatedAt,
			&req.UserName,
		); err != nil {
			return nil, err
		}
		if approverID.Valid {
			aid := approverID.Int64
			req.ApproverID = &aid
		}
		if rejectionReason.Valid {
			rr := rejectionReason.String
			req.RejectionReason = &rr
		}
		requests = append(requests, &req)
	}
	return requests, nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*Request, error) {
	q := `SELECT id, user_id, type, start_date, end_date, reason, status, approver_id, rejection_reason FROM requests WHERE id = $1`
	var req Request
	var approverID sql.NullInt64
	var rejectionReason sql.NullString

	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&req.ID, &req.UserID, &req.Type, &req.StartDate, &req.EndDate, &req.Reason, &req.Status, &approverID, &rejectionReason,
	)
	if err != nil {
		return nil, err
	}
	if approverID.Valid {
		aid := approverID.Int64
		req.ApproverID = &aid
	}
	if rejectionReason.Valid {
		rr := rejectionReason.String
		req.RejectionReason = &rr
	}
	return &req, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id int64, status string, approverID int64, rejectionReason *string) error {
	q := `
		UPDATE requests 
		SET status = $1, approver_id = $2, rejection_reason = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, q, status, approverID, rejectionReason, id)
	return err
}

type Summary struct {
	Total    int64 `json:"total"`
	Pending  int64 `json:"pending"`
	Approved int64 `json:"approved"`
	Rejected int64 `json:"rejected"`
}

func (r *Repository) SummaryBetween(ctx context.Context, from, to time.Time) (*Summary, error) {
	q := `
		WITH agg AS (
			SELECT
				SUM(CASE WHEN status = 'PENDING'  AND created_at >= $1 AND created_at < $2 THEN 1 ELSE 0 END) AS pending,
				SUM(CASE WHEN status = 'APPROVED' AND updated_at >= $1 AND updated_at < $2 THEN 1 ELSE 0 END) AS approved,
				SUM(CASE WHEN status = 'REJECTED' AND updated_at >= $1 AND updated_at < $2 THEN 1 ELSE 0 END) AS rejected
			FROM requests
		)
		SELECT (pending + approved + rejected) AS total, pending, approved, rejected
		FROM agg
	`
	var s Summary
	if err := r.db.QueryRowContext(ctx, q, from, to).Scan(&s.Total, &s.Pending, &s.Approved, &s.Rejected); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) SummaryByUserBetween(ctx context.Context, userID int64, from, to time.Time) (*Summary, error) {
	q := `
		WITH agg AS (
			SELECT
				SUM(CASE WHEN status = 'PENDING'  AND user_id = $1 AND created_at >= $2 AND created_at < $3 THEN 1 ELSE 0 END) AS pending,
				SUM(CASE WHEN status = 'APPROVED' AND user_id = $1 AND updated_at >= $2 AND updated_at < $3 THEN 1 ELSE 0 END) AS approved,
				SUM(CASE WHEN status = 'REJECTED' AND user_id = $1 AND updated_at >= $2 AND updated_at < $3 THEN 1 ELSE 0 END) AS rejected
			FROM requests
		)
		SELECT (pending + approved + rejected) AS total, pending, approved, rejected
		FROM agg
	`
	var s Summary
	if err := r.db.QueryRowContext(ctx, q, userID, from, to).Scan(&s.Total, &s.Pending, &s.Approved, &s.Rejected); err != nil {
		return nil, err
	}
	return &s, nil
}

// FindProcessedBetween returns APPROVED/REJECTED requests in [from, to)
func (r *Repository) FindProcessedBetween(ctx context.Context, from, to time.Time) ([]*Request, error) {
	q := `
		SELECT 
			r.id, r.user_id, r.type, r.start_date, r.end_date, r.reason, r.status, 
			r.approver_id, r.rejection_reason, r.created_at, r.updated_at,
			u.name as user_name, COALESCE(au.name, '') as approver_name
		FROM requests r
		JOIN users u ON r.user_id = u.id
		LEFT JOIN users au ON r.approver_id = au.id
		WHERE r.status IN ('APPROVED','REJECTED')
		  AND r.updated_at >= $1 AND r.updated_at < $2
		ORDER BY r.updated_at DESC
	`
	rows, err := r.db.QueryContext(ctx, q, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*Request
	for rows.Next() {
		var req Request
		var approverID sql.NullInt64
		var rejectionReason sql.NullString
		var approverName sql.NullString

		if err := rows.Scan(
			&req.ID, &req.UserID, &req.Type, &req.StartDate, &req.EndDate, &req.Reason, &req.Status,
			&approverID, &rejectionReason, &req.CreatedAt, &req.UpdatedAt,
			&req.UserName, &approverName,
		); err != nil {
			return nil, err
		}
		if approverID.Valid {
			aid := approverID.Int64
			req.ApproverID = &aid
		}
		if rejectionReason.Valid {
			rr := rejectionReason.String
			req.RejectionReason = &rr
		}
		if approverName.Valid {
			req.ApproverName = approverName.String
		}
		requests = append(requests, &req)
	}
	return requests, nil
}
