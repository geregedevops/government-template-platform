// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package bpm нь /bpm/* HTTP endpoint-уудыг үйлчилдэг — процессын тодорхойлолт
// (CRUD) болон гүйлт (instance эхлүүлэх, даалгавар бөглөх). Бүх endpoint auth
// middleware шаарддаг тул нэргүй хандалт байхгүй.
package bpm

import (
	bpmuc "geregetemplateai/internal/business/usecases/bpm"
)

// Handler нь BPM handler-ийн нэгтгэл; endpoint бүрийн method-ууд өөрсдийн
// файлд (bpm.definitions.go, bpm.runtime.go) тодорхойлогддог.
type Handler struct {
	usecase bpmuc.Usecase
}

func NewHandler(usecase bpmuc.Usecase) Handler {
	return Handler{usecase: usecase}
}
