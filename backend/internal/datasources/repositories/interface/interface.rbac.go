// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package _interface

import (
	"context"

	"geregetemplateai/internal/business/domain"
)

// RBACRepository нь динамик эрх (roles) болон role↔permission холбоосыг
// удирдана. permissions нь код дотор тодорхойлогдсон каталог тул зөвхөн
// уншина.
type RBACRepository interface {
	ListRoles(ctx context.Context) ([]domain.Role, error)
	GetRole(ctx context.Context, id int) (domain.Role, error)
	CreateRole(ctx context.Context, in *domain.Role) (domain.Role, error)
	UpdateRole(ctx context.Context, in *domain.Role) (domain.Role, error)
	DeleteRole(ctx context.Context, id int) error
	// CountUsersWithRole нь тухайн эрхтэй (идэвхтэй) хэрэглэгчийн тоог буцаана —
	// ашиглагдаж буй эрхийг устгахаас сэргийлнэ.
	CountUsersWithRole(ctx context.Context, id int) (int64, error)

	ListPermissions(ctx context.Context) ([]domain.Permission, error)
	GetRolePermissions(ctx context.Context, roleID int) ([]string, error)
	// SetRolePermissions нь нэг role-ийн эрхүүдийг бүхэлд нь солино (replace).
	SetRolePermissions(ctx context.Context, roleID int, keys []string) error
}
