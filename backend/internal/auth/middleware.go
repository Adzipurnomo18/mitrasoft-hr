package auth

import (
	"github.com/gofiber/fiber/v2"
)

func JWTMiddleware(jwtMgr Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies("access_token")
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing token")
		}

		claims, err := jwtMgr.Verify(token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		c.Locals("userID", claims.UserID)
		c.Locals("roles", claims.Roles)

		return c.Next()
	}
}

func RequireRoles(rolesAllowed ...string) fiber.Handler {
	allowed := map[string]struct{}{}
	for _, r := range rolesAllowed {
		allowed[r] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		val := c.Locals("roles")
		roles, ok := val.([]string)
		if !ok {
			return fiber.NewError(fiber.StatusForbidden, "no roles")
		}

		for _, r := range roles {
			if _, ok := allowed[r]; ok {
				return c.Next()
			}
		}

		return fiber.NewError(fiber.StatusForbidden, "insufficient permission")
	}
}
