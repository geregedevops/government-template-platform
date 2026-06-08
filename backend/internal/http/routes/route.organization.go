// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package routes

import (
	"geregetemplateai/internal/business/domain"
	orguc "geregetemplateai/internal/business/usecases/organization"
	orghandler "geregetemplateai/internal/http/handlers/v1/organization"
	"geregetemplateai/internal/http/middlewares"
	"github.com/gofiber/fiber/v3"
)

type organizationRoute struct {
	handler        orghandler.Handler
	router         fiber.Router
	authMiddleware fiber.Handler
	perm           middlewares.PermissionResolver
}

// NewOrganizationRoute нь /organizations/* бүлгийг холбоно — бүгд 'org.manage'
// эрх шаардана (admin автоматаар давна).
func NewOrganizationRoute(router fiber.Router, orgUC orguc.Usecase, authMiddleware fiber.Handler, perm middlewares.PermissionResolver) *organizationRoute {
	return &organizationRoute{
		handler:        orghandler.NewHandler(orgUC),
		router:         router,
		authMiddleware: authMiddleware,
		perm:           perm,
	}
}

func (r *organizationRoute) Routes() {
	v1 := r.router.Group("/v1")
	grp := v1.Group("/organizations")
	grp.Use(r.authMiddleware)
	grp.Use(middlewares.RequirePermission(r.perm, domain.PermOrgManage))

	grp.Get("", r.handler.List)
	grp.Post("", r.handler.Create)
	grp.Put("/:id", r.handler.Update)
	grp.Delete("/:id", r.handler.Delete)
}
