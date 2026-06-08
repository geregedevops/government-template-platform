// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"geregetemplateai/internal/apperror"
	aiuc "geregetemplateai/internal/business/usecases/ai"
	"geregetemplateai/internal/http/auth"
	"geregetemplateai/internal/http/datatransfers/requests"
	v1 "geregetemplateai/internal/http/handlers/v1"
	"geregetemplateai/internal/i18n"
	"geregetemplateai/pkg/logger"
	"geregetemplateai/pkg/validators"
	"github.com/gofiber/fiber/v3"
)

// Chat godoc
// @Summary      AI туслахтай streaming чат (SSE)
// @Description  Хэрэглэгчийн мессежийг харилцан ярианд нэмж, Claude-аас text/event-stream хэлбэрээр хариу урсгана. Event-ууд: meta (conversation_id), delta (текст хэсэг), done (message_id + токен зарцуулалт), error.
// @Tags         ai
// @Accept       json
// @Produce      text/event-stream
// @Security     BearerAuth
// @Param        request  body  requests.AIChatRequest  true  "Chat message"
// @Success      200  {string}  string  "SSE stream"
// @Failure      400  {object}  v1.BaseResponse  "Malformed body"
// @Failure      401  {object}  v1.BaseResponse  "Unauthenticated"
// @Failure      403  {object}  v1.BaseResponse  "Daily limit exceeded"
// @Failure      422  {object}  v1.BaseResponse  "Validation error"
// @Failure      503  {object}  v1.BaseResponse  "AI not configured"
// @Router       /ai/chat [post]
func (h Handler) Chat(c fiber.Ctx) error {
	const (
		controllerName = "ai"
		funcName       = "Chat"
		fileName       = "ai.chat.go"
	)
	ctx := c.Context()

	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}

	var req requests.AIChatRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.WarnWithContext(ctx, "Chat: invalid request body", logger.Fields{
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

	lang := i18n.FromContext(ctx)

	// SSE толгойнууд. X-Accel-Buffering нь nginx зэрэг proxy-ийн буферийг
	// унтрааж, delta-ууд клиентэд шууд хүрэхийг баталгаажуулна.
	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Set(fiber.HeaderCacheControl, "no-cache")
	c.Set("X-Accel-Buffering", "no")

	// Stream writer нь handler буцсаны ДАРАА ажилладаг тул:
	//   - fiber.Ctx-д дахин хүрч болохгүй (request дахин ашиглагдсан байж
	//     болно) — хэрэгтэй бүх утгыг одоо хувьсагчид хуулна;
	//   - request context нь TimeoutMiddleware-ийн 30с deadline-тай бөгөөд
	//     handler буцмагц cancel хийгддэг — WithoutCancel нь утгуудыг
	//     (RLS identity, request id, locale) хадгалж цуцлалтыг нь хаяна,
	//     дараа нь streaming-д зориулсан өөрийн deadline тавина.
	baseCtx := context.WithoutCancel(ctx)
	userID := user.ID
	chatReq := aiuc.ChatRequest{
		UserID:         userID,
		ConversationID: req.ConversationID,
		Message:        req.Message,
		Lang:           string(lang),
	}
	streamTimeout := h.streamTimeout

	return c.SendStreamWriter(func(w *bufio.Writer) {
		streamCtx, cancel := context.WithTimeout(baseCtx, streamTimeout)
		defer cancel()

		write := func(event string, payload any) error {
			data, err := json.Marshal(payload)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data); err != nil {
				return err
			}
			return w.Flush()
		}

		metaSent := false
		res, chatErr := h.usecase.Chat(streamCtx, chatReq, func(delta string) error {
			// Эхний delta-аас өмнө conversation_id аль хэдийн үүссэн байдаг
			// ч usecase нь callback дотор meta өгдөггүй тул эхний delta дээр
			// зөвхөн delta event бичээд, done event-д бүх ID-г өгнө.
			metaSent = true
			return write("delta", map[string]string{"delta": delta})
		})

		if chatErr != nil {
			// Алдааны мессежийг хэрэглэгчийн хэлээр, дотоод шалтгааныг логт.
			msg := "internal server error"
			var domErr *apperror.DomainError
			if errors.As(chatErr, &domErr) && domErr.Type != apperror.ErrTypeInternal {
				msg = domErr.Message
			}
			fields := logger.Fields{
				"controller": controllerName,
				"method":     funcName,
				"file":       fileName,
				"user_id":    userID,
				"error":      chatErr.Error(),
			}
			if errors.As(chatErr, &domErr) && domErr.Cause != nil {
				fields["cause"] = domErr.Cause.Error()
			}
			logger.ErrorWithContext(streamCtx, "Chat stream failed", fields)
			_ = write("error", map[string]any{
				"message": i18n.T(lang, msg),
				"partial": metaSent,
			})
			return
		}

		_ = write("done", map[string]any{
			"conversation_id": res.Conversation.ID,
			"message_id":      res.Reply.ID,
			"input_tokens":    res.InputTokens,
			"output_tokens":   res.OutputTokens,
		})
	})
}
