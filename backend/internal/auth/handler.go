package auth

import (
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

	"hr-portal-backend/internal/user"
)

// Handler menampung dependency untuk fitur auth.
type Handler struct {
	users  *user.Service
	jwtMgr Manager
}

// NewHandler membuat instance Handler baru.
func NewHandler(users *user.Service, jwtMgr Manager) *Handler {
	return &Handler{
		users:  users,
		jwtMgr: jwtMgr,
	}
}

// ===== DTO =====

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ===== Handler =====

// POST /api/auth/login
func (h *Handler) Login(c *fiber.Ctx) error {
	var req loginRequest

	// Parse JSON body
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid payload")
	}
	if req.Email == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email and password required")
	}

	// Auth ke service
	u, err := h.users.Authenticate(req.Email, req.Password)
	if err != nil {
		// Kredensial salah -> 401
		if errors.Is(err, user.ErrInvalidCredentials) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
		}

		// Error lain (DB, dll) -> 500
		log.Printf("failed to login for %s: %v", req.Email, err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to login")
	}

	// Generate JWT
	token, err := h.jwtMgr.GenerateToken(u.ID, u.Roles)
	if err != nil {
		log.Printf("failed to generate token for %s: %v", req.Email, err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to login")
	}

	// Set cookie JWT
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		HTTPOnly: true,
		Secure:   false, // kalau nanti pakai HTTPS bisa diganti true
		SameSite: "Lax",
		Expires:  time.Now().Add(24 * time.Hour),
	})

	// Response ke frontend
	return c.JSON(fiber.Map{
		"id":    u.ID,
		"name":  u.Name,
		"email": u.Email,
		"roles": u.Roles,
	})
}

// POST /api/auth/logout
func (h *Handler) Logout(c *fiber.Ctx) error {
	// Hapus cookie dengan expire ke masa lalu
	c.Cookie(&fiber.Cookie{
		Name:    "access_token",
		Value:   "",
		Path:    "/",
		Expires: time.Now().Add(-time.Hour),
	})
	return c.SendStatus(fiber.StatusNoContent)
}
