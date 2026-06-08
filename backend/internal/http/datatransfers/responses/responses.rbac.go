// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package responses

import (
	"time"

	"geregetemplateai/internal/business/domain"
	"geregetemplateai/internal/business/usecases/rbac"
)

type RoleResponse struct {
	Id          int        `json:"id"`
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	IsSystem    bool       `json:"is_system"`
	Permissions []string   `json:"permissions"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

func FromRoleWithPerms(r rbac.RoleWithPerms) RoleResponse {
	perms := r.Permissions
	if perms == nil {
		perms = []string{}
	}
	return RoleResponse{
		Id:          r.Role.ID,
		Key:         r.Role.Key,
		Name:        r.Role.Name,
		Description: r.Role.Description,
		IsSystem:    r.Role.IsSystem,
		Permissions: perms,
		CreatedAt:   r.Role.CreatedAt,
		UpdatedAt:   r.Role.UpdatedAt,
	}
}

func ToRoleList(items []rbac.RoleWithPerms) []RoleResponse {
	out := make([]RoleResponse, 0, len(items))
	for _, r := range items {
		out = append(out, FromRoleWithPerms(r))
	}
	return out
}

func FromRole(r domain.Role) RoleResponse {
	return RoleResponse{
		Id: r.ID, Key: r.Key, Name: r.Name, Description: r.Description,
		IsSystem: r.IsSystem, Permissions: []string{}, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

type PermissionResponse struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Category string `json:"category"`
}

func ToPermissionList(items []domain.Permission) []PermissionResponse {
	out := make([]PermissionResponse, 0, len(items))
	for _, p := range items {
		out = append(out, PermissionResponse{Key: p.Key, Label: p.Label, Category: p.Category})
	}
	return out
}
