// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package ai

import (
	"net/http"

	aiuc "geregetemplateai/internal/business/usecases/ai"
	"geregetemplateai/internal/http/auth"
	"geregetemplateai/internal/http/datatransfers/responses"
	v1 "geregetemplateai/internal/http/handlers/v1"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// ListConversations godoc
// @Summary      Хэрэглэгчийн AI харилцан яриануудыг жагсаах
// @Tags         ai
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  v1.BaseResponse{data=[]responses.AIConversationResponse}
// @Failure      401  {object}  v1.BaseResponse
// @Router       /ai/conversations [get]
func (h Handler) ListConversations(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	res, err := h.usecase.ListConversations(c.Context(), aiuc.ListConversationsRequest{
		UserID: user.ID,
		Offset: fiber.Query[int](c, "offset", 0),
		Limit:  fiber.Query[int](c, "limit", 20),
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "conversations fetched successfully",
		responses.ToAIConversationList(res.Conversations))
}

// GetMessages godoc
// @Summary      Нэг харилцан ярианы мессежүүдийг авах
// @Tags         ai
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Conversation ID (uuid)"
// @Success      200  {object}  v1.BaseResponse{data=[]responses.AIMessageResponse}
// @Failure      401  {object}  v1.BaseResponse
// @Failure      404  {object}  v1.BaseResponse
// @Router       /ai/conversations/{id}/messages [get]
func (h Handler) GetMessages(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	convID := c.Params("id")
	if _, err := uuid.Parse(convID); err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "conversation not found")
	}
	res, err := h.usecase.GetMessages(c.Context(), aiuc.GetMessagesRequest{
		UserID:         user.ID,
		ConversationID: convID,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "messages fetched successfully",
		responses.ToAIMessageList(res.Messages))
}
