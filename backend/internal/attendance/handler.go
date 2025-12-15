package attendance

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// POST /api/attendance/checkin
func (h *Handler) Checkin(c *fiber.Ctx) error {
	val := c.Locals("userID")
	userID, ok := val.(int64)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user session")
	}
	rec, err := h.svc.Checkin(c.Context(), userID, time.Now().UTC())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(rec)
}

// POST /api/attendance/checkout
func (h *Handler) Checkout(c *fiber.Ctx) error {
	val := c.Locals("userID")
	userID, ok := val.(int64)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user session")
	}
	rec, err := h.svc.Checkout(c.Context(), userID, time.Now().UTC())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(rec)
}

// GET /api/attendance/summary?from=YYYY-MM-DD&to=YYYY-MM-DD
func (h *Handler) GetSummary(c *fiber.Ctx) error {
	val := c.Locals("userID")
	userID, ok := val.(int64)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user session")
	}
	fromStr := c.Query("from")
	toStr := c.Query("to")
	var from, to time.Time
	var err error
	if fromStr == "" || toStr == "" {
		now := time.Now().UTC()
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		to = from.AddDate(0, 1, 0)
	} else {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid from date (YYYY-MM-DD)")
		}
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid to date (YYYY-MM-DD)")
		}
	}
	s, err := h.svc.Summary(c.Context(), userID, from, to)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(s)
}

// GET /api/attendance/list?from=YYYY-MM-DD&to=YYYY-MM-DD
func (h *Handler) GetList(c *fiber.Ctx) error {
	val := c.Locals("userID")
	userID, ok := val.(int64)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user session")
	}
	fromStr := c.Query("from")
	toStr := c.Query("to")
	var from, to time.Time
	var err error
	if fromStr == "" || toStr == "" {
		now := time.Now().UTC()
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		to = from.AddDate(0, 1, 0)
	} else {
		from, err = time.Parse("2006-01-02", fromStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid from date (YYYY-MM-DD)")
		}
		to, err = time.Parse("2006-01-02", toStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid to date (YYYY-MM-DD)")
		}
	}
	items, err := h.svc.List(c.Context(), userID, from, to)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(items)
}

