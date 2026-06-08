// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package ai

import (
	"net/http"

	aiuc "geregetemplateai/internal/business/usecases/ai"
	"geregetemplateai/internal/http/auth"
	"geregetemplateai/internal/http/datatransfers/requests"
	"geregetemplateai/internal/http/datatransfers/responses"
	v1 "geregetemplateai/internal/http/handlers/v1"
	"geregetemplateai/pkg/validators"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// ListKnowledge godoc
// @Summary      AI мэдлэгийн бичлэгүүдийг жагсаах
// @Tags         ai
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  v1.BaseResponse{data=[]responses.AIKnowledgeResponse}
// @Router       /ai/knowledge [get]
func (h Handler) ListKnowledge(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	res, err := h.usecase.ListKnowledge(c.Context(), aiuc.ListKnowledgeRequest{
		UserID: user.ID,
		Offset: fiber.Query[int](c, "offset", 0),
		Limit:  fiber.Query[int](c, "limit", 100),
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "knowledge fetched successfully",
		responses.ToAIKnowledgeList(res.Items))
}

// ListAllKnowledge godoc
// @Summary      Бүх хэрэглэгчийн AI мэдлэг (admin)
// @Tags         ai
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  v1.BaseResponse{data=[]responses.AIKnowledgeResponse}
// @Router       /ai/knowledge/all [get]
func (h Handler) ListAllKnowledge(c fiber.Ctx) error {
	res, err := h.usecase.ListAllKnowledge(c.Context(), aiuc.ListKnowledgeRequest{
		Offset: fiber.Query[int](c, "offset", 0),
		Limit:  fiber.Query[int](c, "limit", 100),
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "knowledge fetched successfully",
		responses.ToAIKnowledgeList(res.Items))
}

// CreateKnowledge godoc
// @Summary      AI мэдлэгийн бичлэг үүсгэх
// @Tags         ai
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  requests.AIKnowledgeRequest  true  "Knowledge"
// @Success      201  {object}  v1.BaseResponse{data=responses.AIKnowledgeResponse}
// @Router       /ai/knowledge [post]
func (h Handler) CreateKnowledge(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	var req requests.AIKnowledgeRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	res, err := h.usecase.CreateKnowledge(c.Context(), aiuc.SaveKnowledgeRequest{
		UserID: user.ID, Title: req.Title, Content: req.Content,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusCreated, "knowledge created successfully",
		responses.FromAIKnowledge(res.Knowledge))
}

// UpdateKnowledge godoc
// @Summary      AI мэдлэгийн бичлэг засах
// @Tags         ai
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  string                       true  "Knowledge ID (uuid)"
// @Param        request  body  requests.AIKnowledgeRequest  true  "Knowledge"
// @Success      200  {object}  v1.BaseResponse{data=responses.AIKnowledgeResponse}
// @Router       /ai/knowledge/{id} [put]
func (h Handler) UpdateKnowledge(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "knowledge not found")
	}
	var req requests.AIKnowledgeRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	res, err := h.usecase.UpdateKnowledge(c.Context(), aiuc.UpdateKnowledgeRequest{
		UserID: user.ID, ID: id, Title: req.Title, Content: req.Content,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "knowledge updated successfully",
		responses.FromAIKnowledge(res.Knowledge))
}

// DeleteKnowledge godoc
// @Summary      AI мэдлэгийн бичлэг устгах
// @Tags         ai
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Knowledge ID (uuid)"
// @Success      200  {object}  v1.BaseResponse
// @Router       /ai/knowledge/{id} [delete]
func (h Handler) DeleteKnowledge(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "knowledge not found")
	}
	if err := h.usecase.DeleteKnowledge(c.Context(), aiuc.DeleteKnowledgeRequest{UserID: user.ID, ID: id}); err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "knowledge deleted successfully", nil)
}
