// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package routes

import (
	"time"

	"geregetemplateai/internal/business/domain"
	aiuc "geregetemplateai/internal/business/usecases/ai"
	aihandler "geregetemplateai/internal/http/handlers/v1/ai"
	"geregetemplateai/internal/http/middlewares"
	"github.com/gofiber/fiber/v3"
)

// aiBodyMaxBytes — чатын payload жижиг JSON (мессеж ≤4000 тэмдэгт ≈ UTF-8
// кириллээр ~12KB) тул 16 KiB хангалттай.
const aiBodyMaxBytes int64 = 16 * 1024

// aiRoute нь /ai/* бүлгийг холбоно — streaming чат + харилцан ярианы
// түүх. Бүх endpoint JWT шаарддаг; нэргүй гадаргуу байхгүй тул auth-ийн
// rate limiter-аас тусдаа (зөөлөн) хязгаартай.
type aiRoute struct {
	handler        aihandler.Handler
	router         fiber.Router
	rateLimiter    *middlewares.RateLimiter
	authMiddleware fiber.Handler
	perm           middlewares.PermissionResolver
}

// NewAIRoute нь route модулийг бүтээдэг. rateLimiter-г дуудагч эзэмшинэ
// (graceful shutdown үед Stop() хийнэ) — auth route-тэй ижил загвар.
func NewAIRoute(router fiber.Router, aiUC aiuc.Usecase, authMiddleware fiber.Handler, rateLimiter *middlewares.RateLimiter, streamTimeout time.Duration, perm middlewares.PermissionResolver) *aiRoute {
	return &aiRoute{
		handler:        aihandler.NewHandler(aiUC, streamTimeout),
		router:         router,
		rateLimiter:    rateLimiter,
		authMiddleware: authMiddleware,
		perm:           perm,
	}
}

// Routes нь /v1/ai бүлэг болон endpoint-уудыг суулгана.
func (r *aiRoute) Routes() {
	v1 := r.router.Group("/v1")
	aiGrp := v1.Group("/ai")
	aiGrp.Use(r.authMiddleware)
	aiGrp.Use(r.rateLimiter.Middleware())
	aiGrp.Use(middlewares.BodySizeLimitMiddleware(aiBodyMaxBytes))

	chat := middlewares.RequirePermission(r.perm, domain.PermAIChat)
	aiGrp.Post("/chat", chat, r.handler.Chat)
	aiGrp.Get("/conversations", chat, r.handler.ListConversations)
	aiGrp.Get("/conversations/:id/messages", chat, r.handler.GetMessages)

	// Мэдлэгийн сан (CRUD) — чат эдгээрийг system prompt-д шигтгэнэ.
	knowledge := middlewares.RequirePermission(r.perm, domain.PermKnowledgeManage)
	aiGrp.Get("/knowledge/all", knowledge, r.handler.ListAllKnowledge)
	aiGrp.Get("/knowledge", knowledge, r.handler.ListKnowledge)
	aiGrp.Post("/knowledge", knowledge, r.handler.CreateKnowledge)
	aiGrp.Put("/knowledge/:id", knowledge, r.handler.UpdateKnowledge)
	aiGrp.Delete("/knowledge/:id", knowledge, r.handler.DeleteKnowledge)
}
