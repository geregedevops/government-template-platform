// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package postgres

import (
	"context"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/datasources/records"
	"geregetemplateai/pkg/logger"

	"gorm.io/gorm"
)

// UpdateRole нь нэг хэрэглэгчийн role_id-г сольдог (admin удирдлага). withRLS
// нь admin/service role-д бүх мөрийг зөвшөөрдөг (users_update policy).
func (r *postgreUserRepository) UpdateRole(ctx context.Context, id string, roleID int) error {
	const (
		repositoryName = "users"
		funcName       = "UpdateRole"
		fileName       = "users.update_role.go"
	)
	var affected int64
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		res := tx.Model(&records.Users{}).
			Where("id = ?", id).
			Update("role_id", roleID)
		affected = res.RowsAffected
		return res.Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to update user role", logger.Fields{
			"repository": repositoryName, "method": funcName, "file": fileName,
			"error": err.Error(), "table": "users", "user_id": id,
		})
		return err
	}
	if affected == 0 {
		return apperror.NotFound("user not found")
	}
	return nil
}
