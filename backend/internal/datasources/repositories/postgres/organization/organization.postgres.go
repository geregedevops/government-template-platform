// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package postgres (organization) нь OrganizationRepository-г Postgres (ltree)
// дээр хэрэгжүүлнэ.
package postgres

import (
	"context"
	"errors"
	"strings"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/domain"
	"geregetemplateai/internal/datasources/records"
	repointerface "geregetemplateai/internal/datasources/repositories/interface"
	"geregetemplateai/internal/datasources/rls"
	"geregetemplateai/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type postgreOrgRepository struct {
	conn *gorm.DB
}

func NewOrganizationRepository(conn *gorm.DB) repointerface.OrganizationRepository {
	return &postgreOrgRepository{conn: conn}
}

func (r *postgreOrgRepository) withRLS(ctx context.Context, fn func(tx *gorm.DB) error) error {
	id, _ := rls.FromContext(ctx)
	return r.conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// app.user_org GUC — organizations-ийн org-scoped RLS (path <@ subtree)
		// үүнийг уншина. Хоосон бол org-хязгаар тавихгүй (бүх мод харагдана).
		if err := tx.Exec(
			`SELECT set_config('app.user_id', ?, true), set_config('app.user_role', ?, true), set_config('app.user_org', ?, true)`,
			id.UserID, string(id.Role), id.OrgID,
		).Error; err != nil {
			return err
		}
		return fn(tx)
	})
}

// orgLabel нь uuid-г хүчинтэй ltree label болгоно ("o" + underscore-той uuid;
// цифрээр эхлэхгүй, [a-zA-Z0-9_] зөвшөөрөгдсөн).
func orgLabel(id string) string {
	return "o" + strings.ReplaceAll(id, "-", "_")
}

func (r *postgreOrgRepository) Create(ctx context.Context, in *domain.Organization) (domain.Organization, error) {
	id := uuid.NewString()
	label := orgLabel(id)

	var stored records.Organizations
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		path := label
		var parentArg interface{}
		if strings.TrimSpace(in.ParentID) != "" {
			var parent records.Organizations
			if err := tx.Where("id = ?", in.ParentID).First(&parent).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return apperror.NotFound("parent organization not found")
				}
				return err
			}
			path = parent.Path + "." + label
			parentArg = in.ParentID
		}
		return tx.Raw(`
			INSERT INTO organizations(id, parent_id, path, name, kind, created_at)
			VALUES (?, ?, ?::ltree, ?, ?, now())
			RETURNING id, parent_id, path, name, kind, created_at, updated_at
		`, id, parentArg, path, in.Name, in.Kind).Scan(&stored).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to insert organization", logger.Fields{
			"repository": "organization", "method": "Create", "error": err.Error(), "table": "organizations",
		})
		return domain.Organization{}, err
	}
	if stored.Id == "" {
		return domain.Organization{}, errors.New("insert succeeded but RETURNING produced no row")
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreOrgRepository) Get(ctx context.Context, id string) (domain.Organization, error) {
	var stored records.Organizations
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Where("id = ?", id).First(&stored).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Organization{}, apperror.NotFound("organization not found")
		}
		return domain.Organization{}, err
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreOrgRepository) List(ctx context.Context) ([]domain.Organization, error) {
	var rows []records.Organizations
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		// path-аар эрэмбэлбэл эцэг бүрийн дараа хүүхдүүд нь дараалж ирэх тул
		// клиент талд мод барих хялбар.
		return tx.Order("path").Find(&rows).Error
	})
	if err != nil {
		logger.ErrorWithContext(ctx, "Failed to list organizations", logger.Fields{
			"repository": "organization", "method": "List", "error": err.Error(), "table": "organizations",
		})
		return nil, err
	}
	out := make([]domain.Organization, 0, len(rows))
	for _, x := range rows {
		out = append(out, x.ToV1Domain())
	}
	return out, nil
}

func (r *postgreOrgRepository) Update(ctx context.Context, id, name, kind string) (domain.Organization, error) {
	var stored records.Organizations
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			UPDATE organizations SET name = ?, kind = ?, updated_at = now()
			WHERE id = ?
			RETURNING id, parent_id, path, name, kind, created_at, updated_at
		`, name, kind, id).Scan(&stored).Error
	})
	if err != nil {
		return domain.Organization{}, err
	}
	if stored.Id == "" {
		return domain.Organization{}, apperror.NotFound("organization not found")
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreOrgRepository) Delete(ctx context.Context, id string) error {
	if id == domain.RootOrgID {
		return apperror.BadRequest("cannot delete the root organization")
	}
	return r.withRLS(ctx, func(tx *gorm.DB) error {
		// ON DELETE CASCADE-ээр дэд мод бүхэлдээ устна.
		res := tx.Exec(`DELETE FROM organizations WHERE id = ?`, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return apperror.NotFound("organization not found")
		}
		return nil
	})
}
