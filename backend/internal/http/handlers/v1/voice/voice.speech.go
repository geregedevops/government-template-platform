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

// Transcribe godoc
// @Summary      Дуу → бичвэр (STT)
// @Description  Base64 аудио хэрчмийг тэр хэл дээр нь бичвэрлэж буцаана (орчуулгагүй). Чатад дуугаар асуухад зориулсан.
// @Tags         voice
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  requests.VoiceTranscribeRequest  true  "Audio clip"
// @Success      200  {object}  v1.BaseResponse{data=responses.VoiceTranscribeResponse}
// @Failure      400  {object}  v1.BaseResponse
// @Failure      401  {object}  v1.BaseResponse
// @Failure      503  {object}  v1.BaseResponse
// @Router       /voice/transcribe [post]
func (h Handler) Transcribe(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	var req requests.VoiceTranscribeRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.WarnWithContext(c.Context(), "Transcribe: invalid request body", logger.Fields{
			"controller": "voice", "method": "Transcribe", "file": "voice.speech.go", "error": err.Error(),
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
	res, err := h.usecase.Transcribe(c.Context(), voiceuc.TranscribeRequest{
		UserID:   user.ID,
		Lang:     req.Lang,
		Audio:    audio,
		MimeType: req.MimeType,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "voice transcribed successfully",
		responses.VoiceTranscribeResponse{Text: res.Text})
}

// Speak godoc
// @Summary      Бичвэр → дуу (TTS)
// @Description  Бичвэрийг яриа болгож, тоглуулахад бэлэн WAV (base64) буцаана. Чатын хариуг чанга уншихад зориулсан.
// @Tags         voice
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body  requests.VoiceSpeakRequest  true  "Text to speak"
// @Success      200  {object}  v1.BaseResponse{data=responses.VoiceSpeakResponse}
// @Failure      400  {object}  v1.BaseResponse
// @Failure      401  {object}  v1.BaseResponse
// @Failure      503  {object}  v1.BaseResponse
// @Router       /voice/speak [post]
func (h Handler) Speak(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	var req requests.VoiceSpeakRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.WarnWithContext(c.Context(), "Speak: invalid request body", logger.Fields{
			"controller": "voice", "method": "Speak", "file": "voice.speech.go", "error": err.Error(),
		})
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	res, err := h.usecase.Speak(c.Context(), voiceuc.SpeakRequest{
		UserID: user.ID,
		Text:   req.Text,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "voice synthesized successfully",
		responses.VoiceSpeakResponse{
			AudioBase64: base64.StdEncoding.EncodeToString(res.AudioWAV),
			AudioMime:   "audio/wav",
		})
}
