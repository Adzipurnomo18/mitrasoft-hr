package rbac

import (
	"github.com/gofiber/fiber/v2"
)

// Handler handles RBAC-related HTTP requests
type Handler struct {
	repo *Repository
}

// NewHandler creates a new RBAC handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// GetMyMenus returns the menu tree for current user
// GET /api/me/menus
func (h *Handler) GetMyMenus(c *fiber.Ctx) error {
	// Get roles from JWT middleware context
	rolesVal := c.Locals("roles")
	roles, ok := rolesVal.([]string)
	if !ok || len(roles) == 0 {
		roles = []string{"EMPLOYEE"} // default role
	}

	ctx := c.Context()

	// Get permissions for user's roles
	permissions, err := h.repo.GetPermissionsByRoles(ctx, roles)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get permissions")
	}

	// Get menus based on permissions
	menus, err := h.repo.GetMenusByPermissions(ctx, permissions)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get menus")
	}

	// Build tree structure
	menuTree := BuildMenuTree(menus)

	return c.JSON(menuTree)
}

// GetMyPermissions returns the permissions for current user
// GET /api/me/permissions
func (h *Handler) GetMyPermissions(c *fiber.Ctx) error {
	rolesVal := c.Locals("roles")
	roles, ok := rolesVal.([]string)
	if !ok || len(roles) == 0 {
		roles = []string{"EMPLOYEE"}
	}

	ctx := c.Context()

	permissions, err := h.repo.GetPermissionsByRoles(ctx, roles)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get permissions")
	}

	return c.JSON(fiber.Map{
		"roles":       roles,
		"permissions": permissions,
	})
}
