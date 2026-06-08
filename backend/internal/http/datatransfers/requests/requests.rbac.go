// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package requests

// CreateRoleRequest нь шинэ эрх үүсгэх body. Key хоосон бол Name-ээс үүснэ.
type CreateRoleRequest struct {
	Key         string   `json:"key" validate:"omitempty,max=50"`
	Name        string   `json:"name" validate:"required,min=2,max=100"`
	Description string   `json:"description" validate:"omitempty,max=500"`
	Permissions []string `json:"permissions" validate:"omitempty,dive,max=64"`
}

// UpdateRoleRequest нь эрхийн нэр/тайлбар + (заавал биш) эрхийн жагсаалтыг солино.
type UpdateRoleRequest struct {
	Name        string   `json:"name" validate:"required,min=2,max=100"`
	Description string   `json:"description" validate:"omitempty,max=500"`
	Permissions []string `json:"permissions" validate:"omitempty,dive,max=64"`
}

// SetRolePermissionsRequest нь нэг эрхийн permission-уудыг бүхэлд нь солино.
type SetRolePermissionsRequest struct {
	Permissions []string `json:"permissions" validate:"dive,max=64"`
}
