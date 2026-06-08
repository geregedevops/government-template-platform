// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package voice

import (
	"net/http"

	voiceuc "geregetemplateai/internal/business/usecases/voice"
	"geregetemplateai/internal/http/auth"
	"geregetemplateai/internal/http/datatransfers/responses"
	v1 "geregetemplateai/internal/http/handlers/v1"
	"github.com/gofiber/fiber/v3"
)

// ListTranslations godoc
// @Summary      Дуу хоолойн орчуулгын түүх
// @Tags         voice
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  v1.BaseResponse{data=[]responses.VoiceTranslationResponse}
// @Failure      401  {object}  v1.BaseResponse
// @Router       /voice/history [get]
func (h Handler) ListTranslations(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	res, err := h.usecase.ListTranslations(c.Context(), voiceuc.ListTranslationsRequest{
		UserID: user.ID,
		Offset: fiber.Query[int](c, "offset", 0),
		Limit:  fiber.Query[int](c, "limit", 20),
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "voice history fetched successfully",
		responses.ToVoiceTranslationList(res.Translations))
}
