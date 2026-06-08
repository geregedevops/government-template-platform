// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package routes

import (
	"geregetemplateai/internal/business/domain"
	voiceuc "geregetemplateai/internal/business/usecases/voice"
	voicehandler "geregetemplateai/internal/http/handlers/v1/voice"
	"geregetemplateai/internal/http/middlewares"
	"github.com/gofiber/fiber/v3"
)

// voiceRoute нь /voice/* бүлгийг холбоно — дуу хоолойн орчуулга + түүх.
// Бүх endpoint JWT шаарддаг. Аудио payload нь глобал 1 MiB body cap дотор
// багтдаг тул route түвшинд тусдаа (илүү уужим) body limit тавихгүй —
// харин Gemini руу 2 дуудлага явдаг тул rate limiter нь AI-аас чанга
// (нэг хэрэглэгчид зардал өндөр).
type voiceRoute struct {
	handler        voicehandler.Handler
	router         fiber.Router
	rateLimiter    *middlewares.RateLimiter
	authMiddleware fiber.Handler
	perm           middlewares.PermissionResolver
}

// NewVoiceRoute нь route модулийг бүтээдэг. rateLimiter-г дуудагч эзэмшинэ
// (graceful shutdown үед Stop() хийнэ) — ai route-тэй ижил загвар.
func NewVoiceRoute(router fiber.Router, voiceUC voiceuc.Usecase, authMiddleware fiber.Handler, rateLimiter *middlewares.RateLimiter, perm middlewares.PermissionResolver) *voiceRoute {
	return &voiceRoute{
		handler:        voicehandler.NewHandler(voiceUC),
		router:         router,
		rateLimiter:    rateLimiter,
		authMiddleware: authMiddleware,
		perm:           perm,
	}
}

// Routes нь /v1/voice бүлэг болон endpoint-уудыг суулгана.
func (r *voiceRoute) Routes() {
	v1 := r.router.Group("/v1")
	voiceGrp := v1.Group("/voice")
	voiceGrp.Use(r.authMiddleware)
	voiceGrp.Use(middlewares.RequirePermission(r.perm, domain.PermVoiceTranslate))
	voiceGrp.Use(r.rateLimiter.Middleware())
	voiceGrp.Post("/translate", r.handler.Translate)
	voiceGrp.Get("/history", r.handler.ListTranslations)
	// Чатын дуу хоолой: дуу→бичвэр (STT) ба бичвэр→дуу (TTS).
	voiceGrp.Post("/transcribe", r.handler.Transcribe)
	voiceGrp.Post("/speak", r.handler.Speak)
}
