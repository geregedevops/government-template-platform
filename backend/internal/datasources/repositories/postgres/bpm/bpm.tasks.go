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

func (r *postgreBPMRepository) CreateTask(ctx context.Context, in *domain.BPMTask) (domain.BPMTask, error) {
	const (
		repositoryName = "bpm"
		funcName       = "CreateTask"
		fileName       = "bpm.tasks.go"
	)
	var stored records.BPMTasks
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			INSERT INTO bpm_tasks(id, instance_id, user_id, node_id, status, form, created_at)
			VALUES (uuid_generate_v4(), ?, ?, ?, 'open', ?::jsonb, now())
			RETURNING id, instance_id, user_id, node_id, status, form, submission, created_at, completed_at
		`, in.InstanceID, in.UserID, in.NodeID, in.Form).Scan(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to insert bpm task", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "bpm_tasks",
		})
		return domain.BPMTask{}, err
	}
	if stored.Id == "" {
		return domain.BPMTask{}, errors.New("insert succeeded but RETURNING produced no row")
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreBPMRepository) GetTask(ctx context.Context, id string) (domain.BPMTask, error) {
	const (
		repositoryName = "bpm"
		funcName       = "GetTask"
		fileName       = "bpm.tasks.go"
	)
	var stored records.BPMTasks
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).First(&stored).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.BPMTask{}, apperror.NotFound("task not found")
		}
		logger.ErrorWithContext(ctx, "Failed to query bpm task", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "bpm_tasks", "task_id": id,
		})
		return domain.BPMTask{}, err
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreBPMRepository) GetOpenTaskByInstance(ctx context.Context, instanceID string) (domain.BPMTask, error) {
	const (
		repositoryName = "bpm"
		funcName       = "GetOpenTaskByInstance"
		fileName       = "bpm.tasks.go"
	)
	var stored records.BPMTasks
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Where("instance_id = ? AND status = ?", instanceID, domain.BPMTaskOpen).
			Order("created_at ASC").First(&stored).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.BPMTask{}, apperror.NotFound("task not found")
		}
		logger.ErrorWithContext(ctx, "Failed to query open bpm task", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "bpm_tasks", "instance_id": instanceID,
		})
		return domain.BPMTask{}, err
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreBPMRepository) CompleteTask(ctx context.Context, id, submission string) (domain.BPMTask, error) {
	const (
		repositoryName = "bpm"
		funcName       = "CompleteTask"
		fileName       = "bpm.tasks.go"
	)
	var stored records.BPMTasks
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			UPDATE bpm_tasks
			SET status = 'completed', submission = ?::jsonb, completed_at = now()
			WHERE id = ? AND status = 'open'
			RETURNING id, instance_id, user_id, node_id, status, form, submission, created_at, completed_at
		`, submission, id).Scan(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to complete bpm task", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "bpm_tasks", "task_id": id,
		})
		return domain.BPMTask{}, err
	}
	if stored.Id == "" {
		// Аль хэдийн дууссан, байхгүй, эсвэл RLS-ээр харагдахгүй.
		return domain.BPMTask{}, apperror.Conflict("task already completed")
	}
	return stored.ToV1Domain(), nil
}
