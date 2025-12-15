package messaging

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

// UserRepo is a minimal interface for fetching user department
type UserRepo interface {
	GetDepartment(ctx context.Context, userID int64) (string, error)
}

type Handler struct {
	repo     *Repository
	userRepo UserRepo
}

func NewHandler(repo *Repository, userRepo UserRepo) *Handler {
	return &Handler{repo: repo, userRepo: userRepo}
}

// GET /api/inbox
func (h *Handler) GetInbox(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	messages, err := h.repo.GetInbox(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch inbox")
	}
	return c.JSON(messages)
}

// POST /api/inbox
func (h *Handler) SendMessage(c *fiber.Ctx) error {
	var req struct {
		ReceiverID int64  `json:"receiver_id"`
		Subject    string `json:"subject"`
		Body       string `json:"body"`
		ParentID   *int64 `json:"parent_id,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}

	senderID := c.Locals("userID").(int64)

	msg := &Message{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		Subject:    req.Subject,
		Body:       req.Body,
		ParentID:   req.ParentID,
	}

	if err := h.repo.SendMessage(c.Context(), msg); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to send message")
	}

	return c.Status(fiber.StatusCreated).JSON(msg)
}

// POST /api/inbox/:id/read
func (h *Handler) MarkMessageRead(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	// In real app, check permission if receiver_id == userID
	if err := h.repo.MarkMessageRead(c.Context(), int64(id)); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to mark read")
	}

	return c.SendStatus(fiber.StatusOK)
}

// GET /api/announcements
func (h *Handler) GetAnnouncements(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	roles := c.Locals("roles").([]string)

	// Fetch user dept
	dept, err := h.userRepo.GetDepartment(c.Context(), userID)
	if err != nil {
		dept = "" // Fallback (maybe new user or error)
	}

	announcements, err := h.repo.GetAnnouncements(c.Context(), userID, dept, roles)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch announcements")
	}

	return c.JSON(announcements)
}

// POST /api/announcements/:id/read
func (h *Handler) MarkAnnouncementRead(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	if err := h.repo.MarkAnnouncementRead(c.Context(), userID, int64(id)); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to mark announcement read")
	}

	return c.SendStatus(fiber.StatusOK)
}

// DELETE /api/announcements/:id
func (h *Handler) DeleteAnnouncement(c *fiber.Ctx) error {
	// Permission checks could be added (CREATE_ANNOUNCEMENTS)
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	if err := h.repo.DeleteAnnouncement(c.Context(), int64(id)); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete announcement")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
// DELETE /api/inbox/:id
func (h *Handler) DeleteMessage(c *fiber.Ctx) error {
	userID := c.Locals("userID").(int64)
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	if err := h.repo.DeleteMessage(c.Context(), int64(id), userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete message")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Helper method for UserRepo needed in main.go
// This should be added to user/repository.go
/*
func (r *Repository) GetDepartment(ctx context.Context, userID int64) (string, error) {
	var dept sql.NullString
	err := r.db.QueryRowContext(ctx, "SELECT department FROM users WHERE id = $1", userID).Scan(&dept)
	return dept.String, err
}
*/
