// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package records

import (
	"time"

	"geregetemplateai/internal/business/domain"
)

// FedPeers нь fed_peers (platform registry) хүснэгтийн GORM model.
type FedPeers struct {
	Id        string     `gorm:"column:id;primaryKey"`
	Key       string     `gorm:"column:key"`
	Name      string     `gorm:"column:name"`
	OrgId     *string    `gorm:"column:org_id"`
	BaseUrl   string     `gorm:"column:base_url"`
	JwksUrl   string     `gorm:"column:jwks_url"`
	Status    string     `gorm:"column:status"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
}

func (FedPeers) TableName() string { return "fed_peers" }

func (r FedPeers) ToV1Domain() domain.FedPeer {
	org := ""
	if r.OrgId != nil {
		org = *r.OrgId
	}
	return domain.FedPeer{
		ID: r.Id, Key: r.Key, Name: r.Name, OrgID: org,
		BaseURL: r.BaseUrl, JWKSURL: r.JwksUrl, Status: r.Status,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

// FedOutbox нь fed_outbox хүснэгтийн GORM model.
type FedOutbox struct {
	Id       string `gorm:"column:id;primaryKey"`
	PeerId   string `gorm:"column:peer_id"`
	Typ      string `gorm:"column:typ"`
	Envelope string `gorm:"column:envelope;type:text"`
	Status   string `gorm:"column:status"`
	Attempts int    `gorm:"column:attempts"`
}

func (FedOutbox) TableName() string { return "fed_outbox" }
