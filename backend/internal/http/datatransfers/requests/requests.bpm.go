// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package requests

import "encoding/json"

// BPMSaveProcessRequest нь процесс үүсгэх / шинэчлэх body юм. BPMN нь цэвэр
// BPMN 2.0 XML (.bpmn) файл (маягтууд нь дотроо embed хийгдсэн) — гүн бүтцийн
// шалгалтыг usecase давхарга (validateDefinition) хийдэг. Хэмжээг route-ийн
// body cap хязгаарладаг.
type BPMSaveProcessRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=200"`
	Description string `json:"description" validate:"omitempty,max=2000"`
	BPMN        string `json:"bpmn" validate:"required"`
	Status      string `json:"status" validate:"omitempty,oneof=draft published"`
}

// BPMSubmitTaskRequest нь нэг form даалгаврыг бөглөх body — хэрэглэгчийн
// хариунууд JSON object хэлбэрээр.
type BPMSubmitTaskRequest struct {
	Data json.RawMessage `json:"data" validate:"required"`
}

// BPMGenerateRequest нь текст тайлбараас AI-аар процесс үүсгэх body.
type BPMGenerateRequest struct {
	Description string `json:"description" validate:"required,min=4,max=4000"`
}

// BPMSaveFormRequest нь хуваалцсан форм үүсгэх/засах body. Schema нь form-js
// схем (JSON object).
type BPMSaveFormRequest struct {
	Name   string          `json:"name" validate:"required,min=1,max=200"`
	Schema json.RawMessage `json:"schema" validate:"required"`
}
