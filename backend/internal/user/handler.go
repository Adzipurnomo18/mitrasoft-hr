package user

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GET /api/employees
func (h *Handler) ListEmployees(c *fiber.Ctx) error {
	search := c.Query("search", "")
	ctx := c.Context()

	list, err := h.svc.ListEmployees(ctx, search)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to fetch employees")
	}
	return c.JSON(list)
}

// GET /api/employees/next-code?department=IT
func (h *Handler) GetNextEmployeeCode(c *fiber.Ctx) error {
	department := c.Query("department", "")
	if department == "" {
		return fiber.NewError(fiber.StatusBadRequest, "department query parameter is required")
	}

	code, err := h.svc.GetNextEmployeeCode(c.Context(), department)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to generate employee code")
	}

	return c.JSON(fiber.Map{
		"employee_code": code,
		"department":    department,
	})
}

// POST /api/employees
func (h *Handler) CreateEmployee(c *fiber.Ctx) error {
	var in EmployeeInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}

	emp, err := h.svc.CreateEmployee(c.Context(), in)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create employee")
	}
	return c.Status(fiber.StatusCreated).JSON(emp)
}

// PUT /api/employees/:id
func (h *Handler) UpdateEmployee(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	var in EmployeeInput
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}

	emp, err := h.svc.UpdateEmployee(c.Context(), id, in)
	if err != nil {
		if err == ErrNotFound {
			return fiber.NewError(fiber.StatusNotFound, "employee not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update employee")
	}
	return c.JSON(emp)
}

// DELETE /api/employees/:id
func (h *Handler) DeleteEmployee(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	if err := h.svc.DeleteEmployee(c.Context(), id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete employee")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// DELETE /api/employees/by-code/:code
func (h *Handler) DeleteEmployeeByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid employee code")
	}
	if err := h.svc.DeleteEmployeeByCode(c.Context(), code); err != nil {
		if err == ErrNotFound {
			return fiber.NewError(fiber.StatusNotFound, "employee not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete employee")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// DELETE /api/employees/:id/hard - permanent delete
func (h *Handler) HardDeleteEmployee(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	if err := h.svc.HardDeleteEmployee(c.Context(), id); err != nil {
		if err == ErrNotFound {
			return fiber.NewError(fiber.StatusNotFound, "employee not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete employee")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// DELETE /api/employees/by-code/:code/hard - permanent delete by code
func (h *Handler) HardDeleteEmployeeByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	if code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid employee code")
	}
	if err := h.svc.HardDeleteEmployeeByCode(c.Context(), code); err != nil {
		if err == ErrNotFound {
			return fiber.NewError(fiber.StatusNotFound, "employee not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete employee")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// GET /api/me - Get current user's profile
func (h *Handler) GetMyProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	id, ok := userID.(int64)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user session")
	}

	user, err := h.svc.GetByID(id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	return c.JSON(user)
}

// PUT /api/me - Update current user's profile
func (h *Handler) UpdateMyProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	id, ok := userID.(int64)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user session")
	}

	var input ProfileInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}

	user, err := h.svc.UpdateProfile(c.Context(), id, &input)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(user)
}
