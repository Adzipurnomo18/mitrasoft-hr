package requests

import (
    "strconv"
    "time"

    "github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateRequest(c *fiber.Ctx) error {
    val := c.Locals("userID")
    userID, ok := val.(int64)
    if !ok {
        return fiber.NewError(fiber.StatusUnauthorized, "invalid user session")
    }

	var req struct {
		Type      string `json:"type"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Reason    string `json:"reason"`
		// StartTime/EndTime could be separate if needed, assuming ISO8601 strings
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	start, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid start date format (RFC3339 required)"})
	}
	end, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid end date format (RFC3339 required)"})
	}

	created, err := h.service.CreateRequest(c.Context(), userID, req.Type, start, end, req.Reason)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *Handler) GetMyRequests(c *fiber.Ctx) error {
    val := c.Locals("userID")
    userID, ok := val.(int64)
    if !ok {
        return fiber.NewError(fiber.StatusUnauthorized, "invalid user session")
    }

	requests, err := h.service.GetMyRequests(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(requests)
}

func (h *Handler) GetPendingRequests(c *fiber.Ctx) error {
	// TODO: Verify permission (APPROVE_LEAVE or APPROVE_OVERTIME)
	// For now, assume protected route checks general access, but specific permission check is better.

	requests, err := h.service.GetPendingRequests(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(requests)
}

func (h *Handler) ApproveRequest(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

    val := c.Locals("userID")
    approverID, ok := val.(int64)
    if !ok {
        return fiber.NewError(fiber.StatusUnauthorized, "invalid user session")
    }

	if err := h.service.ApproveRequest(c.Context(), id, approverID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Request approved"})
}

func (h *Handler) RejectRequest(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID"})
	}

    val := c.Locals("userID")
    approverID, ok := val.(int64)
    if !ok {
        return fiber.NewError(fiber.StatusUnauthorized, "invalid user session")
    }

	var body struct {
		Reason string `json:"reason"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := h.service.RejectRequest(c.Context(), id, approverID, body.Reason); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Request rejected"})
}
