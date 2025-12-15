package attendance

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

func (r *Repository) ensureTable(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS attendance (
			id SERIAL PRIMARY KEY,
			user_id INT REFERENCES users(id) ON DELETE CASCADE,
			date DATE NOT NULL,
			checkin_time TIMESTAMP,
			checkout_time TIMESTAMP,
			status VARCHAR(20) DEFAULT 'ABSENT',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (user_id, date)
		)
	`)
	return err
}

// Ensure a record exists for today; returns the record
func (r *Repository) EnsureToday(ctx context.Context, userID int64, now time.Time) (*Record, error) {
	if err := r.ensureTable(ctx); err != nil {
		return nil, err
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	var rec Record
	var ci, co sql.NullTime
	qsel := `SELECT id, user_id, date, checkin_time, checkout_time, status, created_at FROM attendance WHERE user_id = $1 AND date = $2 LIMIT 1`
	err := r.db.QueryRowContext(ctx, qsel, userID, today).Scan(
		&rec.ID, &rec.UserID, &rec.Date, &ci, &co, &rec.Status, &rec.CreatedAt,
	)
	if err == sql.ErrNoRows {
		qins := `INSERT INTO attendance (user_id, date, status) VALUES ($1, $2, 'ABSENT') RETURNING id, user_id, date, checkin_time, checkout_time, status, created_at`
		err = r.db.QueryRowContext(ctx, qins, userID, today).Scan(
			&rec.ID, &rec.UserID, &rec.Date, &ci, &co, &rec.Status, &rec.CreatedAt,
		)
	}
	if err != nil {
		return nil, err
	}
	if ci.Valid {
		t := ci.Time
		rec.CheckinTime = &t
	}
	if co.Valid {
		t := co.Time
		rec.CheckoutTime = &t
	}
	return &rec, nil
}

func (r *Repository) SetCheckin(ctx context.Context, userID int64, now time.Time) (*Record, error) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	q := `
		UPDATE attendance
		SET checkin_time = COALESCE(checkin_time, $1),
		    status = CASE WHEN EXTRACT(HOUR FROM $1) < 9 THEN 'ON_TIME' ELSE 'LATE' END
		WHERE user_id = $2 AND date = $3
		RETURNING id, user_id, date, checkin_time, checkout_time, status, created_at
	`
	var rec Record
	var ci, co sql.NullTime
	err := r.db.QueryRowContext(ctx, q, now, userID, today).Scan(
		&rec.ID, &rec.UserID, &rec.Date, &ci, &co, &rec.Status, &rec.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if ci.Valid {
		t := ci.Time
		rec.CheckinTime = &t
	}
	if co.Valid {
		t := co.Time
		rec.CheckoutTime = &t
	}
	return &rec, nil
}

func (r *Repository) SetCheckout(ctx context.Context, userID int64, now time.Time) (*Record, error) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	q := `
		UPDATE attendance
		SET checkout_time = $1
		WHERE user_id = $2 AND date = $3
		RETURNING id, user_id, date, checkin_time, checkout_time, status, created_at
	`
	var rec Record
	var ci, co sql.NullTime
	err := r.db.QueryRowContext(ctx, q, now, userID, today).Scan(
		&rec.ID, &rec.UserID, &rec.Date, &ci, &co, &rec.Status, &rec.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if ci.Valid {
		t := ci.Time
		rec.CheckinTime = &t
	}
	if co.Valid {
		t := co.Time
		rec.CheckoutTime = &t
	}
	return &rec, nil
}

func (r *Repository) GetSummary(ctx context.Context, userID int64, from, to time.Time) (*Summary, error) {
	q := `
		SELECT
			SUM(CASE WHEN status IN ('ON_TIME','LATE') THEN 1 ELSE 0 END) AS present,
			SUM(CASE WHEN status = 'ON_TIME' THEN 1 ELSE 0 END) AS on_time,
			SUM(CASE WHEN status = 'LATE' THEN 1 ELSE 0 END) AS late,
			SUM(CASE WHEN status = 'ABSENT' THEN 1 ELSE 0 END) AS absent
		FROM attendance
		WHERE user_id = $1 AND date >= $2 AND date < $3
	`
	var s Summary
	err := r.db.QueryRowContext(ctx, q, userID, from, to).Scan(&s.Present, &s.OnTime, &s.Late, &s.Absent)
	if err != nil {
		return nil, err
	}
	s.From = from
	s.To = to
	// Working days = total rows in range; fallback present+absent
	q2 := `SELECT COUNT(*) FROM attendance WHERE user_id = $1 AND date >= $2 AND date < $3`
	var cnt int
	if err := r.db.QueryRowContext(ctx, q2, userID, from, to).Scan(&cnt); err == nil {
		s.WorkingDays = cnt
	}
	return &s, nil
}

func (r *Repository) ListBetween(ctx context.Context, userID int64, from, to time.Time) ([]*Record, error) {
	q := `
		SELECT id, user_id, date, checkin_time, checkout_time, status, created_at
		FROM attendance
		WHERE user_id = $1 AND date >= $2 AND date < $3
		ORDER BY date ASC
	`
	rows, err := r.db.QueryContext(ctx, q, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Record
	for rows.Next() {
		var rec Record
		var ci, co sql.NullTime
		if err := rows.Scan(&rec.ID, &rec.UserID, &rec.Date, &ci, &co, &rec.Status, &rec.CreatedAt); err != nil {
			return nil, err
		}
		if ci.Valid {
			t := ci.Time
			rec.CheckinTime = &t
		}
		if co.Valid {
			t := co.Time
			rec.CheckoutTime = &t
		}
		out = append(out, &rec)
	}
	return out, nil
}
