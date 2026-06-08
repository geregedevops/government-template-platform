// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares

import (
	"context"
	"net/http"

	httpauth "geregetemplateai/internal/http/auth"
	V1Handler "geregetemplateai/internal/http/handlers/v1"

	"github.com/gofiber/fiber/v3"
)

// PermissionResolver нь нэг role-ийн эрхүүдийг буцаана (rbac.Usecase үүнийг
// хангадаг). Энд interface болгож тодорхойлсон нь import cycle-ээс сэргийлнэ.
type PermissionResolver interface {
	Resolve(ctx context.Context, roleID int) ([]string, error)
}

// RequirePermission нь тухайн эрхгүй хэрэглэгчийг 403-аар татгалзана. authMiddleware-
// ийн ДАРАА ажиллах ёстой (CurrentUser locals-д байх ёстой). admin (IsAdmin) нь
// бүх шалгалтыг давна. Resolve алдаа гарвал fail-closed (403).
func RequirePermission(resolver PermissionResolver, perm string) fiber.Handler {
	return func(c fiber.Ctx) error {
		user, err := httpauth.CurrentUserFromContext(c)
		if err != nil {
			return V1Handler.NewErrorResponse(c, http.StatusUnauthorized, "invalid token")
		}
		if user.IsAdmin {
			return c.Next()
		}
		// Энэ нэмэлтээс өмнө олгогдсон токенд RoleID байхгүй (=0) — энгийн
		// хэрэглэгч (2) гэж үзнэ (admin аль хэдийн дээр давсан).
		roleID := user.RoleID
		if roleID == 0 {
			roleID = 2
		}
		perms, err := resolver.Resolve(c.Context(), roleID)
		if err != nil {
			return V1Handler.NewErrorResponse(c, http.StatusForbidden, "you don't have access for this action")
		}
		for _, p := range perms {
			if p == perm {
				return c.Next()
			}
		}
		return V1Handler.NewErrorResponse(c, http.StatusForbidden, "you don't have access for this action")
	}
}
