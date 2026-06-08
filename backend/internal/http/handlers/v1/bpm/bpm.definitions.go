// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package bpm

import (
	"net/http"
	"strings"

	bpmuc "geregetemplateai/internal/business/usecases/bpm"
	"geregetemplateai/internal/http/auth"
	"geregetemplateai/internal/http/datatransfers/requests"
	"geregetemplateai/internal/http/datatransfers/responses"
	v1 "geregetemplateai/internal/http/handlers/v1"
	"geregetemplateai/pkg/validators"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// GenerateProcess godoc
// @Summary      Текст тайлбараас AI-аар BPM процесс үүсгэх
// @Tags         bpm
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  requests.BPMGenerateRequest  true  "Description"
// @Success      201  {object}  v1.BaseResponse{data=responses.BPMProcessResponse}
// @Failure      401  {object}  v1.BaseResponse
// @Failure      422  {object}  v1.BaseResponse
// @Failure      503  {object}  v1.BaseResponse
// @Router       /bpm/generate [post]
func (h Handler) GenerateProcess(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	var req requests.BPMGenerateRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	lang := "en"
	if strings.HasPrefix(strings.ToLower(c.Get("Accept-Language")), "mn") {
		lang = "mn"
	}
	res, err := h.usecase.GenerateProcess(c.Context(), bpmuc.GenerateProcessRequest{
		UserID:      user.ID,
		OrgID:       user.OrgID,
		Description: req.Description,
		Lang:        lang,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusCreated, "process generated successfully",
		responses.FromBPMProcess(res.Process))
}

// CreateProcess godoc
// @Summary      Шинэ BPM процесс үүсгэх
// @Tags         bpm
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  requests.BPMSaveProcessRequest  true  "Process"
// @Success      201  {object}  v1.BaseResponse{data=responses.BPMProcessResponse}
// @Failure      401  {object}  v1.BaseResponse
// @Failure      422  {object}  v1.BaseResponse
// @Router       /bpm/processes [post]
func (h Handler) CreateProcess(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	var req requests.BPMSaveProcessRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	res, err := h.usecase.CreateProcess(c.Context(), bpmuc.SaveProcessRequest{
		UserID:      user.ID,
		OrgID:       user.OrgID,
		Name:        req.Name,
		Description: req.Description,
		Definition:  req.BPMN,
		Status:      req.Status,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusCreated, "process created successfully",
		responses.FromBPMProcess(res.Process))
}

// UpdateProcess godoc
// @Summary      BPM процесс шинэчлэх
// @Tags         bpm
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  string                          true  "Process ID (uuid)"
// @Param        request  body  requests.BPMSaveProcessRequest  true  "Process"
// @Success      200  {object}  v1.BaseResponse{data=responses.BPMProcessResponse}
// @Failure      401  {object}  v1.BaseResponse
// @Failure      404  {object}  v1.BaseResponse
// @Failure      422  {object}  v1.BaseResponse
// @Router       /bpm/processes/{id} [put]
func (h Handler) UpdateProcess(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "process not found")
	}
	var req requests.BPMSaveProcessRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	res, err := h.usecase.UpdateProcess(c.Context(), bpmuc.UpdateProcessRequest{
		UserID:      user.ID,
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Definition:  req.BPMN,
		Status:      req.Status,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "process updated successfully",
		responses.FromBPMProcess(res.Process))
}

// GetProcess godoc
// @Summary      BPM процессыг ID-ээр авах
// @Tags         bpm
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Process ID (uuid)"
// @Success      200  {object}  v1.BaseResponse{data=responses.BPMProcessResponse}
// @Failure      401  {object}  v1.BaseResponse
// @Failure      404  {object}  v1.BaseResponse
// @Router       /bpm/processes/{id} [get]
func (h Handler) GetProcess(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "process not found")
	}
	res, err := h.usecase.GetProcess(c.Context(), bpmuc.GetProcessRequest{UserID: user.ID, ID: id})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "process fetched successfully",
		responses.FromBPMProcess(res.Process))
}

// ListProcesses godoc
// @Summary      Хэрэглэгчийн BPM процессуудыг жагсаах
// @Tags         bpm
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  v1.BaseResponse{data=[]responses.BPMProcessResponse}
// @Failure      401  {object}  v1.BaseResponse
// @Router       /bpm/processes [get]
func (h Handler) ListProcesses(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	res, err := h.usecase.ListProcesses(c.Context(), bpmuc.ListProcessesRequest{
		UserID: user.ID,
		Offset: fiber.Query[int](c, "offset", 0),
		Limit:  fiber.Query[int](c, "limit", 20),
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "processes fetched successfully",
		responses.ToBPMProcessList(res.Processes))
}

// DeleteProcess godoc
// @Summary      BPM процесс устгах
// @Tags         bpm
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Process ID (uuid)"
// @Success      200  {object}  v1.BaseResponse
// @Failure      401  {object}  v1.BaseResponse
// @Failure      404  {object}  v1.BaseResponse
// @Router       /bpm/processes/{id} [delete]
func (h Handler) DeleteProcess(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "process not found")
	}
	if err := h.usecase.DeleteProcess(c.Context(), bpmuc.GetProcessRequest{UserID: user.ID, ID: id}); err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "process deleted successfully", nil)
}
