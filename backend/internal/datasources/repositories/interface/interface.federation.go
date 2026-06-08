// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package _interface

import (
	"context"

	"geregetemplateai/internal/business/domain"
)

// FederationRepository нь федерацийн registry + outbox/inbox-ийн gateway.
type FederationRepository interface {
	// --- Peers (registry) ---
	CreatePeer(ctx context.Context, in *domain.FedPeer) (domain.FedPeer, error)
	GetPeer(ctx context.Context, id string) (domain.FedPeer, error)
	GetPeerByKey(ctx context.Context, key string) (domain.FedPeer, error)
	ListPeers(ctx context.Context) ([]domain.FedPeer, error)
	UpdatePeer(ctx context.Context, in *domain.FedPeer) (domain.FedPeer, error)
	DeletePeer(ctx context.Context, id string) error

	// --- Outbox (durable delivery) ---
	EnqueueOutbox(ctx context.Context, peerID, typ, envelope string) error
	// DueOutbox нь илгээхэд бэлэн (pending, next_attempt_at <= now) мессежүүдийг авна.
	DueOutbox(ctx context.Context, limit int) ([]domain.FedOutboxMessage, error)
	MarkOutboxSent(ctx context.Context, id string) error
	// MarkOutboxRetry нь backoff-той дахин оролдоно; attempts давсан бол dead.
	MarkOutboxRetry(ctx context.Context, id, errMsg string, maxAttempts int) error

	// --- Inbox (idempotency) ---
	// SeenInbox нь jti аль хэдийн боловсруулагдсан эсэхийг буцааж, шинэ бол
	// бүртгэнэ (true ⇒ давхардсан, алгас).
	SeenInbox(ctx context.Context, jti, typ, iss string) (bool, error)
}
