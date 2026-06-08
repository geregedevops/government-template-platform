// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package postgres

import (
	"context"
	"errors"

	"geregetemplateai/internal/business/domain"
	"geregetemplateai/internal/datasources/records"
	"geregetemplateai/pkg/logger"

	"gorm.io/gorm"
)

// maxTranslationList нь нэг хуудсанд буцаах орчуулгын дээд тоо — буруу
// ажиллаж буй дуудагч бүх хүснэгтийг татаж чадахгүй.
const maxTranslationList = 50

func (r *postgreVoiceRepository) CreateTranslation(ctx context.Context, in *domain.VoiceTranslation) (domain.VoiceTranslation, error) {
	const (
		repositoryName = "voice"
		funcName       = "CreateTranslation"
		fileName       = "voice.translations.go"
	)
	var stored records.VoiceTranslations
	// INSERT ... RETURNING — ai.conversations-тэй ижил шалтгаанаар (ID-г
	// сервер талд uuid_generate_v4() үүсгэдэг, нэг round-trip).
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			INSERT INTO voice_translations(id, user_id, source_lang, target_lang, source_text, translated_text, model, created_at)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?, ?, now())
			RETURNING id, user_id, source_lang, target_lang, source_text, translated_text, model, created_at
		`, in.UserID, in.SourceLang, in.TargetLang, in.SourceText, in.TranslatedText, in.Model).Scan(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to insert voice translation", logger.Fields{
			"repository": repositoryName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"table":      "voice_translations",
		})
		return domain.VoiceTranslation{}, err
	}
	if stored.Id == "" {
		err := errors.New("insert succeeded but RETURNING produced no row")
		logger.ErrorWithContext(ctx, "Insert returned no row", logger.Fields{
			"repository": repositoryName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"table":      "voice_translations",
		})
		return domain.VoiceTranslation{}, err
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreVoiceRepository) ListTranslations(ctx context.Context, userID string, offset, limit int) ([]domain.VoiceTranslation, error) {
	const (
		repositoryName = "voice"
		funcName       = "ListTranslations"
		fileName       = "voice.translations.go"
	)
	if limit <= 0 || limit > maxTranslationList {
		limit = maxTranslationList
	}
	if offset < 0 {
		offset = 0
	}
	var stored []records.VoiceTranslations
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		// RLS аль хэдийн эзэмшигчээр шүүдэг ч WHERE-ийг ил бичнэ —
		// defense-in-depth, мөн admin context-д ч зөв үр дүн өгнө.
		return tx.Where("user_id = ?", userID).
			Order("created_at DESC").
			Offset(offset).Limit(limit).
			Find(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list voice translations", logger.Fields{
			"repository": repositoryName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"table":      "voice_translations",
			"user_id":    userID,
		})
		return nil, err
	}
	out := make([]domain.VoiceTranslation, 0, len(stored))
	for _, s := range stored {
		out = append(out, s.ToV1Domain())
	}
	return out, nil
}
