// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package records

import (
	"time"

	"geregetemplateai/internal/business/domain"
)

// Roles нь roles хүснэгтийн GORM model.
type Roles struct {
	Id          int        `gorm:"column:id;primaryKey"`
	Key         string     `gorm:"column:key"`
	Name        string     `gorm:"column:name"`
	Description string     `gorm:"column:description"`
	IsSystem    bool       `gorm:"column:is_system"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at"`
}

func (Roles) TableName() string { return "roles" }

func (r Roles) ToV1Domain() domain.Role {
	return domain.Role{
		ID:          r.Id,
		Key:         r.Key,
		Name:        r.Name,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// Permissions нь permissions каталогийн GORM model.
type Permissions struct {
	Key      string `gorm:"column:key;primaryKey"`
	Label    string `gorm:"column:label"`
	Category string `gorm:"column:category"`
}

func (Permissions) TableName() string { return "permissions" }

func (p Permissions) ToV1Domain() domain.Permission {
	return domain.Permission{Key: p.Key, Label: p.Label, Category: p.Category}
}

// RolePermissions нь role↔permission холбоосын GORM model.
type RolePermissions struct {
	RoleId        int    `gorm:"column:role_id;primaryKey"`
	PermissionKey string `gorm:"column:permission_key;primaryKey"`
}

func (RolePermissions) TableName() string { return "role_permissions" }
