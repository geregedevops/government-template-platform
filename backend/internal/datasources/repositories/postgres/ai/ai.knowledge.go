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

// maxKnowledgeList нь нэг хуудсанд (мөн prompt-д шигтгэхэд) буцаах бичлэгийн
// дээд тоо.
const maxKnowledgeList = 100

func (r *postgreAIRepository) CreateKnowledge(ctx context.Context, in *domain.AIKnowledge) (domain.AIKnowledge, error) {
	var stored records.AIKnowledgeRecords
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			INSERT INTO ai_knowledge(id, user_id, title, content, created_at)
			VALUES (uuid_generate_v4(), ?, ?, ?, now())
			RETURNING id, user_id, title, content, created_at, updated_at
		`, in.UserID, in.Title, in.Content).Scan(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to insert ai knowledge", logger.Fields{
			"repository": "ai", "method": "CreateKnowledge", "error": err.Error(), "table": "ai_knowledge",
		})
		return domain.AIKnowledge{}, err
	}
	if stored.Id == "" {
		return domain.AIKnowledge{}, errors.New("insert succeeded but RETURNING produced no row")
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreAIRepository) GetKnowledge(ctx context.Context, id string) (domain.AIKnowledge, error) {
	var stored records.AIKnowledgeRecords
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).First(&stored).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.AIKnowledge{}, apperror.NotFound("knowledge not found")
		}
		logger.ErrorWithContext(ctx, "Failed to query ai knowledge", logger.Fields{
			"repository": "ai", "method": "GetKnowledge", "error": err.Error(), "table": "ai_knowledge",
		})
		return domain.AIKnowledge{}, err
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreAIRepository) ListKnowledge(ctx context.Context, userID string, offset, limit int) ([]domain.AIKnowledge, error) {
	if limit <= 0 || limit > maxKnowledgeList {
		limit = maxKnowledgeList
	}
	if offset < 0 {
		offset = 0
	}
	var stored []records.AIKnowledgeRecords
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Where("user_id = ?", userID).
			Order("created_at DESC").
			Offset(offset).Limit(limit).
			Find(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list ai knowledge", logger.Fields{
			"repository": "ai", "method": "ListKnowledge", "error": err.Error(), "table": "ai_knowledge",
		})
		return nil, err
	}
	out := make([]domain.AIKnowledge, 0, len(stored))
	for _, s := range stored {
		out = append(out, s.ToV1Domain())
	}
	return out, nil
}

// ListAllKnowledge нь бүх хэрэглэгчийн мэдлэгийг эзэмшигчийн имэйлтэй нь буцаана.
// withRLS-ийн дотор admin role нь ai_knowledge + users-ийн бүх мөрийг хардаг тул
// admin бүх системийн мэдлэгийг авна; энгийн хэрэглэгч зөвхөн өөрийнхөө.
func (r *postgreAIRepository) ListAllKnowledge(ctx context.Context, offset, limit int) ([]domain.AIKnowledge, error) {
	if limit <= 0 || limit > maxKnowledgeList {
		limit = maxKnowledgeList
	}
	if offset < 0 {
		offset = 0
	}
	type joinRow struct {
		records.AIKnowledgeRecords
		OwnerEmail string `gorm:"column:owner_email"`
	}
	var rows []joinRow
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			SELECT k.id, k.user_id, k.title, k.content, k.created_at, k.updated_at,
			       u.email AS owner_email
			FROM ai_knowledge k
			LEFT JOIN users u ON u.id = k.user_id
			ORDER BY k.created_at DESC
			OFFSET ? LIMIT ?
		`, offset, limit).Scan(&rows).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list all ai knowledge", logger.Fields{
			"repository": "ai", "method": "ListAllKnowledge", "error": err.Error(), "table": "ai_knowledge",
		})
		return nil, err
	}
	out := make([]domain.AIKnowledge, 0, len(rows))
	for _, x := range rows {
		d := x.AIKnowledgeRecords.ToV1Domain()
		d.OwnerEmail = x.OwnerEmail
		out = append(out, d)
	}
	return out, nil
}

func (r *postgreAIRepository) UpdateKnowledge(ctx context.Context, in *domain.AIKnowledge) (domain.AIKnowledge, error) {
	var stored records.AIKnowledgeRecords
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			UPDATE ai_knowledge
			SET title = ?, content = ?, updated_at = now()
			WHERE id = ?
			RETURNING id, user_id, title, content, created_at, updated_at
		`, in.Title, in.Content, in.ID).Scan(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to update ai knowledge", logger.Fields{
			"repository": "ai", "method": "UpdateKnowledge", "error": err.Error(), "table": "ai_knowledge",
		})
		return domain.AIKnowledge{}, err
	}
	if stored.Id == "" {
		return domain.AIKnowledge{}, apperror.NotFound("knowledge not found")
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreAIRepository) DeleteKnowledge(ctx context.Context, id string) error {
	var affected int64
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		res := tx.Exec(`DELETE FROM ai_knowledge WHERE id = ?`, id)
		affected = res.RowsAffected
		return res.Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to delete ai knowledge", logger.Fields{
			"repository": "ai", "method": "DeleteKnowledge", "error": err.Error(), "table": "ai_knowledge",
		})
		return err
	}
	if affected == 0 {
		return apperror.NotFound("knowledge not found")
	}
	return nil
}
