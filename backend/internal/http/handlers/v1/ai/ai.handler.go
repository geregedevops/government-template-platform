// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package ai нь /ai/* HTTP endpoint-уудыг үйлчилдэг — streaming чат
// (SSE), харилцан ярианы жагсаалт ба түүх. Бүх endpoint auth middleware
// шаарддаг тул нэргүй хандалт байхгүй.
package ai

import (
	"time"

	aiuc "geregetemplateai/internal/business/usecases/ai"
)

// Handler нь AI handler-ийн нэгтгэл; endpoint бүрийн method-ууд өөрсдийн
// файлд (ai.chat.go, ai.conversations.go) тодорхойлогддог.
type Handler struct {
	usecase aiuc.Usecase
	// streamTimeout нь нэг streaming хариуны дээд хугацаа. Глобал
	// TimeoutMiddleware-ийн 30с нь streaming-д хүрэлцэхгүй байж болзошгүй
	// тул stream дотроо тусдаа deadline ашиглана.
	streamTimeout time.Duration
}

func NewHandler(usecase aiuc.Usecase, streamTimeout time.Duration) Handler {
	if streamTimeout <= 0 {
		streamTimeout = 120 * time.Second
	}
	return Handler{usecase: usecase, streamTimeout: streamTimeout}
}
