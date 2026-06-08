// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package postgres

import (
	"context"

	"geregetemplateai/internal/business/domain"
	"geregetemplateai/internal/datasources/records"
	"geregetemplateai/pkg/logger"

	"gorm.io/gorm"
)

// maxEventList нь нэг instance-ийн буцаах event-ийн дээд тоо.
const maxEventList = 500

func (r *postgreBPMRepository) CreateEvent(ctx context.Context, in *domain.BPMEvent) error {
	const (
		repositoryName = "bpm"
		funcName       = "CreateEvent"
		fileName       = "bpm.events.go"
	)
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Exec(`
			INSERT INTO bpm_events(id, instance_id, user_id, type, node_id, detail, created_at)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?, now())
		`, in.InstanceID, in.UserID, in.Type, in.NodeID, in.Detail).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to insert bpm event", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "bpm_events", "instance_id": in.InstanceID,
		})
		return err
	}
	return nil
}

func (r *postgreBPMRepository) ListEvents(ctx context.Context, instanceID string, limit int) ([]domain.BPMEvent, error) {
	const (
		repositoryName = "bpm"
		funcName       = "ListEvents"
		fileName       = "bpm.events.go"
	)
	if limit <= 0 || limit > maxEventList {
		limit = maxEventList
	}
	var stored []records.BPMEvents
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Where("instance_id = ?", instanceID).
			Order("created_at ASC").
			Limit(limit).
			Find(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list bpm events", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "bpm_events", "instance_id": instanceID,
		})
		return nil, err
	}
	out := make([]domain.BPMEvent, 0, len(stored))
	for _, s := range stored {
		out = append(out, s.ToV1Domain())
	}
	return out, nil
}
