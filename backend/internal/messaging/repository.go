package messaging

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type Message struct {
	ID           int64     `json:"id"`
	SenderID     int64     `json:"sender_id"`
	ReceiverID   int64     `json:"receiver_id"`
	Subject      string    `json:"subject"`
	Body         string    `json:"body"`
	IsRead       bool      `json:"is_read"`
	CreatedAt    time.Time `json:"created_at"`
	ParentID     *int64    `json:"parent_id,omitempty"`
	SenderName   string    `json:"sender_name"`
	ReceiverName string    `json:"receiver_name"`
}

type Announcement struct {
	ID                int64     `json:"id"`
	Title             string    `json:"title"`
	Content           string    `json:"content"`
	TargetDepartments []string  `json:"target_departments"`
	TargetRoles       []string  `json:"target_roles"`
	CreatedBy         int64     `json:"created_by"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	IsRead            bool      `json:"is_read,omitempty"` // for user context
}

// GetInbox returns messages received by a user
func (r *Repository) GetInbox(ctx context.Context, userID int64) ([]*Message, error) {
	q := `
		SELECT 
			m.id, m.sender_id, m.receiver_id, m.subject, m.body, m.is_read, m.created_at, m.parent_id,
			s.name as sender_name, r.name as receiver_name
		FROM messages m
		JOIN users s ON m.sender_id = s.id
		JOIN users r ON m.receiver_id = r.id
		WHERE m.receiver_id = $1
		ORDER BY m.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var m Message
		var parentID sql.NullInt64
		if err := rows.Scan(
			&m.ID, &m.SenderID, &m.ReceiverID, &m.Subject, &m.Body, &m.IsRead, &m.CreatedAt, &parentID,
			&m.SenderName, &m.ReceiverName,
		); err != nil {
			return nil, err
		}
		if parentID.Valid {
			pid := parentID.Int64
			m.ParentID = &pid
		}
		messages = append(messages, &m)
	}
	return messages, nil
}

// SendMessage sends a new message
func (r *Repository) SendMessage(ctx context.Context, m *Message) error {
	q := `
		INSERT INTO messages (sender_id, receiver_id, subject, body, parent_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, q, m.SenderID, m.ReceiverID, m.Subject, m.Body, m.ParentID).Scan(&m.ID, &m.CreatedAt)
}

// MarkMessageRead marks a message as read
func (r *Repository) MarkMessageRead(ctx context.Context, messageID int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE messages SET is_read = TRUE WHERE id = $1", messageID)
	return err
}

// GetAnnouncements returns announcements relevant to the user
func (r *Repository) GetAnnouncements(ctx context.Context, userID int64, dept string, roles []string) ([]*Announcement, error) {
	// Logic: Active announcements targeting ALL OR (user's dept OR user's role)
	// Also check if read
	q := `
		SELECT 
			a.id, a.title, a.content, a.target_departments, a.target_roles, a.created_by, a.created_at,
			EXISTS(SELECT 1 FROM announcement_reads ar WHERE ar.announcement_id = a.id AND ar.user_id = $1) as is_read
		FROM announcements a
		WHERE a.is_active = TRUE
		AND (
			(COALESCE(array_length(a.target_departments, 1), 0) = 0 AND COALESCE(array_length(a.target_roles, 1), 0) = 0) -- Target ALL
			OR
			($2 = ANY(a.target_departments)) -- Target Dept
			OR
			($3 && a.target_roles) -- Target Role (overlap check)
		)
		ORDER BY a.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, q, userID, dept, pq.Array(roles))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var announcements []*Announcement
	for rows.Next() {
		var a Announcement
		var targetDepts, targetRoles []string

		if err := rows.Scan(
			&a.ID, &a.Title, &a.Content, pq.Array(&targetDepts), pq.Array(&targetRoles), &a.CreatedBy, &a.CreatedAt, &a.IsRead,
		); err != nil {
			return nil, err
		}
		a.TargetDepartments = targetDepts
		a.TargetRoles = targetRoles
		announcements = append(announcements, &a)
	}
	return announcements, nil
}

// MarkAnnouncementRead marks an announcement as read by a user
func (r *Repository) MarkAnnouncementRead(ctx context.Context, userID, announcementID int64) error {
	q := `
		INSERT INTO announcement_reads (user_id, announcement_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, announcement_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, q, userID, announcementID)
	return err
}

// DeleteMessage deletes a message from the inbox if the user is the receiver
func (r *Repository) DeleteMessage(ctx context.Context, messageID int64, userID int64) error {
	q := `DELETE FROM messages WHERE id = $1 AND receiver_id = $2`
	result, err := r.db.ExecContext(ctx, q, messageID, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows // Or custom error "message not found or not owned"
	}
	return nil
}
