// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package responses

import (
	"time"

	"geregetemplateai/internal/business/domain"
)

// OrganizationResponse нь байгууллагын модны нэг зангилаа. ParentId хоосон бол root.
type OrganizationResponse struct {
	Id        string     `json:"id"`
	ParentId  string     `json:"parent_id"`
	Path      string     `json:"path"`
	Name      string     `json:"name"`
	Kind      string     `json:"kind"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func FromOrganization(o domain.Organization) OrganizationResponse {
	return OrganizationResponse{
		Id:        o.ID,
		ParentId:  o.ParentID,
		Path:      o.Path,
		Name:      o.Name,
		Kind:      o.Kind,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

func ToOrganizationList(items []domain.Organization) []OrganizationResponse {
	out := make([]OrganizationResponse, 0, len(items))
	for _, o := range items {
		out = append(out, FromOrganization(o))
	}
	return out
}
