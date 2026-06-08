// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package responses

import (
	"time"

	"geregetemplateai/internal/business/domain"
)

// VoiceTranslateResponse нь нэг орчуулгын бүрэн үр дүн — бичвэр + тоглуулахад
// бэлэн аудио (base64 WAV). audio_base64 нь зөвхөн шинэ орчуулга дээр
// буцна; түүхэнд хадгалагдахгүй.
type VoiceTranslateResponse struct {
	Id             string    `json:"id"`
	SourceLang     string    `json:"source_lang"`
	TargetLang     string    `json:"target_lang"`
	SourceText     string    `json:"source_text"`
	TranslatedText string    `json:"translated_text"`
	AudioBase64    string    `json:"audio_base64"`
	AudioMime      string    `json:"audio_mime"`
	CreatedAt      time.Time `json:"created_at"`
}

// VoiceTranscribeResponse нь дуу→бичвэр (STT)-ийн үр дүн.
type VoiceTranscribeResponse struct {
	Text string `json:"text"`
}

// VoiceSpeakResponse нь бичвэр→дуу (TTS)-ийн үр дүн (base64 WAV).
type VoiceSpeakResponse struct {
	AudioBase64 string `json:"audio_base64"`
	AudioMime   string `json:"audio_mime"`
}

// VoiceTranslationResponse нь түүхийн нэг бичлэгийн DTO (аудиогүй).
type VoiceTranslationResponse struct {
	Id             string    `json:"id"`
	SourceLang     string    `json:"source_lang"`
	TargetLang     string    `json:"target_lang"`
	SourceText     string    `json:"source_text"`
	TranslatedText string    `json:"translated_text"`
	CreatedAt      time.Time `json:"created_at"`
}

func FromVoiceTranslation(t domain.VoiceTranslation) VoiceTranslationResponse {
	return VoiceTranslationResponse{
		Id:             t.ID,
		SourceLang:     t.SourceLang,
		TargetLang:     t.TargetLang,
		SourceText:     t.SourceText,
		TranslatedText: t.TranslatedText,
		CreatedAt:      t.CreatedAt,
	}
}

func ToVoiceTranslationList(items []domain.VoiceTranslation) []VoiceTranslationResponse {
	out := make([]VoiceTranslationResponse, 0, len(items))
	for _, t := range items {
		out = append(out, FromVoiceTranslation(t))
	}
	return out
}
