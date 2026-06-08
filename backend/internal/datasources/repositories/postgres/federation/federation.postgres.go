// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package postgres (federation) нь FederationRepository-г Postgres дээр хэрэгжүүлнэ.
package postgres

import (
	"context"
	"errors"
	"time"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/domain"
	"geregetemplateai/internal/datasources/records"
	repointerface "geregetemplateai/internal/datasources/repositories/interface"
	"geregetemplateai/internal/datasources/rls"
	"geregetemplateai/pkg/logger"

	"gorm.io/gorm"
)

type postgreFedRepository struct {
	conn *gorm.DB
}

func NewFederationRepository(conn *gorm.DB) repointerface.FederationRepository {
	return &postgreFedRepository{conn: conn}
}

func (r *postgreFedRepository) withRLS(ctx context.Context, fn func(tx *gorm.DB) error) error {
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

// --- Peers ---

func (r *postgreFedRepository) CreatePeer(ctx context.Context, in *domain.FedPeer) (domain.FedPeer, error) {
	var stored records.FedPeers
	var org interface{}
	if in.OrgID != "" {
		org = in.OrgID
	}
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			INSERT INTO fed_peers(id, key, name, org_id, base_url, jwks_url, status, created_at)
			VALUES (uuid_generate_v4(), ?, ?, ?, ?, ?, ?, now())
			RETURNING id, key, name, org_id, base_url, jwks_url, status, created_at, updated_at
		`, in.Key, in.Name, org, in.BaseURL, in.JWKSURL, in.Status).Scan(&stored).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.FedPeer{}, apperror.Conflict("peer key already exists")
		}
		logger.ErrorWithContext(ctx, "Failed to insert fed peer", logger.Fields{"repository": "federation", "method": "CreatePeer", "error": err.Error()})
		return domain.FedPeer{}, err
	}
	if stored.Id == "" {
		return domain.FedPeer{}, apperror.Conflict("peer key already exists")
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreFedRepository) GetPeer(ctx context.Context, id string) (domain.FedPeer, error) {
	return r.getPeerBy(ctx, "id = ?", id)
}
func (r *postgreFedRepository) GetPeerByKey(ctx context.Context, key string) (domain.FedPeer, error) {
	return r.getPeerBy(ctx, "key = ?", key)
}
func (r *postgreFedRepository) getPeerBy(ctx context.Context, cond string, arg string) (domain.FedPeer, error) {
	var stored records.FedPeers
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Where(cond, arg).First(&stored).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.FedPeer{}, apperror.NotFound("peer not found")
		}
		return domain.FedPeer{}, err
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreFedRepository) ListPeers(ctx context.Context) ([]domain.FedPeer, error) {
	var rows []records.FedPeers
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Order("created_at DESC").Find(&rows).Error
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.FedPeer, 0, len(rows))
	for _, x := range rows {
		out = append(out, x.ToV1Domain())
	}
	return out, nil
}

func (r *postgreFedRepository) UpdatePeer(ctx context.Context, in *domain.FedPeer) (domain.FedPeer, error) {
	var stored records.FedPeers
	var org interface{}
	if in.OrgID != "" {
		org = in.OrgID
	}
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			UPDATE fed_peers SET name = ?, org_id = ?, base_url = ?, jwks_url = ?, status = ?, updated_at = now()
			WHERE id = ?
			RETURNING id, key, name, org_id, base_url, jwks_url, status, created_at, updated_at
		`, in.Name, org, in.BaseURL, in.JWKSURL, in.Status, in.ID).Scan(&stored).Error
	})
	if err != nil {
		return domain.FedPeer{}, err
	}
	if stored.Id == "" {
		return domain.FedPeer{}, apperror.NotFound("peer not found")
	}
	return stored.ToV1Domain(), nil
}

func (r *postgreFedRepository) DeletePeer(ctx context.Context, id string) error {
	return r.withRLS(ctx, func(tx *gorm.DB) error {
		res := tx.Exec(`DELETE FROM fed_peers WHERE id = ?`, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return apperror.NotFound("peer not found")
		}
		return nil
	})
}

// --- Outbox ---

func (r *postgreFedRepository) EnqueueOutbox(ctx context.Context, peerID, typ, envelope string) error {
	return r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Exec(`
			INSERT INTO fed_outbox(id, peer_id, typ, envelope, status, attempts, next_attempt_at, created_at)
			VALUES (uuid_generate_v4(), ?, ?, ?, 'pending', 0, now(), now())
		`, peerID, typ, envelope).Error
	})
}

func (r *postgreFedRepository) DueOutbox(ctx context.Context, limit int) ([]domain.FedOutboxMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var rows []records.FedOutbox
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Raw(`
			SELECT id, peer_id, typ, envelope, status, attempts FROM fed_outbox
			WHERE status = 'pending' AND next_attempt_at <= now()
			ORDER BY next_attempt_at LIMIT ?
		`, limit).Scan(&rows).Error
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.FedOutboxMessage, 0, len(rows))
	for _, x := range rows {
		out = append(out, domain.FedOutboxMessage{ID: x.Id, PeerID: x.PeerId, Typ: x.Typ, Envelope: x.Envelope, Status: x.Status, Attempts: x.Attempts})
	}
	return out, nil
}

func (r *postgreFedRepository) MarkOutboxSent(ctx context.Context, id string) error {
	return r.withRLS(ctx, func(tx *gorm.DB) error {
		return tx.Exec(`UPDATE fed_outbox SET status='sent' WHERE id = ?`, id).Error
	})
}

func (r *postgreFedRepository) MarkOutboxRetry(ctx context.Context, id, errMsg string, maxAttempts int) error {
	return r.withRLS(ctx, func(tx *gorm.DB) error {
		// attempts+1; босго давбал dead, эс бөгөөс exponential backoff (2^n минут).
		return tx.Exec(`
			UPDATE fed_outbox SET
				attempts = attempts + 1,
				last_error = ?,
				status = CASE WHEN attempts + 1 >= ? THEN 'dead' ELSE 'pending' END,
				next_attempt_at = now() + (power(2, least(attempts + 1, 8)) * interval '1 minute')
			WHERE id = ?
		`, errMsg, maxAttempts, id).Error
	})
}

// --- Inbox (dedup) ---

func (r *postgreFedRepository) SeenInbox(ctx context.Context, jti, typ, iss string) (bool, error) {
	seen := false
	err := r.withRLS(ctx, func(tx *gorm.DB) error {
		res := tx.Exec(`INSERT INTO fed_inbox(jti, typ, iss, received_at) VALUES (?, ?, ?, ?) ON CONFLICT (jti) DO NOTHING`, jti, typ, iss, time.Now())
		if res.Error != nil {
			return res.Error
		}
		seen = res.RowsAffected == 0 // 0 ⇒ аль хэдийн байсан (давхардсан)
		return nil
	})
	return seen, err
}
