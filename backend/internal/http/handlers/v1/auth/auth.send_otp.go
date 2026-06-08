// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package auth

import (
	"net/http"

	authuc "geregetemplateai/internal/business/usecases/auth"
	"geregetemplateai/internal/http/datatransfers/requests"
	v1 "geregetemplateai/internal/http/handlers/v1"
	"geregetemplateai/pkg/audit"
	"geregetemplateai/pkg/logger"
	"geregetemplateai/pkg/validators"
	"github.com/gofiber/fiber/v3"
)

// SendOTP godoc
// @Summary      Хэрэглэгчийн email рүү OTP код илгээх
// @Description  6 оронтой OTP үүсгэж, TTL-тэйгээр Redis-д хадгалж, async mailer-ээр и-мэйлийг дараалалд оруулна. HTTP хариу нь бодит SMTP хүргэлт дээр биш, дараалалд оруулах үед буцдаг.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      requests.SendOTPRequest  true  "Email to send OTP to"
// @Success      200      {object}  v1.BaseResponse  "OTP enqueued"
// @Failure      404      {object}  v1.BaseResponse  "Email not registered"
// @Failure      400      {object}  v1.BaseResponse  "Account already activated"
// @Failure      422      {object}  v1.BaseResponse  "Validation error"
// @Failure      500      {object}  v1.BaseResponse  "Failed to enqueue mail"
// @Router       /auth/send-otp [post]
func (h Handler) SendOTP(c fiber.Ctx) error {
	const (
		controllerName = "auth"
		funcName       = "SendOTP"
		fileName       = "auth.send_otp.go"
	)
	ctx := c.Context()
	var req requests.SendOTPRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.WarnWithContext(ctx, "SendOTP: invalid request body", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
		})
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		logger.WarnWithContext(ctx, "SendOTP: validation error", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"request": logger.Fields{
				"email": req.Email,
			},
		})
		return v1.RespondWithError(c, err)
	}

	if err := h.usecase.SendOTP(ctx, authuc.SendOTPRequest{Email: req.Email}); err != nil {
		logger.ErrorWithContext(ctx, "SendOTP failed in controller", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"email":      req.Email,
		})
		return v1.RespondWithError(c, err)
	}

	ev := auditFromFiber(c)
	ev.Type = audit.EventOTPSent
	ev.Success = true
	ev.Email = req.Email
	audit.Record(ev)

	// Статик key (i18n-д орчуулагдана) + email-ийг data-д өгнө — динамик
	// мессеж catalogMN-тай таарахгүй байсныг зассан.
	return v1.NewSuccessResponse(c, http.StatusOK, "otp code has been sent", fiber.Map{"email": req.Email})
}
