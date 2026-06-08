// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package voice

import (
	"context"
	"strings"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/domain"
	"geregetemplateai/pkg/logger"
)

// Transcribe нь аудиог тэр хэл дээрээ бичвэрлэнэ (орчуулгагүй). Чатад
// дуугаар асуухад ашиглана — өдрийн тоолуур хэрэглэхгүй, route-ийн
// rate limiter л хамгаална (чат дотор олон удаа дуудагдаж болзошгүй).
func (u *usecase) Transcribe(ctx context.Context, req TranscribeRequest) (TranscribeResponse, error) {
	if !u.cfg.Enabled {
		return TranscribeResponse{}, apperror.Unavailable("voice service is not configured")
	}
	if req.Lang != domain.VoiceLangMN && req.Lang != domain.VoiceLangEN {
		return TranscribeResponse{}, apperror.BadRequest("language must be mn or en")
	}
	if len(req.Audio) == 0 {
		return TranscribeResponse{}, apperror.BadRequest("audio is required")
	}
	if len(req.Audio) > u.cfg.MaxAudioBytes {
		return TranscribeResponse{}, apperror.BadRequest("audio is too large")
	}

	text, _, err := u.voicer.Transcribe(ctx, req.Audio, req.MimeType, req.Lang)
	if err != nil {
		logger.ErrorWithContext(ctx, "voice transcribe failed", logger.Fields{
			"usecase": "voice",
			"method":  "Transcribe",
			"error":   err.Error(),
		})
		return TranscribeResponse{}, apperror.Wrap(apperror.Unavailable("voice transcription failed"), err)
	}
	return TranscribeResponse{Text: text}, nil
}

// Speak нь бичвэрийг яриа болгоно (TTS → WAV). Чатын хариуг чанга уншихад
// ашиглана.
func (u *usecase) Speak(ctx context.Context, req SpeakRequest) (SpeakResponse, error) {
	if !u.cfg.Enabled {
		return SpeakResponse{}, apperror.Unavailable("voice service is not configured")
	}
	if strings.TrimSpace(req.Text) == "" {
		return SpeakResponse{}, apperror.BadRequest("text is required")
	}

	wav, _, err := u.voicer.Synthesize(ctx, req.Text)
	if err != nil {
		logger.ErrorWithContext(ctx, "voice speak failed", logger.Fields{
			"usecase": "voice",
			"method":  "Speak",
			"error":   err.Error(),
		})
		return SpeakResponse{}, apperror.Wrap(apperror.Unavailable("voice synthesis failed"), err)
	}
	return SpeakResponse{AudioWAV: wav}, nil
}
