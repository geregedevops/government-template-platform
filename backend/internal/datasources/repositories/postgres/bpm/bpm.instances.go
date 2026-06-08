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

func (r *postgreBPMRepository) CreateInstance(ctx context.Context, in *domain.BPMProcessInstance) (domain.BPMProcessInstance, error) {
	const (
		repositoryName = "bpm"
		funcName       = "CreateInstance"
		fileName       = "bpm.instances.go"
	)
	var stored records.BPMProcessInstances
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			INSERT INTO bpm_process_instances(id, definition_id, user_id, status, current_node_id, definition_snapshot, parent_instance_id, origin_peer, variables, created_at)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?, ?, ?, ?::jsonb, now())
			RETURNING id, definition_id, user_id, status, current_node_id, definition_snapshot, parent_instance_id, origin_peer, variables, created_at, updated_at, completed_at
		`, in.DefinitionID, in.UserID, in.Status, in.CurrentNodeID, in.DefinitionSnapshot, nullIfEmpty(in.ParentInstanceID), in.OriginPeer, in.Variables).Scan(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to insert bpm instance", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "bpm_process_instances",
		})
		return domain.BPMProcessInstance{}, err
	}
	if stored.Id == "" {
		return domain.BPMProcessInstance{}, errors.New("insert succeeded but RETURNING produced no row")
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreBPMRepository) GetInstance(ctx context.Context, id string) (domain.BPMProcessInstance, error) {
	const (
		repositoryName = "bpm"
		funcName       = "GetInstance"
		fileName       = "bpm.instances.go"
	)
	var stored records.BPMProcessInstances
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).First(&stored).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.BPMProcessInstance{}, apperror.NotFound("instance not found")
		}
		logger.ErrorWithContext(ctx, "Failed to query bpm instance", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "bpm_process_instances", "instance_id": id,
		})
		return domain.BPMProcessInstance{}, err
	}
	return stored.ToV1Domain(), nil
}

// maxInstanceList нь нэг хуудсанд буцаах гүйлтийн дээд тоо.
const maxInstanceList = 100

func (r *postgreBPMRepository) ListInstances(ctx context.Context, definitionID string, offset, limit int) ([]domain.BPMProcessInstance, error) {
	const (
		repositoryName = "bpm"
		funcName       = "ListInstances"
		fileName       = "bpm.instances.go"
	)
	if limit <= 0 || limit > maxInstanceList {
		limit = maxInstanceList
	}
	if offset < 0 {
		offset = 0
	}
	var stored []records.BPMProcessInstances
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Where("definition_id = ?", definitionID).
			Order("created_at DESC").
			Offset(offset).Limit(limit).
			Find(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list bpm instances", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "bpm_process_instances", "definition_id": definitionID,
		})
		return nil, err
	}
	out := make([]domain.BPMProcessInstance, 0, len(stored))
	for _, s := range stored {
		out = append(out, s.ToV1Domain())
	}
	return out, nil
}

func (r *postgreBPMRepository) UpdateInstance(ctx context.Context, in *domain.BPMProcessInstance) (domain.BPMProcessInstance, error) {
	const (
		repositoryName = "bpm"
		funcName       = "UpdateInstance"
		fileName       = "bpm.instances.go"
	)
	var stored records.BPMProcessInstances
	// completed_at-г status-аас хамааруулж тохируулна: completed/cancelled/failed
	// үед now(), эс бөгөөс хэвээр (NULL).
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			UPDATE bpm_process_instances
			SET status = ?, current_node_id = ?, variables = ?::jsonb, updated_at = now(),
			    completed_at = CASE WHEN ? IN ('completed','cancelled','failed') THEN now() ELSE completed_at END
			WHERE id = ?
			RETURNING id, definition_id, user_id, status, current_node_id, definition_snapshot, parent_instance_id, origin_peer, variables, created_at, updated_at, completed_at
		`, in.Status, in.CurrentNodeID, in.Variables, in.Status, in.ID).Scan(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to update bpm instance", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "bpm_process_instances", "instance_id": in.ID,
		})
		return domain.BPMProcessInstance{}, err
	}
	if stored.Id == "" {
		return domain.BPMProcessInstance{}, apperror.NotFound("instance not found")
	}
	return stored.ToV1Domain(), nil
}

// nullIfEmpty нь хоосон string-ийг SQL NULL болгоно (uuid FK баганад).
func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
