// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package voice

import (
	"encoding/base64"
	"net/http"

	voiceuc "geregetemplateai/internal/business/usecases/voice"
	"geregetemplateai/internal/http/auth"
	"geregetemplateai/internal/http/datatransfers/requests"
	"geregetemplateai/internal/http/datatransfers/responses"
	v1 "geregetemplateai/internal/http/handlers/v1"
	"geregetemplateai/pkg/logger"
	"geregetemplateai/pkg/validators"
	"github.com/gofiber/fiber/v3"
)

// Translate godoc
// @Summary      Дуу хоолойн орчуулга (MN↔EN)
// @Description  Base64 аудио хэрчмийг хүлээн авч, бичвэрлэж (STT), нөгөө хэл рүү орчуулж, орчуулгыг яриа болгож (TTS) буцаана. Хариунд эх бичвэр, орчуулга болон тоглуулахад бэлэн WAV (base64) орно.
// @Tags         voice
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  requests.VoiceTranslateRequest  true  "Voice clip"
// @Success      200  {object}  v1.BaseResponse{data=responses.VoiceTranslateResponse}
// @Failure      400  {object}  v1.BaseResponse  "Malformed body / audio"
// @Failure      401  {object}  v1.BaseResponse  "Unauthenticated"
// @Failure      403  {object}  v1.BaseResponse  "Daily limit exceeded"
// @Failure      422  {object}  v1.BaseResponse  "Validation error"
// @Failure      503  {object}  v1.BaseResponse  "Voice service not configured"
// @Router       /voice/translate [post]
func (h Handler) Translate(c fiber.Ctx) error {
	const (
		controllerName = "voice"
		funcName       = "Translate"
		fileName       = "voice.translate.go"
	)
	ctx := c.Context()

	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}

	var req requests.VoiceTranslateRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.WarnWithContext(ctx, "Translate: invalid request body", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
		})
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}

	audio, err := base64.StdEncoding.DecodeString(req.AudioBase64)
	if err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "audio is not valid base64")
	}

	res, err := h.usecase.Translate(ctx, voiceuc.TranslateRequest{
		UserID:     user.ID,
		SourceLang: req.SourceLang,
		Audio:      audio,
		MimeType:   req.MimeType,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}

	return v1.NewSuccessResponse(c, http.StatusOK, "voice translated successfully",
		responses.VoiceTranslateResponse{
			Id:             res.Translation.ID,
			SourceLang:     res.Translation.SourceLang,
			TargetLang:     res.Translation.TargetLang,
			SourceText:     res.Translation.SourceText,
			TranslatedText: res.Translation.TranslatedText,
			AudioBase64:    base64.StdEncoding.EncodeToString(res.AudioWAV),
			AudioMime:      "audio/wav",
			CreatedAt:      res.Translation.CreatedAt,
		})
}
