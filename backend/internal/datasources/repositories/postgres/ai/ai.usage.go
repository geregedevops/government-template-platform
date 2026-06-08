// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package postgres

import (
	"context"

	"geregetemplateai/internal/business/domain"
	"geregetemplateai/pkg/logger"

	"gorm.io/gorm"
)

func (r *postgreAIRepository) RecordUsage(ctx context.Context, in *domain.AIUsage) error {
	const (
		repositoryName = "ai"
		funcName       = "RecordUsage"
		fileName       = "ai.usage.go"
	)
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Exec(`
			INSERT INTO ai_usage(id, user_id, conversation_id, model, input_tokens, output_tokens, created_at)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?, now())
		`, in.UserID, in.ConversationID, in.Model, in.InputTokens, in.OutputTokens).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to record ai usage", logger.Fields{
			"repository":      repositoryName,
			"method":          funcName,
			"file":            fileName,
			"error":           err.Error(),
			"table":           "ai_usage",
			"conversation_id": in.ConversationID,
		})
		return err
	}
	return nil
}
