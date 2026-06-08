// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package bpm

import (
	"net/http"

	bpmuc "geregetemplateai/internal/business/usecases/bpm"
	"geregetemplateai/internal/http/auth"
	"geregetemplateai/internal/http/datatransfers/requests"
	"geregetemplateai/internal/http/datatransfers/responses"
	v1 "geregetemplateai/internal/http/handlers/v1"
	"geregetemplateai/pkg/validators"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// ListForms godoc
// @Summary      Хуваалцсан формуудыг жагсаах
// @Tags         bpm
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  v1.BaseResponse{data=[]responses.BPMFormResponse}
// @Router       /bpm/forms [get]
func (h Handler) ListForms(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	res, err := h.usecase.ListForms(c.Context(), bpmuc.ListFormsRequest{
		UserID: user.ID,
		Offset: fiber.Query[int](c, "offset", 0),
		Limit:  fiber.Query[int](c, "limit", 100),
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "forms fetched successfully", responses.ToBPMFormList(res.Forms))
}

// CreateForm godoc
// @Summary      Хуваалцсан форм үүсгэх
// @Tags         bpm
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  requests.BPMSaveFormRequest  true  "Form"
// @Success      201  {object}  v1.BaseResponse{data=responses.BPMFormResponse}
// @Router       /bpm/forms [post]
func (h Handler) CreateForm(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	var req requests.BPMSaveFormRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	res, err := h.usecase.CreateForm(c.Context(), bpmuc.SaveFormRequest{
		UserID: user.ID,
		Name:   req.Name,
		Schema: string(req.Schema),
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusCreated, "form created successfully", responses.FromBPMForm(res.Form))
}

// UpdateForm godoc
// @Summary      Хуваалцсан форм засах (latest-wins)
// @Tags         bpm
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  string                       true  "Form ID"
// @Param        request  body  requests.BPMSaveFormRequest  true  "Form"
// @Success      200  {object}  v1.BaseResponse{data=responses.BPMFormResponse}
// @Router       /bpm/forms/{id} [put]
func (h Handler) UpdateForm(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid form id")
	}
	var req requests.BPMSaveFormRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	res, err := h.usecase.UpdateForm(c.Context(), bpmuc.SaveFormRequest{
		UserID: user.ID,
		ID:     id,
		Name:   req.Name,
		Schema: string(req.Schema),
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "form updated successfully", responses.FromBPMForm(res.Form))
}

// DeleteForm godoc
// @Summary      Хуваалцсан форм устгах
// @Tags         bpm
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Form ID"
// @Success      200  {object}  v1.BaseResponse
// @Router       /bpm/forms/{id} [delete]
func (h Handler) DeleteForm(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid form id")
	}
	if err := h.usecase.DeleteForm(c.Context(), bpmuc.GetFormRequest{UserID: user.ID, ID: id}); err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "form deleted successfully", nil)
}
