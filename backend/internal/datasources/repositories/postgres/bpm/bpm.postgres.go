// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package postgres (bpm) нь BPMRepository-ийн GORM/Postgres хэрэгжүүлэлт юм.
// ai/voice repository-тэй ижил загвар: method бүр өөрийн файлд, query бүр
// withRLS-ээр RLS session GUC-тэй транзакцид ажиллана.
package postgres

import (
	"context"
	"fmt"

	repointerface "geregetemplateai/internal/datasources/repositories/interface"
	"geregetemplateai/internal/datasources/rls"

	"gorm.io/gorm"
)

type postgreBPMRepository struct {
	conn *gorm.DB
}

func NewBPMRepository(conn *gorm.DB) repointerface.BPMRepository {
	return &postgreBPMRepository{conn: conn}
}

// withRLS нь ai/users repository-ийн ижил нэртэй функцын хуулбар — query-г
// транзакцид боож app.user_id / app.user_role GUC-уудыг SET LOCAL хийнэ.
// context-д Identity байхгүй бол хоосон GUC → RLS бүх мөрийг хаана
// (fail-closed).
func (r *postgreBPMRepository) withRLS(ctx context.Context, fn func(tx *gorm.DB) error) error {
	id, _ := rls.FromContext(ctx)
	return r.conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// app.user_org — процессын org-scoped RLS (admin өөрийн дэд мод) уншина.
		if err := tx.Exec(
			`SELECT set_config('app.user_id', ?, true), set_config('app.user_role', ?, true), set_config('app.user_org', ?, true)`,
			id.UserID, string(id.Role), id.OrgID,
		).Error; err != nil {
			return fmt.Errorf("set rls session context: %w", err)
		}
		return fn(tx)
	})
}
