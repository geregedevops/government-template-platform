// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package voice нь /voice/* HTTP endpoint-уудыг үйлчилдэг — дуу хоолойн
// орчуулга (STT→орчуулга→TTS) ба орчуулгын түүх. Бүх endpoint auth
// middleware шаарддаг тул нэргүй хандалт байхгүй.
package voice

import (
	voiceuc "geregetemplateai/internal/business/usecases/voice"
)

// Handler нь voice handler-ийн нэгтгэл; endpoint бүрийн method-ууд өөрсдийн
// файлд (voice.translate.go, voice.history.go) тодорхойлогддог.
type Handler struct {
	usecase voiceuc.Usecase
}

func NewHandler(usecase voiceuc.Usecase) Handler {
	return Handler{usecase: usecase}
}
