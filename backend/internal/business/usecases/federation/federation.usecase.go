// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package federation нь node хооронд гарын үсэгтэй мессеж солих цөм: platform
// registry (peers), гарах outbox (durability), орох inbound (verify + dedup +
// dispatch). Гарын үсгийг pkg/fedsign (ES256) хийнэ.
package federation

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/domain"
	repointerface "geregetemplateai/internal/datasources/repositories/interface"
	"geregetemplateai/internal/datasources/rls"
	"geregetemplateai/pkg/fedsign"
	"geregetemplateai/pkg/logger"
)

const (
	outboxMaxAttempts = 8
	inboundPath       = "/api/fed/inbound" // peer-ийн BFF inbound зам
)

// DelegationHandler нь delegatedTask-ийн орох мессежүүдийг BPM engine-д өгөх
// хил (bpm usecase хэрэгжүүлнэ). nil бол delegation мессеж зөвхөн бүртгэгдэнэ.
type DelegationHandler interface {
	StartDelegated(ctx context.Context, processKey string, vars json.RawMessage, originPeer, parentInstance string) error
	ResumeDelegated(ctx context.Context, parentInstance string, vars json.RawMessage, status string) error
}

// Usecase нь федерацийн үйлдлүүд.
type Usecase interface {
	Configured() bool
	NodeID() string
	JWKS() ([]byte, bool)
	SetDelegationHandler(h DelegationHandler)

	RegisterPeer(ctx context.Context, in domain.FedPeer) (domain.FedPeer, error)
	ListPeers(ctx context.Context) ([]domain.FedPeer, error)
	UpdatePeer(ctx context.Context, in domain.FedPeer) (domain.FedPeer, error)
	DeletePeer(ctx context.Context, id string) error

	// Send нь peer руу (id-аар) гарын үсэгтэй мессеж дараалалд (outbox) нэмнэ.
	Send(ctx context.Context, peerID, typ string, body json.RawMessage) error
	// SendByKey нь peer-ийг key-ээр олж outbox-д нэмнэ (delegatedTask ашиглана).
	SendByKey(ctx context.Context, peerKey, typ string, body json.RawMessage) error
	// HandleInbound нь орж ирсэн гарын үсэгтэй мессежийг шалгаж dispatch хийнэ.
	HandleInbound(ctx context.Context, token []byte) (map[string]any, error)
	// ProcessOutbox нь дараалал дахь бэлэн мессежүүдийг илгээх ажлын нэг алхам.
	ProcessOutbox(ctx context.Context) (sent int)
}

type usecase struct {
	repo       repointerface.FederationRepository
	signer     *fedsign.Signer // nil ⇒ федераци идэвхгүй
	httpc      *http.Client
	nowFn      func() time.Time
	delegation DelegationHandler
}

func NewUsecase(repo repointerface.FederationRepository, signer *fedsign.Signer) Usecase {
	return &usecase{
		repo:   repo,
		signer: signer,
		httpc:  &http.Client{Timeout: 15 * time.Second},
		nowFn:  time.Now,
	}
}

func (u *usecase) SetDelegationHandler(h DelegationHandler) { u.delegation = h }

func (u *usecase) Configured() bool { return u.signer != nil }
func (u *usecase) NodeID() string {
	if u.signer == nil {
		return ""
	}
	return u.signer.NodeID()
}
func (u *usecase) JWKS() ([]byte, bool) {
	if u.signer == nil {
		return nil, false
	}
	b, err := u.signer.JWKS()
	return b, err == nil
}

// --- Peer registry (admin) ---

func (u *usecase) RegisterPeer(ctx context.Context, in domain.FedPeer) (domain.FedPeer, error) {
	if strings.TrimSpace(in.Key) == "" {
		return domain.FedPeer{}, apperror.BadRequest("peer key is required")
	}
	if strings.TrimSpace(in.BaseURL) == "" || strings.TrimSpace(in.JWKSURL) == "" {
		return domain.FedPeer{}, apperror.BadRequest("base_url and jwks_url are required")
	}
	if in.Status == "" {
		in.Status = domain.FedPeerPending
	}
	return u.repo.CreatePeer(ctx, &in)
}
func (u *usecase) ListPeers(ctx context.Context) ([]domain.FedPeer, error) {
	return u.repo.ListPeers(ctx)
}
func (u *usecase) UpdatePeer(ctx context.Context, in domain.FedPeer) (domain.FedPeer, error) {
	return u.repo.UpdatePeer(ctx, &in)
}
func (u *usecase) DeletePeer(ctx context.Context, id string) error {
	return u.repo.DeletePeer(ctx, id)
}

// --- Send (enqueue) ---

func (u *usecase) Send(ctx context.Context, peerID, typ string, body json.RawMessage) error {
	if u.signer == nil {
		return apperror.Unavailable("federation is not configured")
	}
	peer, err := u.repo.GetPeer(ctx, peerID)
	if err != nil {
		return err
	}
	if peer.Status != domain.FedPeerActive {
		return apperror.BadRequest("peer is not active")
	}
	if body == nil {
		body = json.RawMessage("{}")
	}
	env, _, err := u.signer.Sign(typ, peer.Key, body, u.nowFn())
	if err != nil {
		return apperror.InternalCause(err)
	}
	return u.repo.EnqueueOutbox(ctx, peer.ID, typ, env)
}

// SendByKey нь peer-ийг key-ээр олж outbox-д нэмнэ. delegatedTask нь хэрэглэгчийн
// контекстээс дуудагдаж болзошгүй тул registry/outbox хандалтыг service RLS-ээр
// (бичих эрхтэй) гүйцэтгэнэ.
func (u *usecase) SendByKey(ctx context.Context, peerKey, typ string, body json.RawMessage) error {
	if u.signer == nil {
		return apperror.Unavailable("federation is not configured")
	}
	sctx := rls.WithService(ctx)
	peer, err := u.repo.GetPeerByKey(sctx, peerKey)
	if err != nil {
		return err
	}
	if peer.Status != domain.FedPeerActive {
		return apperror.BadRequest("peer is not active")
	}
	if body == nil {
		body = json.RawMessage("{}")
	}
	env, _, err := u.signer.Sign(typ, peer.Key, body, u.nowFn())
	if err != nil {
		return apperror.InternalCause(err)
	}
	return u.repo.EnqueueOutbox(sctx, peer.ID, typ, env)
}

// --- Inbound (verify + dedup + dispatch) ---

func (u *usecase) HandleInbound(ctx context.Context, token []byte) (map[string]any, error) {
	if u.signer == nil {
		return nil, apperror.Unavailable("federation is not configured")
	}
	ctx = rls.WithService(ctx)
	tok := strings.TrimSpace(string(token))
	kid, iss, err := fedsign.PeekUnverified(tok)
	if err != nil {
		return nil, apperror.BadRequest("invalid envelope")
	}
	peer, err := u.repo.GetPeerByKey(ctx, iss)
	if err != nil {
		return nil, apperror.Unauthorized("unknown peer")
	}
	if peer.Status != domain.FedPeerActive {
		return nil, apperror.Forbidden("peer is not active")
	}
	pub, err := u.fetchPeerKey(ctx, peer, kid)
	if err != nil {
		return nil, apperror.Unauthorized("cannot resolve peer key")
	}
	env, err := fedsign.Verify(tok, pub, u.signer.NodeID())
	if err != nil {
		logger.WarnWithContext(ctx, "fed inbound: verify failed", logger.Fields{"iss": iss, "error": err.Error()})
		return nil, apperror.Unauthorized("signature verification failed")
	}
	// Idempotency: ижил jti-г дахин боловсруулахгүй.
	if seen, derr := u.repo.SeenInbox(ctx, env.ID, env.Typ, iss); derr == nil && seen {
		return map[string]any{"ok": true, "duplicate": true, "node": u.signer.NodeID()}, nil
	}
	return u.dispatch(ctx, peer, env)
}

// dispatch нь typ-ээр мессежийг чиглүүлнэ. Одоогоор ping (round-trip батлах);
// delegation.* зэргийг BPM increment-д нэмнэ.
func (u *usecase) dispatch(ctx context.Context, peer domain.FedPeer, env fedsign.Envelope) (map[string]any, error) {
	switch env.Typ {
	case "ping":
		logger.InfoWithContext(ctx, "fed inbound: ping", logger.Fields{"from": peer.Key, "jti": env.ID})
		return map[string]any{"ok": true, "pong": true, "node": u.signer.NodeID()}, nil
	case "delegation.request":
		if u.delegation == nil {
			return map[string]any{"ok": true, "accepted": true, "note": "no delegation handler"}, nil
		}
		var b struct {
			ProcessKey     string          `json:"process_key"`
			Variables      json.RawMessage `json:"variables"`
			ParentInstance string          `json:"parent_instance"`
		}
		_ = json.Unmarshal(env.Body, &b)
		if err := u.delegation.StartDelegated(ctx, b.ProcessKey, b.Variables, peer.Key, b.ParentInstance); err != nil {
			return nil, err
		}
		return map[string]any{"ok": true, "started": true, "node": u.signer.NodeID()}, nil
	case "delegation.callback":
		if u.delegation == nil {
			return map[string]any{"ok": true, "accepted": true, "note": "no delegation handler"}, nil
		}
		var b struct {
			ParentInstance string          `json:"parent_instance"`
			Variables      json.RawMessage `json:"variables"`
			Status         string          `json:"status"`
		}
		_ = json.Unmarshal(env.Body, &b)
		if err := u.delegation.ResumeDelegated(ctx, b.ParentInstance, b.Variables, b.Status); err != nil {
			return nil, err
		}
		return map[string]any{"ok": true, "resumed": true, "node": u.signer.NodeID()}, nil
	default:
		logger.InfoWithContext(ctx, "fed inbound: accepted (no handler)", logger.Fields{"from": peer.Key, "typ": env.Typ, "jti": env.ID})
		return map[string]any{"ok": true, "accepted": true, "node": u.signer.NodeID()}, nil
	}
}

// --- Outbox worker ---

func (u *usecase) ProcessOutbox(ctx context.Context) int {
	if u.signer == nil {
		return 0
	}
	ctx = rls.WithService(ctx)
	due, err := u.repo.DueOutbox(ctx, 50)
	if err != nil {
		logger.ErrorWithContext(ctx, "fed outbox: due query failed", logger.Fields{"error": err.Error()})
		return 0
	}
	sent := 0
	for _, m := range due {
		peer, err := u.repo.GetPeer(ctx, m.PeerID)
		if err != nil {
			_ = u.repo.MarkOutboxRetry(ctx, m.ID, "peer lookup: "+err.Error(), outboxMaxAttempts)
			continue
		}
		if derr := u.deliver(ctx, peer, m.Envelope); derr != nil {
			_ = u.repo.MarkOutboxRetry(ctx, m.ID, derr.Error(), outboxMaxAttempts)
			continue
		}
		_ = u.repo.MarkOutboxSent(ctx, m.ID)
		sent++
	}
	return sent
}

func (u *usecase) deliver(ctx context.Context, peer domain.FedPeer, envelope string) error {
	url := strings.TrimRight(peer.BaseURL, "/") + inboundPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte(envelope)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/jose")
	resp, err := u.httpc.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("peer responded %d", resp.StatusCode)
	}
	return nil
}

// fetchPeerKey нь peer-ийн jwks_url-аас kid-тэй нийтийн түлхүүрийг татна.
func (u *usecase) fetchPeerKey(ctx context.Context, peer domain.FedPeer, kid string) (*ecdsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, peer.JWKSURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := u.httpc.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("jwks responded %d", resp.StatusCode)
	}
	return fedsign.PublicKeyFromJWKS(body, kid)
}
