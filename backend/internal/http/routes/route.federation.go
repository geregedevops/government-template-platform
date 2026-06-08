// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package routes

import (
	"geregetemplateai/internal/business/domain"
	feduc "geregetemplateai/internal/business/usecases/federation"
	fedhandler "geregetemplateai/internal/http/handlers/v1/federation"
	"geregetemplateai/internal/http/middlewares"
	"github.com/gofiber/fiber/v3"
)

type federationRoute struct {
	handler        fedhandler.Handler
	router         fiber.Router
	authMiddleware fiber.Handler
	perm           middlewares.PermissionResolver
}

// NewFederationRoute нь /fed/* бүлгийг холбоно. Peer registry нь 'fed.manage'
// эрх шаардана; харин /fed/inbound нь НЭВТРЭЛТГҮЙ — peer-ийн ES256 гарын
// үсгээр баталгаажна (machine-to-machine).
func NewFederationRoute(router fiber.Router, fedUC feduc.Usecase, authMiddleware fiber.Handler, perm middlewares.PermissionResolver) *federationRoute {
	return &federationRoute{
		handler:        fedhandler.NewHandler(fedUC),
		router:         router,
		authMiddleware: authMiddleware,
		perm:           perm,
	}
}

func (r *federationRoute) Routes() {
	v1 := r.router.Group("/v1")

	// Inbound — нэвтрэлтгүй, гарын үсгээр баталгаажна.
	v1.Post("/fed/inbound", r.handler.Inbound)

	// Удирдлага — admin/fed.manage.
	grp := v1.Group("/fed")
	grp.Use(r.authMiddleware)
	grp.Use(middlewares.RequirePermission(r.perm, domain.PermFedManage))
	grp.Get("/status", r.handler.Status)
	grp.Get("/peers", r.handler.ListPeers)
	grp.Post("/peers", r.handler.CreatePeer)
	grp.Put("/peers/:id", r.handler.UpdatePeer)
	grp.Delete("/peers/:id", r.handler.DeletePeer)
	grp.Post("/peers/:id/ping", r.handler.Ping)
	grp.Post("/flush", r.handler.Flush)
}
