// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares

import (
	"time"

	"geregetemplateai/pkg/logger"

	"github.com/gofiber/fiber/v3"
)

// requestIDContextKey нь RequestIDHeader-г тусгана. Тодорхой байх үүднээс
// дотооддоо хадгалсан.
const requestIDContextKey = "X-Request-ID"

// AccessLogMiddleware нь хүсэлт тус бүрд нэг бүтэцлэгдсэн (zap) access log
// бичнэ — статусаар нь level сонгоно (5xx=error, 4xx=warn, бусад=info).
// Өмнө нь fmt.Printf-ээр ANSI өнгөтэй stdout руу бичдэг байсныг (zap-ийг
// тойрч) бүтэцлэгдсэн логтой нэгтгэв.
func AccessLogMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		status := c.Response().StatusCode()
		requestID := "-"
		if v, ok := c.Locals(requestIDContextKey).(string); ok && v != "" {
			requestID = v
		}

		fields := logger.Fields{
			"request_id": requestID,
			"status":     status,
			"method":     c.Method(),
			"path":       c.Path(),
			"latency_ms": time.Since(start).Milliseconds(),
			"ip":         c.IP(),
			"user_agent": c.Get("User-Agent"),
		}
		if err != nil {
			fields["error"] = err.Error()
		}

		switch {
		case status >= 500:
			logger.Error("http request", fields)
		case status >= 400:
			logger.Warn("http request", fields)
		default:
			logger.Info("http request", fields)
		}

		return err
	}
}
