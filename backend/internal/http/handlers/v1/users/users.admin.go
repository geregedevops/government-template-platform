// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package users

import (
	"net/http"

	"geregetemplateai/internal/business/usecases/users"
	httpauth "geregetemplateai/internal/http/auth"
	"geregetemplateai/internal/http/datatransfers/requests"
	"geregetemplateai/internal/http/datatransfers/responses"
	v1 "geregetemplateai/internal/http/handlers/v1"
	"geregetemplateai/pkg/validators"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// ListUsers godoc
// @Summary      Бүх хэрэглэгчдийг жагсаах (admin)
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  v1.BaseResponse{data=[]responses.UserResponse}
// @Failure      401  {object}  v1.BaseResponse
// @Failure      403  {object}  v1.BaseResponse
// @Router       /users [get]
func (h Handler) ListUsers(c fiber.Ctx) error {
	res, err := h.usecase.ListUsers(c.Context(), users.ListUsersRequest{
		Offset: fiber.Query[int](c, "offset", 0),
		Limit:  fiber.Query[int](c, "limit", 200),
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "users fetched successfully", responses.ToUserList(res.Users))
}

// CreateUser godoc
// @Summary      Шинэ хэрэглэгч үүсгэх (admin)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  requests.AdminCreateUserRequest  true  "User"
// @Success      201  {object}  v1.BaseResponse{data=responses.UserResponse}
// @Router       /users [post]
func (h Handler) CreateUser(c fiber.Ctx) error {
	var req requests.AdminCreateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	res, err := h.usecase.AdminCreateUser(c.Context(), users.AdminCreateUserRequest{
		Username: req.Username, Email: req.Email, Password: req.Password, RoleID: req.RoleID,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusCreated, "user created successfully", responses.FromV1Domain(res.User))
}

// UpdateUserRole godoc
// @Summary      Хэрэглэгчийн эрхийг солих (admin)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  string                            true  "User ID (uuid)"
// @Param        request  body  requests.AdminUpdateRoleRequest   true  "Role"
// @Success      200  {object}  v1.BaseResponse
// @Router       /users/{id}/role [patch]
func (h Handler) UpdateUserRole(c fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "user not found")
	}
	var req requests.AdminUpdateRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	if err := h.usecase.UpdateRole(c.Context(), users.UpdateRoleRequest{ID: id, RoleID: req.RoleID}); err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "role updated successfully", nil)
}

// UpdateUserOrg godoc
// @Summary      Хэрэглэгчийг байгууллагад шилжүүлэх (admin)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  string                          true  "User ID (uuid)"
// @Param        request  body  requests.AdminUpdateOrgRequest  true  "Org"
// @Success      200  {object}  v1.BaseResponse
// @Router       /users/{id}/org [patch]
func (h Handler) UpdateUserOrg(c fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "user not found")
	}
	var req requests.AdminUpdateOrgRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	if err := h.usecase.UpdateOrg(c.Context(), users.UpdateOrgRequest{ID: id, OrgID: req.OrgID}); err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "user organization updated successfully", nil)
}

// DeleteUser godoc
// @Summary      Хэрэглэгчийг устгах (admin)
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "User ID (uuid)"
// @Success      200  {object}  v1.BaseResponse
// @Router       /users/{id} [delete]
func (h Handler) DeleteUser(c fiber.Ctx) error {
	actor, err := httpauth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "user not found")
	}
	// Admin өөрийгөө устгахаас сэргийлнэ (өөрийгөө түгжихгүйн тулд).
	if id == actor.ID {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "cannot delete yourself")
	}
	if err := h.usecase.DeleteUser(c.Context(), users.DeleteUserRequest{ID: id}); err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "user deleted successfully", nil)
}
