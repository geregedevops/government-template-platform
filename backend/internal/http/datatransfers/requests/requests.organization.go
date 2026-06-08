// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package requests

// OrgCreateRequest нь parent доор шинэ байгууллага үүсгэх body.
type OrgCreateRequest struct {
	ParentID string `json:"parent_id" validate:"required"`
	Name     string `json:"name" validate:"required,min=1,max=200"`
	Kind     string `json:"kind" validate:"omitempty,oneof=ministry agency soe"`
}

// OrgUpdateRequest нь нэр/төрлийг шинэчлэх body (hierarchy хөдөлгөхгүй).
type OrgUpdateRequest struct {
	Name string `json:"name" validate:"required,min=1,max=200"`
	Kind string `json:"kind" validate:"omitempty,oneof=root ministry agency soe"`
}
