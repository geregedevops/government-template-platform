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

// UpdateOrg нь хэрэглэгчийг өөр байгууллагад шилжүүлнэ (admin удирдлага).
// withRLS нь admin/service role-д бүх мөрийг зөвшөөрдөг; org_id-г энгийн
// хэрэглэгч өөрөө өөрчлөхийг H6 trigger хориглоно.
func (r *postgreUserRepository) UpdateOrg(ctx context.Context, id, orgID string) error {
	var affected int64
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		res := tx.Model(&records.Users{}).
			Where("id = ?", id).
			Update("org_id", orgID)
		affected = res.RowsAffected
		return res.Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to update user org", logger.Fields{
			"repository": "users", "method": "UpdateOrg", "file": "users.update_org.go",
			"error": err.Error(), "table": "users", "user_id": id,
		})
		return err
	}
	if affected == 0 {
		return apperror.NotFound("user not found")
	}
	return nil
}
