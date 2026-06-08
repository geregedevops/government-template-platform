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

// maxDefinitionList нь нэг хуудсанд буцаах процессын дээд тоо.
const maxDefinitionList = 50

func (r *postgreBPMRepository) CreateDefinition(ctx context.Context, in *domain.BPMProcessDefinition) (domain.BPMProcessDefinition, error) {
	const (
		repositoryName = "bpm"
		funcName       = "CreateDefinition"
		fileName       = "bpm.definitions.go"
	)
	var stored records.BPMProcessDefinitions
	// INSERT ... RETURNING — ai.conversations.go-тэй ижил. definition нь BPMN
	// 2.0 XML-ийг хадгалдаг энгийн text багана (migration 10).
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		orgID := in.OrgID
		if orgID == "" {
			orgID = domain.RootOrgID
		}
		return tx.Raw(`
			INSERT INTO bpm_process_definitions(id, user_id, org_id, name, description, definition, status, version, created_at)
			VALUES (uuid_generate_v4(), ?, ?::uuid, ?, ?, ?, ?, 1, now())
			RETURNING id, user_id, org_id, name, description, definition, status, version, created_at, updated_at
		`, in.UserID, orgID, in.Name, in.Description, in.Definition, in.Status).Scan(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to insert bpm definition", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "bpm_process_definitions",
		})
		return domain.BPMProcessDefinition{}, err
	}
	if stored.Id == "" {
		err := errors.New("insert succeeded but RETURNING produced no row")
		logger.ErrorWithContext(ctx, "Insert returned no row", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "bpm_process_definitions",
		})
		return domain.BPMProcessDefinition{}, err
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreBPMRepository) GetDefinition(ctx context.Context, id string) (domain.BPMProcessDefinition, error) {
	const (
		repositoryName = "bpm"
		funcName       = "GetDefinition"
		fileName       = "bpm.definitions.go"
	)
	var stored records.BPMProcessDefinitions
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).First(&stored).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// RLS-ээр харагдахгүй мөр ч мөн адил NotFound — enumeration хаана.
			return domain.BPMProcessDefinition{}, apperror.NotFound("process not found")
		}
		logger.ErrorWithContext(ctx, "Failed to query bpm definition", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "bpm_process_definitions", "definition_id": id,
		})
		return domain.BPMProcessDefinition{}, err
	}
	return stored.ToV1Domain(), nil
}

// GetDefinitionByName нь нэрээр хамгийн сүүлд шинэчлэгдсэн тодорхойлолтыг олно
// (delegatedTask-ийн зорилтот процесс). Service/admin RLS-ээр дуудагдана.
func (r *postgreBPMRepository) GetDefinitionByName(ctx context.Context, name string) (domain.BPMProcessDefinition, error) {
	var stored records.BPMProcessDefinitions
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Where("name = ?", name).
			Order("COALESCE(updated_at, created_at) DESC").
			First(&stored).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.BPMProcessDefinition{}, apperror.NotFound("process not found")
		}
		return domain.BPMProcessDefinition{}, err
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreBPMRepository) ListDefinitions(ctx context.Context, userID string, offset, limit int) ([]domain.BPMProcessDefinition, error) {
	const (
		repositoryName = "bpm"
		funcName       = "ListDefinitions"
		fileName       = "bpm.definitions.go"
	)
	if limit <= 0 || limit > maxDefinitionList {
		limit = maxDefinitionList
	}
	if offset < 0 {
		offset = 0
	}
	var stored []records.BPMProcessDefinitions
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		// user_id-аар шүүхгүй — харагдах хүрээг RLS шийднэ: энгийн хэрэглэгч
		// зөвхөн өөрийнхөө, admin өөрийн org-subtree-ийн процессыг хардаг.
		_ = userID
		return tx.Order("COALESCE(updated_at, created_at) DESC").
			Offset(offset).Limit(limit).
			Find(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list bpm definitions", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "bpm_process_definitions", "user_id": userID,
		})
		return nil, err
	}
	out := make([]domain.BPMProcessDefinition, 0, len(stored))
	for _, s := range stored {
		out = append(out, s.ToV1Domain())
	}
	return out, nil
}

func (r *postgreBPMRepository) UpdateDefinition(ctx context.Context, in *domain.BPMProcessDefinition) (domain.BPMProcessDefinition, error) {
	const (
		repositoryName = "bpm"
		funcName       = "UpdateDefinition"
		fileName       = "bpm.definitions.go"
	)
	var stored records.BPMProcessDefinitions
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			UPDATE bpm_process_definitions
			SET name = ?, description = ?, definition = ?, status = ?, updated_at = now()
			WHERE id = ?
			RETURNING id, user_id, org_id, name, description, definition, status, version, created_at, updated_at
		`, in.Name, in.Description, in.Definition, in.Status, in.ID).Scan(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to update bpm definition", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "bpm_process_definitions", "definition_id": in.ID,
		})
		return domain.BPMProcessDefinition{}, err
	}
	if stored.Id == "" {
		// RLS-ээр харагдахгүй / байхгүй мөр — NotFound (enumeration хаана).
		return domain.BPMProcessDefinition{}, apperror.NotFound("process not found")
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreBPMRepository) DeleteDefinition(ctx context.Context, id string) error {
	const (
		repositoryName = "bpm"
		funcName       = "DeleteDefinition"
		fileName       = "bpm.definitions.go"
	)
	var affected int64
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		res := tx.Exec(`DELETE FROM bpm_process_definitions WHERE id = ?`, id)
		affected = res.RowsAffected
		return res.Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to delete bpm definition", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "bpm_process_definitions", "definition_id": id,
		})
		return err
	}
	if affected == 0 {
		return apperror.NotFound("process not found")
	}
	return nil
}
