// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package requests

// VoiceTranslateRequest нь POST /voice/translate-ийн body юм. Аудиог base64
// болгон дамжуулна (JSON-д тохиромжтой); MimeType нь browser-ийн бичсэн
// форматыг (жнь "audio/webm") Gemini рүү дамжуулахад хэрэгтэй.
type VoiceTranslateRequest struct {
	SourceLang string `json:"source_lang" validate:"required,oneof=mn en"`
	// MimeType-г allowlist-аар хязгаарлана — browser-ийн MediaRecorder-аас
	// гардаг танигдсан багц л зөвшөөрнө (codecs параметрийг frontend хасдаг).
	// Дурын мөрийг Gemini рүү дамжуулахгүй.
	MimeType    string `json:"mime_type" validate:"required,oneof=audio/webm audio/mp4 audio/ogg audio/wav"`
	AudioBase64 string `json:"audio_base64" validate:"required"`
}

// VoiceTranscribeRequest нь POST /voice/transcribe-ийн body — чатад дуугаар
// асуухад аудиог бичвэр болгоно (орчуулгагүй).
type VoiceTranscribeRequest struct {
	Lang        string `json:"lang" validate:"required,oneof=mn en"`
	MimeType    string `json:"mime_type" validate:"required,oneof=audio/webm audio/mp4 audio/ogg audio/wav"`
	AudioBase64 string `json:"audio_base64" validate:"required"`
}

// VoiceSpeakRequest нь POST /voice/speak-ийн body — чатын хариуг чанга
// уншуулна. Текстийн дээд урт нь TTS-ийн кост/хугацааг хязгаарлана.
type VoiceSpeakRequest struct {
	Text string `json:"text" validate:"required,min=1,max=5000"`
}
