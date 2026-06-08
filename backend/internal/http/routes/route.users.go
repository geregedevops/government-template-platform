// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package routes

import (
	"geregetemplateai/internal/business/domain"
	"geregetemplateai/internal/business/usecases/users"
	usershandler "geregetemplateai/internal/http/handlers/v1/users"
	"geregetemplateai/internal/http/middlewares"
	"github.com/gofiber/fiber/v3"
)

// usersRoute нь /users/* бүлгийг холбоно — хэрэглэгчийн өөрийнх нь
// профайл / өгөгдөлд хамаарах endpoint-ууд. Auth урсгалууд нь
// route.auth.go-д байрладаг.
type usersRoute struct {
	handler        usershandler.Handler
	router         fiber.Router
	authMiddleware fiber.Handler
	perm           middlewares.PermissionResolver
}

// NewUsersRoute нь route модулийг бүтээдэг. authMiddleware нь нэвтэрсэн
// хэрэглэгчийг шалгана; хэрэглэгч удирдлагын endpoint-ууд нь нэмж
// 'users.manage' эрхийг шаардана (RequirePermission).
func NewUsersRoute(router fiber.Router, usersUC users.Usecase, authMiddleware fiber.Handler, perm middlewares.PermissionResolver) *usersRoute {
	return &usersRoute{
		handler:        usershandler.NewHandler(usersUC),
		router:         router,
		authMiddleware: authMiddleware,
		perm:           perm,
	}
}

// Routes нь /users бүлэг болон түүний endpoint-уудыг суулгана.
func (r *usersRoute) Routes() {
	v1 := r.router.Group("/v1")
	usersGrp := v1.Group("/users")
	usersGrp.Use(r.authMiddleware)
	usersGrp.Get("/me", r.handler.GetUserData)

	// Хэрэглэгч удирдлага — 'users.manage' эрх шаардана (admin автоматаар давна).
	manage := middlewares.RequirePermission(r.perm, domain.PermUsersManage)
	usersGrp.Get("/", manage, r.handler.ListUsers)
	usersGrp.Post("/", manage, r.handler.CreateUser)
	usersGrp.Patch("/:id/role", manage, r.handler.UpdateUserRole)
	usersGrp.Patch("/:id/org", manage, r.handler.UpdateUserOrg)
	usersGrp.Delete("/:id", manage, r.handler.DeleteUser)
}
