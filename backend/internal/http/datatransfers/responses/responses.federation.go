// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package responses

import (
	"time"

	"geregetemplateai/internal/business/domain"
)

// FedPeerResponse нь федерацийн гишүүн node.
type FedPeerResponse struct {
	Id        string     `json:"id"`
	Key       string     `json:"key"`
	Name      string     `json:"name"`
	OrgId     string     `json:"org_id"`
	BaseUrl   string     `json:"base_url"`
	JwksUrl   string     `json:"jwks_url"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func FromFedPeer(p domain.FedPeer) FedPeerResponse {
	return FedPeerResponse{
		Id: p.ID, Key: p.Key, Name: p.Name, OrgId: p.OrgID,
		BaseUrl: p.BaseURL, JwksUrl: p.JWKSURL, Status: p.Status,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func ToFedPeerList(items []domain.FedPeer) []FedPeerResponse {
	out := make([]FedPeerResponse, 0, len(items))
	for _, p := range items {
		out = append(out, FromFedPeer(p))
	}
	return out
}

// FedStatusResponse нь энэ node-ийн федерацийн төлөв.
type FedStatusResponse struct {
	Configured bool   `json:"configured"`
	NodeId     string `json:"node_id"`
}
