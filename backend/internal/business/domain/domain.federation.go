// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package domain

import "time"

// Peer-ийн төлөв.
const (
	FedPeerPending   = "pending"
	FedPeerActive    = "active"
	FedPeerSuspended = "suspended"
)

// Outbox мессежийн төлөв.
const (
	FedOutboxPending = "pending"
	FedOutboxSent    = "sent"
	FedOutboxDead    = "dead"
)

// FedPeer нь федерацийн гишүүн node (platform registry-ийн бичлэг).
type FedPeer struct {
	ID        string
	Key       string // node-ийн ялгах нэр (envelope iss/aud)
	Name      string
	OrgID     string
	BaseURL   string // нийтийн суурь URL (inbound = base + /api/fed/inbound)
	JWKSURL   string // гарын үсгийн нийтийн түлхүүрийн эх
	Status    string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// FedOutboxMessage нь гарах гарын үсэгтэй мессежийн дараалал дахь бичлэг.
type FedOutboxMessage struct {
	ID       string
	PeerID   string
	Typ      string
	Envelope string
	Status   string
	Attempts int
}
