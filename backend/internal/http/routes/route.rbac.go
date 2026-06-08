// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package routes

import (
	"geregetemplateai/internal/business/domain"
	rbacuc "geregetemplateai/internal/business/usecases/rbac"
	rbachandler "geregetemplateai/internal/http/handlers/v1/rbac"
	"geregetemplateai/internal/http/middlewares"
	"github.com/gofiber/fiber/v3"
)

type rbacRoute struct {
	handler        rbachandler.Handler
	usecase        rbacuc.Usecase
	router         fiber.Router
	authMiddleware fiber.Handler
}

// NewRBACRoute нь /rbac/* бүлгийг холбоно. /rbac/me нь нэвтэрсэн хэрэглэгч бүрт
// нээлттэй (өөрийн эрхээ авах); бусад нь 'roles.manage' эрх шаардана.
func NewRBACRoute(router fiber.Router, rbacUC rbacuc.Usecase, authMiddleware fiber.Handler) *rbacRoute {
	return &rbacRoute{
		handler:        rbachandler.NewHandler(rbacUC),
		usecase:        rbacUC,
		router:         router,
		authMiddleware: authMiddleware,
	}
}

func (r *rbacRoute) Routes() {
	v1 := r.router.Group("/v1")
	grp := v1.Group("/rbac")
	grp.Use(r.authMiddleware)

	// Нэвтэрсэн хэрэглэгч бүр өөрийн эрхүүдээ авч болно (frontend цэс шүүхэд).
	grp.Get("/me", r.handler.MyPermissions)

	// Удирдлага — 'roles.manage' эрх шаардана (admin автоматаар давна).
	manage := middlewares.RequirePermission(r.usecase, domain.PermRolesManage)
	grp.Get("/roles", manage, r.handler.ListRoles)
	grp.Get("/permissions", manage, r.handler.ListPermissions)
	grp.Post("/roles", manage, r.handler.CreateRole)
	grp.Put("/roles/:id", manage, r.handler.UpdateRole)
	grp.Put("/roles/:id/permissions", manage, r.handler.SetRolePermissions)
	grp.Delete("/roles/:id", manage, r.handler.DeleteRole)
}
