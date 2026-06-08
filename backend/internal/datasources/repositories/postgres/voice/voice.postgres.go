// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package postgres (voice) нь VoiceRepository-ийн GORM/Postgres
// хэрэгжүүлэлт юм. AI repository-тэй ижил загвар: method бүр өөрийн файлд,
// query бүр withRLS-ээр RLS session GUC-тэй транзакцид ажиллана.
package postgres

import (
	"context"
	"fmt"

	repointerface "geregetemplateai/internal/datasources/repositories/interface"
	"geregetemplateai/internal/datasources/rls"

	"gorm.io/gorm"
)

type postgreVoiceRepository struct {
	conn *gorm.DB
}

func NewVoiceRepository(conn *gorm.DB) repointerface.VoiceRepository {
	return &postgreVoiceRepository{conn: conn}
}

// withRLS нь AI repository-ийн ижил нэртэй функцын хуулбар — query-г
// транзакцид боож app.user_id / app.user_role GUC-уудыг тогтооно.
// context-д Identity байхгүй бол хоосон GUC → RLS бүх мөрийг хаана
// (fail-closed).
func (r *postgreVoiceRepository) withRLS(ctx context.Context, fn func(tx *gorm.DB) error) error {
	id, _ := rls.FromContext(ctx)
	return r.conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(
			`SELECT set_config('app.user_id', ?, true), set_config('app.user_role', ?, true)`,
			id.UserID, string(id.Role),
		).Error; err != nil {
			return fmt.Errorf("set rls session context: %w", err)
		}
		return fn(tx)
	})
}
