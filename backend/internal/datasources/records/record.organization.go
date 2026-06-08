// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package records

import (
	"time"

	"geregetemplateai/internal/business/domain"
)

// Organizations нь organizations хүснэгтийн GORM model. ParentId нь root үед NULL
// тул *string. Path нь ltree-г текстээр уншина.
type Organizations struct {
	Id        string     `gorm:"column:id;primaryKey"`
	ParentId  *string    `gorm:"column:parent_id"`
	Path      string     `gorm:"column:path"`
	Name      string     `gorm:"column:name"`
	Kind      string     `gorm:"column:kind"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
}

func (Organizations) TableName() string { return "organizations" }

func (r Organizations) ToV1Domain() domain.Organization {
	parent := ""
	if r.ParentId != nil {
		parent = *r.ParentId
	}
	return domain.Organization{
		ID:        r.Id,
		ParentID:  parent,
		Path:      r.Path,
		Name:      r.Name,
		Kind:      r.Kind,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
