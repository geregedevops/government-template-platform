// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package organization нь байгууллагын модны HTTP handler-ууд.
package organization

import (
	"net/http"

	orguc "geregetemplateai/internal/business/usecases/organization"
	"geregetemplateai/internal/http/datatransfers/requests"
	"geregetemplateai/internal/http/datatransfers/responses"
	v1 "geregetemplateai/internal/http/handlers/v1"
	"geregetemplateai/pkg/validators"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase orguc.Usecase
}

func NewHandler(usecase orguc.Usecase) Handler {
	return Handler{usecase: usecase}
}

// List godoc
// @Summary      Байгууллагын модыг жагсаах
// @Tags         organizations
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  v1.BaseResponse{data=[]responses.OrganizationResponse}
// @Router       /organizations [get]
func (h Handler) List(c fiber.Ctx) error {
	res, err := h.usecase.List(c.Context())
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "organizations fetched successfully", responses.ToOrganizationList(res.Orgs))
}

// Create godoc
// @Summary      Байгууллага үүсгэх (parent доор)
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  requests.OrgCreateRequest  true  "Organization"
// @Success      201  {object}  v1.BaseResponse{data=responses.OrganizationResponse}
// @Router       /organizations [post]
func (h Handler) Create(c fiber.Ctx) error {
	var req requests.OrgCreateRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	res, err := h.usecase.Create(c.Context(), orguc.SaveRequest{
		ParentID: req.ParentID, Name: req.Name, Kind: req.Kind,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusCreated, "organization created successfully", responses.FromOrganization(res.Org))
}

// Update godoc
// @Summary      Байгууллагын нэр/төрөл засах
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  string                     true  "Org ID"
// @Param        request  body  requests.OrgUpdateRequest  true  "Organization"
// @Success      200  {object}  v1.BaseResponse{data=responses.OrganizationResponse}
// @Router       /organizations/{id} [put]
func (h Handler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid organization id")
	}
	var req requests.OrgUpdateRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	res, err := h.usecase.Update(c.Context(), orguc.SaveRequest{ID: id, Name: req.Name, Kind: req.Kind})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "organization updated successfully", responses.FromOrganization(res.Org))
}

// Delete godoc
// @Summary      Байгууллага устгах (дэд мод бүхэлдээ)
// @Tags         organizations
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Org ID"
// @Success      200  {object}  v1.BaseResponse
// @Router       /organizations/{id} [delete]
func (h Handler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid organization id")
	}
	if err := h.usecase.Delete(c.Context(), id); err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "organization deleted successfully", nil)
}
