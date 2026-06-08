// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package postgres (rbac) нь RBACRepository-г Postgres дээр хэрэгжүүлнэ.
package postgres

import (
	"context"
	"errors"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/domain"
	"geregetemplateai/internal/datasources/records"
	repointerface "geregetemplateai/internal/datasources/repositories/interface"
	"geregetemplateai/internal/datasources/rls"
	"geregetemplateai/pkg/logger"

	"gorm.io/gorm"
)

type postgreRBACRepository struct {
	conn *gorm.DB
}

func NewRBACRepository(conn *gorm.DB) repointerface.RBACRepository {
	return &postgreRBACRepository{conn: conn}
}

// withRLS нь users.postgres.go-той ижил — нэг транзакцид app.user_id/app.user_role
// GUC-уудыг тогтооно.
func (r *postgreRBACRepository) withRLS(ctx context.Context, fn func(tx *gorm.DB) error) error {
	id, _ := rls.FromContext(ctx)
	return r.conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			`SELECT set_config('app.user_id', ?, true), set_config('app.user_role', ?, true)`,
			id.UserID, string(id.Role),
		).Error; err != nil {
			return err
		}
		return fn(tx)
	})
}

func (r *postgreRBACRepository) ListRoles(ctx context.Context) ([]domain.Role, error) {
	var rows []records.Roles
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Order("id ASC").Find(&rows).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list roles", logger.Fields{"repository": "rbac", "error": err.Error()})
		return nil, err
	}
	out := make([]domain.Role, 0, len(rows))
	for _, x := range rows {
		out = append(out, x.ToV1Domain())
	}
	return out, nil
}

func (r *postgreRBACRepository) GetRole(ctx context.Context, id int) (domain.Role, error) {
	var row records.Roles
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).First(&row).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Role{}, apperror.NotFound("role not found")
		}
		return domain.Role{}, err
	}
	return row.ToV1Domain(), nil
}

func (r *postgreRBACRepository) CreateRole(ctx context.Context, in *domain.Role) (domain.Role, error) {
	var row records.Roles
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			INSERT INTO roles(key, name, description, is_system, created_at)
			VALUES (?, ?, ?, false, now())
			RETURNING id, key, name, description, is_system, created_at, updated_at
		`, in.Key, in.Name, in.Description).Scan(&row).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.Role{}, apperror.Conflict("role key already exists")
		}
		logger.ErrorWithContext(ctx, "Failed to create role", logger.Fields{"repository": "rbac", "error": err.Error()})
		return domain.Role{}, err
	}
	if row.Id == 0 {
		return domain.Role{}, apperror.Conflict("role key already exists")
	}
	return row.ToV1Domain(), nil
}

func (r *postgreRBACRepository) UpdateRole(ctx context.Context, in *domain.Role) (domain.Role, error) {
	var row records.Roles
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			UPDATE roles SET name = ?, description = ?, updated_at = now()
			WHERE id = ?
			RETURNING id, key, name, description, is_system, created_at, updated_at
		`, in.Name, in.Description, in.ID).Scan(&row).Error
	})
	if err != nil {
		return domain.Role{}, err
	}
	if row.Id == 0 {
		return domain.Role{}, apperror.NotFound("role not found")
	}
	return row.ToV1Domain(), nil
}

func (r *postgreRBACRepository) DeleteRole(ctx context.Context, id int) error {
	var affected int64
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		res := tx.Exec(`DELETE FROM roles WHERE id = ? AND is_system = false`, id)
		affected = res.RowsAffected
		return res.Error
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		// Систем эрх эсвэл байхгүй.
		return apperror.NotFound("role not found or is a system role")
	}
	return nil
}

func (r *postgreRBACRepository) CountUsersWithRole(ctx context.Context, id int) (int64, error) {
	var n int64
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Model(&records.Users{}).Where("role_id = ?", id).Count(&n).Error
	})
	return n, err
}

func (r *postgreRBACRepository) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	var rows []records.Permissions
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Order("category ASC, key ASC").Find(&rows).Error
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.Permission, 0, len(rows))
	for _, x := range rows {
		out = append(out, x.ToV1Domain())
	}
	return out, nil
}

func (r *postgreRBACRepository) GetRolePermissions(ctx context.Context, roleID int) ([]string, error) {
	var keys []string
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Model(&records.RolePermissions{}).
			Where("role_id = ?", roleID).
			Pluck("permission_key", &keys).Error
	})
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *postgreRBACRepository) SetRolePermissions(ctx context.Context, roleID int, keys []string) error {
	return r.withRLS(ctx, func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM role_permissions WHERE role_id = ?`, roleID).Error; err != nil {
			return err
		}
		for _, k := range keys {
			if err := tx.Exec(
				`INSERT INTO role_permissions(role_id, permission_key) VALUES (?, ?)
				 ON CONFLICT DO NOTHING`, roleID, k).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
