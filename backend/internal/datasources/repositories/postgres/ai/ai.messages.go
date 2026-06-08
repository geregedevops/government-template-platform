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

// maxMessageHistory нь нэг харилцан ярианаас ачаалах мессежийн дээд тоо.
// Anthropic руу явуулах контекстын хэмжээг (токен кост) хязгаарлана.
const maxMessageHistory = 40

func (r *postgreAIRepository) StoreMessage(ctx context.Context, in *domain.AIMessage) (domain.AIMessage, error) {
	const (
		repositoryName = "ai"
		funcName       = "StoreMessage"
		fileName       = "ai.messages.go"
	)
	var stored records.AIMessages
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		if err := tx.Raw(`
			INSERT INTO ai_messages(id, conversation_id, user_id, role, content, created_at)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, now())
			RETURNING id, conversation_id, user_id, role, content, created_at
		`, in.ConversationID, in.UserID, in.Role, in.Content).Scan(&stored).Error; err != nil {
			return err
		}
		// Харилцан ярианы "сүүлд идэвхтэй" хугацааг нэг транзакцид сэргээнэ —
		// жагсаалтын эрэмбэ (COALESCE(updated_at, created_at) DESC) зөв байна.
		return tx.Exec(`UPDATE ai_conversations SET updated_at = now() WHERE id = ?`, in.ConversationID).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to insert ai message", logger.Fields{
			"repository":      repositoryName,
			"method":          funcName,
			"file":            fileName,
			"error":           err.Error(),
			"table":           "ai_messages",
			"conversation_id": in.ConversationID,
		})
		return domain.AIMessage{}, err
	}
	if stored.Id == "" {
		err := errors.New("insert succeeded but RETURNING produced no row")
		logger.ErrorWithContext(ctx, "Insert returned no row", logger.Fields{
			"repository": repositoryName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"table":      "ai_messages",
		})
		return domain.AIMessage{}, err
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreAIRepository) ListMessages(ctx context.Context, conversationID string, limit int) ([]domain.AIMessage, error) {
	const (
		repositoryName = "ai"
		funcName       = "ListMessages"
		fileName       = "ai.messages.go"
	)
	// limit > maxMessageHistory үед Claude контекстийг хамгаалж таслана.
	// limit <= 0 нь "бүх мессеж" гэсэн утга (түүхэн харагдацад тасрахгүй) —
	// LIMIT огт тавихгүй.
	if limit > maxMessageHistory {
		limit = maxMessageHistory
	}
	var stored []records.AIMessages
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		// Сүүлийн N мессежийг авч (DESC + limit), дараа нь хуучнаас шинэ рүү
		// эргүүлнэ — урт яриан дээр эхний биш СҮҮЛИЙН контекст хадгалагдана.
		q := tx.Where("conversation_id = ?", conversationID).
			Order("created_at DESC")
		if limit > 0 {
			q = q.Limit(limit)
		}
		return q.Find(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list ai messages", logger.Fields{
			"repository":      repositoryName,
			"method":          funcName,
			"file":            fileName,
			"error":           err.Error(),
			"table":           "ai_messages",
			"conversation_id": conversationID,
		})
		return nil, err
	}
	out := make([]domain.AIMessage, len(stored))
	for i, s := range stored {
		out[len(stored)-1-i] = s.ToV1Domain()
	}
	return out, nil
}
