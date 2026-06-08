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

// maxFormList нь нэг хуудсанд буцаах хуваалцсан формын дээд тоо.
const maxFormList = 100

func (r *postgreBPMRepository) CreateForm(ctx context.Context, in *domain.BPMForm) (domain.BPMForm, error) {
	var stored records.BPMForms
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			INSERT INTO bpm_forms(id, user_id, name, schema, created_at)
			VALUES (uuid_generate_v4(), ?, ?, ?::jsonb, now())
			RETURNING id, user_id, name, schema, created_at, updated_at
		`, in.UserID, in.Name, in.Schema).Scan(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to insert bpm form", logger.Fields{
			"repository": "bpm", "method": "CreateForm", "file": "bpm.forms.go", "error": err.Error(), "table": "bpm_forms",
		})
		return domain.BPMForm{}, err
	}
	if stored.Id == "" {
		return domain.BPMForm{}, errors.New("insert succeeded but RETURNING produced no row")
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreBPMRepository) GetForm(ctx context.Context, id string) (domain.BPMForm, error) {
	var stored records.BPMForms
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).First(&stored).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.BPMForm{}, apperror.NotFound("form not found")
		}
		logger.ErrorWithContext(ctx, "Failed to get bpm form", logger.Fields{
			"repository": "bpm", "method": "GetForm", "file": "bpm.forms.go", "error": err.Error(), "table": "bpm_forms",
		})
		return domain.BPMForm{}, err
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreBPMRepository) ListForms(ctx context.Context, userID string, offset, limit int) ([]domain.BPMForm, error) {
	if limit <= 0 || limit > maxFormList {
		limit = maxFormList
	}
	if offset < 0 {
		offset = 0
	}
	var stored []records.BPMForms
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Where("user_id = ?", userID).
			Order("created_at DESC").
			Offset(offset).Limit(limit).
			Find(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list bpm forms", logger.Fields{
			"repository": "bpm", "method": "ListForms", "file": "bpm.forms.go", "error": err.Error(), "table": "bpm_forms",
		})
		return nil, err
	}
	out := make([]domain.BPMForm, 0, len(stored))
	for _, s := range stored {
		out = append(out, s.ToV1Domain())
	}
	return out, nil
}

func (r *postgreBPMRepository) UpdateForm(ctx context.Context, in *domain.BPMForm) (domain.BPMForm, error) {
	var stored records.BPMForms
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			UPDATE bpm_forms SET name = ?, schema = ?::jsonb, updated_at = now()
			WHERE id = ?
			RETURNING id, user_id, name, schema, created_at, updated_at
		`, in.Name, in.Schema, in.ID).Scan(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to update bpm form", logger.Fields{
			"repository": "bpm", "method": "UpdateForm", "file": "bpm.forms.go", "error": err.Error(), "table": "bpm_forms",
		})
		return domain.BPMForm{}, err
	}
	if stored.Id == "" {
		return domain.BPMForm{}, apperror.NotFound("form not found")
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreBPMRepository) DeleteForm(ctx context.Context, id string) error {
	return r.withRLS(ctx, func(tx *gorm.DB) error {
		res := tx.Exec(`DELETE FROM bpm_forms WHERE id = ?`, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return apperror.NotFound("form not found")
		}
		return nil
	})
}
