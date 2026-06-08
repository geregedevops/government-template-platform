// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package domain

import "time"

// Дуу хоолойн орчуулгын domain entity-үүд. domain.ai.go-тэй ижил зарчмаар
// зөвхөн стандарт сангаас хамаарна — HTTP, GORM, Gemini зэрэг гадаад
// хамаарлууд энд орохгүй.

// Дэмжигдсэн хэлний кодууд — DB CHECK constraint болон Gemini prompt-той
// ЯГ таарах ёстой.
const (
	VoiceLangMN = "mn"
	VoiceLangEN = "en"
)

// VoiceTranslation нь нэг дуу хоолойн орчуулгын бичлэг (эх яриа → орчуулга).
type VoiceTranslation struct {
	ID             string
	UserID         string
	SourceLang     string
	TargetLang     string
	SourceText     string
	TranslatedText string
	Model          string
	CreatedAt      time.Time
}

// VoiceUsage нь нэг дуу хоолойн дуудлагын токен зарцуулалтын бичлэг —
// кост хяналт, хэрэглэгч тус бүрийн метеринг, аудитад ашиглагдана.
type VoiceUsage struct {
	ID            string
	UserID        string
	TranslationID string
	Model         string
	InputTokens   int
	OutputTokens  int
	CreatedAt     time.Time
}
