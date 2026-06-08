// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package postgres

import (
	"context"

	"geregetemplateai/internal/business/domain"
	"geregetemplateai/pkg/logger"

	"gorm.io/gorm"
)

func (r *postgreVoiceRepository) RecordUsage(ctx context.Context, in *domain.VoiceUsage) error {
	const (
		repositoryName = "voice"
		funcName       = "RecordUsage"
		fileName       = "voice.usage.go"
	)
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Exec(`
			INSERT INTO voice_usage(id, user_id, translation_id, model, input_tokens, output_tokens, created_at)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?, now())
		`, in.UserID, in.TranslationID, in.Model, in.InputTokens, in.OutputTokens).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to record voice usage", logger.Fields{
			"repository":     repositoryName,
			"method":         funcName,
			"file":           fileName,
			"error":          err.Error(),
			"table":          "voice_usage",
			"translation_id": in.TranslationID,
		})
		return err
	}
	return nil
}
