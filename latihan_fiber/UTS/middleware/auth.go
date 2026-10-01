package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"siakaduts/models"
	"siakaduts/utils"
)

const localsUser = "authUser"

// RequireAuth: semua endpoint kecuali login wajib membawa
// header Authorization: Bearer <token>. Gagal -> 401.
func RequireAuth(jwtm *utils.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get(fiber.HeaderAuthorization)
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return utils.Err(c, fiber.StatusUnauthorized, "Token tidak ada atau format salah")
		}
		au, expired, err := jwtm.Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			if expired {
				return utils.Err(c, fiber.StatusUnauthorized, "Token kedaluwarsa")
			}
			return utils.Err(c, fiber.StatusUnauthorized, "Token tidak valid")
		}
		c.Locals(localsUser, au)
		return c.Next()
	}
}

// CurrentUser membaca klaim yang dititipkan RequireAuth.
func CurrentUser(c *fiber.Ctx) (models.AuthUser, bool) {
	u, ok := c.Locals(localsUser).(models.AuthUser)
	return u, ok
}

// RequireRole menolak role lain dengan 403.
func RequireRole(roles ...string) fiber.Handler {
	allow := map[string]bool{}
	for _, r := range roles {
		allow[r] = true
	}
	return func(c *fiber.Ctx) error {
		u, ok := CurrentUser(c)
		if !ok {
			return utils.Err(c, fiber.StatusUnauthorized, "Belum terautentikasi")
		}
		if !allow[u.Role] {
			return utils.Err(c, fiber.StatusForbidden, "Role "+u.Role+" tidak berhak mengakses endpoint ini")
		}
		return c.Next()
	}
}
