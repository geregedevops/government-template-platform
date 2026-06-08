// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package voice нь платформын дуу хоолойн орчуулгын (Gemini) business
// логикийг агуулна: аудиог бичвэрлэх (STT), Монгол↔Англи орчуулах, орчуулсан
// бичвэрийг яриа болгох (TTS), өдрийн хязгаар сахиулах, токен зарцуулалт
// бичих. AI чат (ai package)-тай ижил Clean Architecture загвартай.
package voice

import (
	"context"

	"geregetemplateai/internal/business/domain"
)

// Usecase нь оролтын хил (input boundary) юм. ai.Usecase-тэй ижил
// Request/Response struct загвар.
type Usecase interface {
	// Translate нь нэг аудио хэрчмийг бичвэрлэж, нөгөө хэл рүү орчуулж,
	// орчуулгыг яриа болгоно. Үр дүн нь хадгалагдсан орчуулга + WAV аудио.
	Translate(ctx context.Context, req TranslateRequest) (TranslateResponse, error)
	// ListTranslations нь хэрэглэгчийн орчуулгын түүхийг буцаана.
	ListTranslations(ctx context.Context, req ListTranslationsRequest) (ListTranslationsResponse, error)
	// Transcribe нь аудиог тэр хэл дээрээ бичвэрлэнэ (орчуулгагүй) — чатад
	// дуугаар асуухад ашиглана.
	Transcribe(ctx context.Context, req TranscribeRequest) (TranscribeResponse, error)
	// Speak нь бичвэрийг яриа болгоно (TTS) — чатын хариуг чанга уншихад
	// ашиглана. Үр дүн нь тоглуулахад бэлэн WAV.
	Speak(ctx context.Context, req SpeakRequest) (SpeakResponse, error)
}

type (
	TranslateRequest struct {
		UserID     string
		SourceLang string // "mn" | "en"
		Audio      []byte // түүхий аудио байт (handler base64-аас задалсан)
		MimeType   string // аудионы MIME (жнь "audio/webm")
	}
	TranslateResponse struct {
		Translation domain.VoiceTranslation
		AudioWAV    []byte // тоглуулахад бэлэн WAV (handler base64 болгоно)
	}

	ListTranslationsRequest struct {
		UserID string
		Offset int
		Limit  int
	}
	ListTranslationsResponse struct {
		Translations []domain.VoiceTranslation
	}

	TranscribeRequest struct {
		UserID   string
		Lang     string // "mn" | "en" — бичвэрлэх хэлний зөвлөмж
		Audio    []byte
		MimeType string
	}
	TranscribeResponse struct {
		Text string
	}

	SpeakRequest struct {
		UserID string
		Text   string
	}
	SpeakResponse struct {
		AudioWAV []byte
	}
)
