// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package requests

// FedPeerCreateRequest нь шинэ peer (гишүүн node) бүртгэх хүсэлт.
type FedPeerCreateRequest struct {
	Key     string `json:"key" validate:"required,max=120"`
	Name    string `json:"name" validate:"max=200"`
	OrgID   string `json:"org_id" validate:"omitempty,uuid"`
	BaseURL string `json:"base_url" validate:"required,url"`
	JWKSURL string `json:"jwks_url" validate:"required,url"`
	Status  string `json:"status" validate:"omitempty,oneof=pending active suspended"`
}

// FedPeerUpdateRequest нь peer-ийн талбаруудыг засах хүсэлт.
type FedPeerUpdateRequest struct {
	Name    string `json:"name" validate:"max=200"`
	OrgID   string `json:"org_id" validate:"omitempty,uuid"`
	BaseURL string `json:"base_url" validate:"required,url"`
	JWKSURL string `json:"jwks_url" validate:"required,url"`
	Status  string `json:"status" validate:"required,oneof=pending active suspended"`
}
