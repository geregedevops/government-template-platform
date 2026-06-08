// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package voice

import (
	"context"
	"time"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/domain"
	"geregetemplateai/pkg/logger"
)

// Translate нь дуу хоолойн орчуулгын бүрэн pipeline-г гүйцэтгэнэ:
//
//  1. өдрийн хязгаар шалгах (Redis тоолуур, ai.Chat-тэй ижил fail-open);
//  2. аудиог бичвэрлэж нөгөө хэл рүү орчуулах (Gemini multimodal, нэг дуудлага);
//  3. орчуулсан бичвэрийг яриа болгох (Gemini TTS → WAV);
//  4. орчуулгыг хадгалж, токен зарцуулалт бичих.
func (u *usecase) Translate(ctx context.Context, req TranslateRequest) (TranslateResponse, error) {
	if !u.cfg.Enabled {
		return TranslateResponse{}, apperror.Unavailable("voice service is not configured")
	}

	// Эх хэлийг шалгаж, зорилтот хэлийг тодорхойлно (хоёр хэлний хооронд).
	var targetLang string
	switch req.SourceLang {
	case domain.VoiceLangMN:
		targetLang = domain.VoiceLangEN
	case domain.VoiceLangEN:
		targetLang = domain.VoiceLangMN
	default:
		return TranslateResponse{}, apperror.BadRequest("source language must be mn or en")
	}

	if len(req.Audio) == 0 {
		return TranslateResponse{}, apperror.BadRequest("audio is required")
	}
	if len(req.Audio) > u.cfg.MaxAudioBytes {
		return TranslateResponse{}, apperror.BadRequest("audio is too large")
	}

	// (1) Өдрийн хязгаар — ai.Chat-тэй ижил загвар.
	if u.cfg.DailyRequestLimit > 0 && u.cache != nil {
		key := DailyCountKey(req.UserID, time.Now())
		count, err := u.cache.Incr(ctx, key)
		if err != nil {
			// Redis-ийн түр саатал бүх үйлчилгээг унагах ёсгүй — нээлттэй
			// бүтэлгүйтнэ (auth middleware cutoff-той ижил шийдвэр).
			logger.WarnWithContext(ctx, "voice daily counter unavailable, allowing request", logger.Fields{
				"usecase": "voice",
				"method":  "Translate",
				"error":   err.Error(),
			})
		} else {
			if count == 1 {
				// TTL суулгаж чадахгүй бол түлхүүр мөнхөрч хэрэглэгчийг
				// байнга хязгаарлах эрсдэлтэй тул алдааг log хийнэ.
				if expErr := u.cache.Expire(ctx, key, 25*time.Hour); expErr != nil {
					logger.WarnWithContext(ctx, "voice daily counter: failed to set TTL", logger.Fields{
						"usecase": "voice",
						"method":  "Translate",
						"error":   expErr.Error(),
					})
				}
			}
			if count > int64(u.cfg.DailyRequestLimit) {
				return TranslateResponse{}, apperror.Forbidden("voice daily request limit exceeded")
			}
		}
	}

	// (2) Бичвэрлэх + орчуулах (нэг multimodal дуудлага).
	tr, err := u.voicer.TranscribeAndTranslate(ctx, req.Audio, req.MimeType, req.SourceLang, targetLang)
	if err != nil {
		logger.ErrorWithContext(ctx, "voice transcribe/translate failed", logger.Fields{
			"usecase": "voice",
			"method":  "Translate",
			"error":   err.Error(),
		})
		return TranslateResponse{}, apperror.Wrap(apperror.Unavailable("voice translation failed"), err)
	}

	// (3) Орчуулгыг яриа болгох (TTS → WAV). TTS амжилтгүй болсон ч (жнь
	// богино/онцгой текст дээр Gemini аудио үүсгэхгүй байх) бичвэр орчуулга
	// нь хэрэгтэй хэвээр тул бүх хүсэлтийг унагахгүй — аудиогүйгээр буцна
	// (frontend нь audio байхгүй үед зүгээр текст харуулна).
	wav, ttsUsage, err := u.voicer.Synthesize(ctx, tr.TranslatedText)
	if err != nil {
		logger.WarnWithContext(ctx, "voice synthesize failed (returning text only)", logger.Fields{
			"usecase": "voice",
			"method":  "Translate",
			"error":   err.Error(),
		})
		// ttsUsage нь алдааны үед аль хэдийн zero — wav-ийг л цэвэрлэнэ.
		wav = nil
	}

	// (4) Орчуулгыг хадгалах.
	saved, err := u.repo.CreateTranslation(ctx, &domain.VoiceTranslation{
		UserID:         req.UserID,
		SourceLang:     req.SourceLang,
		TargetLang:     targetLang,
		SourceText:     tr.SourceText,
		TranslatedText: tr.TranslatedText,
		Model:          u.cfg.Model,
	})
	if err != nil {
		return TranslateResponse{}, mapRepoError(err, "create translation")
	}

	// Токен зарцуулалт — STT/орчуулга + TTS дуудлагуудын нийлбэр. Бичилт
	// амжилтгүй болсон ч хариуг унагахгүй (метеринг нь best-effort).
	if usageErr := u.repo.RecordUsage(ctx, &domain.VoiceUsage{
		UserID:        req.UserID,
		TranslationID: saved.ID,
		Model:         u.cfg.Model,
		InputTokens:   tr.Usage.InputTokens + ttsUsage.InputTokens,
		OutputTokens:  tr.Usage.OutputTokens + ttsUsage.OutputTokens,
	}); usageErr != nil {
		logger.WarnWithContext(ctx, "voice usage record failed (non-fatal)", logger.Fields{
			"usecase":        "voice",
			"method":         "Translate",
			"translation_id": saved.ID,
			"error":          usageErr.Error(),
		})
	}

	return TranslateResponse{Translation: saved, AudioWAV: wav}, nil
}

// ListTranslations нь хэрэглэгчийн орчуулгын түүхийг буцаана.
func (u *usecase) ListTranslations(ctx context.Context, req ListTranslationsRequest) (ListTranslationsResponse, error) {
	translations, err := u.repo.ListTranslations(ctx, req.UserID, req.Offset, req.Limit)
	if err != nil {
		return ListTranslationsResponse{}, mapRepoError(err, "list translations")
	}
	return ListTranslationsResponse{Translations: translations}, nil
}
