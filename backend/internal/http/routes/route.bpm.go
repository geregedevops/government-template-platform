// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package routes

import (
	"geregetemplateai/internal/business/domain"
	bpmuc "geregetemplateai/internal/business/usecases/bpm"
	bpmhandler "geregetemplateai/internal/http/handlers/v1/bpm"
	"geregetemplateai/internal/http/middlewares"
	"github.com/gofiber/fiber/v3"
)

// bpmBodyMaxBytes нь процессын тодорхойлолтын body-ийн дээд хэмжээ. Процессын
// график (node/edge/форм) нь чатаас том байж болзошгүй тул глобал 1 MiB cap-
// аас доош тусдаа (илүү уужим биш — харин тодорхой) хязгаар тавина.
const bpmBodyMaxBytes int64 = 256 * 1024

// bpmRoute нь /bpm/* бүлгийг холбоно — процессын CRUD + гүйлт. Бүх endpoint
// JWT шаарддаг. ai/voice route-тэй ижил загвар: rateLimiter-г дуудагч эзэмшинэ.
type bpmRoute struct {
	handler        bpmhandler.Handler
	router         fiber.Router
	rateLimiter    *middlewares.RateLimiter
	authMiddleware fiber.Handler
	perm           middlewares.PermissionResolver
}

func NewBPMRoute(router fiber.Router, bpmUC bpmuc.Usecase, authMiddleware fiber.Handler, rateLimiter *middlewares.RateLimiter, perm middlewares.PermissionResolver) *bpmRoute {
	return &bpmRoute{
		handler:        bpmhandler.NewHandler(bpmUC),
		router:         router,
		rateLimiter:    rateLimiter,
		authMiddleware: authMiddleware,
		perm:           perm,
	}
}

// Routes нь /v1/bpm бүлэг болон endpoint-уудыг суулгана.
func (r *bpmRoute) Routes() {
	v1 := r.router.Group("/v1")
	bpmGrp := v1.Group("/bpm")
	bpmGrp.Use(r.authMiddleware)
	bpmGrp.Use(middlewares.RequirePermission(r.perm, domain.PermBPMManage))
	bpmGrp.Use(r.rateLimiter.Middleware())
	bpmGrp.Use(middlewares.BodySizeLimitMiddleware(bpmBodyMaxBytes))

	// AI-аар текстээс процесс үүсгэх.
	bpmGrp.Post("/generate", r.handler.GenerateProcess)

	// Процессын тодорхойлолт (CRUD).
	bpmGrp.Post("/processes", r.handler.CreateProcess)
	bpmGrp.Get("/processes", r.handler.ListProcesses)
	bpmGrp.Get("/processes/:id", r.handler.GetProcess)
	bpmGrp.Put("/processes/:id", r.handler.UpdateProcess)
	bpmGrp.Delete("/processes/:id", r.handler.DeleteProcess)

	// Гүйлт (instance эхлүүлэх, идэвхтэй даалгавар, бөглөх).
	bpmGrp.Post("/processes/:id/start", r.handler.StartInstance)
	bpmGrp.Get("/processes/:id/instances", r.handler.ListInstances)
	bpmGrp.Get("/instances/:id/task", r.handler.GetActiveTask)
	bpmGrp.Get("/instances/:id/events", r.handler.ListEvents)
	bpmGrp.Post("/tasks/:id/submit", r.handler.SubmitTask)

	// Хуваалцсан форм сан (олон процесс дунд).
	bpmGrp.Get("/forms", r.handler.ListForms)
	bpmGrp.Post("/forms", r.handler.CreateForm)
	bpmGrp.Put("/forms/:id", r.handler.UpdateForm)
	bpmGrp.Delete("/forms/:id", r.handler.DeleteForm)
}
