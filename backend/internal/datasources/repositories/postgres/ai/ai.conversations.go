// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package postgres

import (
	"context"
	"errors"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/domain"
	"geregetemplateai/internal/datasources/records"
	"geregetemplateai/pkg/logger"

	"gorm.io/gorm"
)

// maxConversationList нь нэг хуудсанд буцаах харилцан ярианы дээд тоо —
// буруу ажиллаж буй дуудагч бүх хүснэгтийг татаж чадахгүй.
const maxConversationList = 50

func (r *postgreAIRepository) CreateConversation(ctx context.Context, userID, title string) (domain.AIConversation, error) {
	const (
		repositoryName = "ai"
		funcName       = "CreateConversation"
		fileName       = "ai.conversations.go"
	)
	var stored records.AIConversations
	// INSERT ... RETURNING — users.store.go-тэй ижил шалтгаанаар (ID-г
	// сервер талд uuid_generate_v4() үүсгэдэг, нэг round-trip).
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			INSERT INTO ai_conversations(id, user_id, title, created_at)
			VALUES (uuid_generate_v4(), ?, ?, now())
			RETURNING id, user_id, title, created_at, updated_at
		`, userID, title).Scan(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to insert ai conversation", logger.Fields{
			"repository": repositoryName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"table":      "ai_conversations",
		})
		return domain.AIConversation{}, err
	}
	if stored.Id == "" {
		err := errors.New("insert succeeded but RETURNING produced no row")
		logger.ErrorWithContext(ctx, "Insert returned no row", logger.Fields{
			"repository": repositoryName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"table":      "ai_conversations",
		})
		return domain.AIConversation{}, err
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreAIRepository) GetConversation(ctx context.Context, id string) (domain.AIConversation, error) {
	const (
		repositoryName = "ai"
		funcName       = "GetConversation"
		fileName       = "ai.conversations.go"
	)
	var stored records.AIConversations
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).First(&stored).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// RLS-ээр харагдахгүй мөр ч мөн адил NotFound — өөр хэрэглэгчийн
			// харилцан яриа байгаа эсэхийг тоолох (enumeration) боломж олгохгүй.
			return domain.AIConversation{}, apperror.NotFound("conversation not found")
		}
		logger.ErrorWithContext(ctx, "Failed to query ai conversation", logger.Fields{
			"repository":      repositoryName,
			"method":          funcName,
			"file":            fileName,
			"error":           err.Error(),
			"table":           "ai_conversations",
			"conversation_id": id,
		})
		return domain.AIConversation{}, err
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreAIRepository) ListConversations(ctx context.Context, userID string, offset, limit int) ([]domain.AIConversation, error) {
	const (
		repositoryName = "ai"
		funcName       = "ListConversations"
		fileName       = "ai.conversations.go"
	)
	if limit <= 0 || limit > maxConversationList {
		limit = maxConversationList
	}
	if offset < 0 {
		offset = 0
	}
	var stored []records.AIConversations
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		// RLS аль хэдийн эзэмшигчээр шүүдэг ч WHERE-ийг ил бичнэ —
		// defense-in-depth, мөн admin context-д ч зөв үр дүн өгнө.
		return tx.Where("user_id = ?", userID).
			Order("COALESCE(updated_at, created_at) DESC").
			Offset(offset).Limit(limit).
			Find(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list ai conversations", logger.Fields{
			"repository": repositoryName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"table":      "ai_conversations",
			"user_id":    userID,
		})
		return nil, err
	}
	out := make([]domain.AIConversation, 0, len(stored))
	for _, s := range stored {
		out = append(out, s.ToV1Domain())
	}
	return out, nil
}
