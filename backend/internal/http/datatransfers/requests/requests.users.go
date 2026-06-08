// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package requests

// AdminCreateUserRequest нь admin-аар шинэ хэрэглэгч үүсгэх body. RoleID нь
// 1 (admin) эсвэл 2 (user).
type AdminCreateUserRequest struct {
	Username string `json:"username" validate:"required,min=2,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
	RoleID   int    `json:"role_id" validate:"required,min=1"`
}

// AdminUpdateRoleRequest нь хэрэглэгчийн эрхийг солих body (динамик role_id).
type AdminUpdateRoleRequest struct {
	RoleID int `json:"role_id" validate:"required,min=1"`
}

// AdminUpdateOrgRequest нь хэрэглэгчийг байгууллагад шилжүүлэх body.
type AdminUpdateOrgRequest struct {
	OrgID string `json:"org_id" validate:"required,uuid"`
}
