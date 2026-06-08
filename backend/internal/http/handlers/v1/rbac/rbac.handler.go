// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package rbac нь /rbac/* endpoint-уудыг үйлчилнэ — эрх (roles) CRUD, эрхийн
// каталог, мөн одоогийн хэрэглэгчийн эрхийг буцаах /rbac/me.
package rbac

import (
	"net/http"
	"strconv"

	rbacuc "geregetemplateai/internal/business/usecases/rbac"
	httpauth "geregetemplateai/internal/http/auth"
	"geregetemplateai/internal/http/datatransfers/requests"
	"geregetemplateai/internal/http/datatransfers/responses"
	v1 "geregetemplateai/internal/http/handlers/v1"
	"geregetemplateai/pkg/validators"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	usecase rbacuc.Usecase
}

func NewHandler(usecase rbacuc.Usecase) Handler {
	return Handler{usecase: usecase}
}

// MyPermissions godoc
// @Summary      Одоогийн хэрэглэгчийн эрхүүд
// @Tags         rbac
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  v1.BaseResponse{data=[]string}
// @Router       /rbac/me [get]
func (h Handler) MyPermissions(c fiber.Ctx) error {
	user, err := httpauth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	// RoleID байхгүй (хуучин токен) бол IsAdmin-аас гаргана (admin=1, бусад=2).
	roleID := user.RoleID
	if roleID == 0 {
		if user.IsAdmin {
			roleID = 1
		} else {
			roleID = 2
		}
	}
	perms, err := h.usecase.Resolve(c.Context(), roleID)
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	if perms == nil {
		perms = []string{}
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "permissions fetched successfully", perms)
}

// ListRoles godoc
// @Summary      Эрхүүдийг (permission-уудтай нь) жагсаах
// @Tags         rbac
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  v1.BaseResponse{data=[]responses.RoleResponse}
// @Router       /rbac/roles [get]
func (h Handler) ListRoles(c fiber.Ctx) error {
	res, err := h.usecase.ListRoles(c.Context())
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "roles fetched successfully", responses.ToRoleList(res))
}

// ListPermissions godoc
// @Summary      Эрхийн каталог
// @Tags         rbac
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  v1.BaseResponse{data=[]responses.PermissionResponse}
// @Router       /rbac/permissions [get]
func (h Handler) ListPermissions(c fiber.Ctx) error {
	res, err := h.usecase.ListPermissions(c.Context())
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "permissions fetched successfully", responses.ToPermissionList(res))
}

// CreateRole godoc
// @Summary      Эрх үүсгэх
// @Tags         rbac
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  requests.CreateRoleRequest  true  "Role"
// @Success      201  {object}  v1.BaseResponse{data=responses.RoleResponse}
// @Router       /rbac/roles [post]
func (h Handler) CreateRole(c fiber.Ctx) error {
	var req requests.CreateRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	role, err := h.usecase.CreateRole(c.Context(), rbacuc.CreateRoleRequest{
		Key: req.Key, Name: req.Name, Description: req.Description, Permissions: req.Permissions,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusCreated, "role created successfully", responses.FromRole(role))
}

// UpdateRole godoc
// @Summary      Эрх засах
// @Tags         rbac
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  int                         true  "Role ID"
// @Param        request  body  requests.UpdateRoleRequest  true  "Role"
// @Success      200  {object}  v1.BaseResponse{data=responses.RoleResponse}
// @Router       /rbac/roles/{id} [put]
func (h Handler) UpdateRole(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "role not found")
	}
	var req requests.UpdateRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	role, err := h.usecase.UpdateRole(c.Context(), rbacuc.UpdateRoleRequest{
		ID: id, Name: req.Name, Description: req.Description, Permissions: req.Permissions,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "role updated successfully", responses.FromRole(role))
}

// SetRolePermissions godoc
// @Summary      Эрхийн permission-уудыг тохируулах
// @Tags         rbac
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  int                                  true  "Role ID"
// @Param        request  body  requests.SetRolePermissionsRequest   true  "Permissions"
// @Success      200  {object}  v1.BaseResponse
// @Router       /rbac/roles/{id}/permissions [put]
func (h Handler) SetRolePermissions(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "role not found")
	}
	var req requests.SetRolePermissionsRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	if err := h.usecase.SetRolePermissions(c.Context(), id, req.Permissions); err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "role permissions updated successfully", nil)
}

// DeleteRole godoc
// @Summary      Эрх устгах
// @Tags         rbac
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "Role ID"
// @Success      200  {object}  v1.BaseResponse
// @Router       /rbac/roles/{id} [delete]
func (h Handler) DeleteRole(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "role not found")
	}
	if err := h.usecase.DeleteRole(c.Context(), id); err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "role deleted successfully", nil)
}
