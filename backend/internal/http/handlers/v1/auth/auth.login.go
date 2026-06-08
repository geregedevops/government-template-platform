// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package auth

import (
	"net/http"

	authuc "geregetemplateai/internal/business/usecases/auth"
	"geregetemplateai/internal/http/datatransfers/requests"
	"geregetemplateai/internal/http/datatransfers/responses"
	v1 "geregetemplateai/internal/http/handlers/v1"
	"geregetemplateai/pkg/audit"
	"geregetemplateai/pkg/logger"
	"geregetemplateai/pkg/validators"
	"github.com/gofiber/fiber/v3"
)

// Login godoc
// @Summary      Баталгаажуулж токен хос олгох
// @Description  Access токен (богино TTL) болон refresh токен (урт TTL) буцаана. Хэрэглэгчийг тоолохоос сэргийлэхийн тулд буруу нууц үг болон тодорхойгүй email ижил хугацаа зарцуулна.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      requests.LoginRequest  true  "Login credentials"
// @Success      200      {object}  v1.BaseResponse{data=responses.UserResponse}  "Tokens issued"
// @Failure      400      {object}  v1.BaseResponse                                "Malformed JSON body"
// @Failure      401      {object}  v1.BaseResponse                                "Invalid email or password"
// @Failure      403      {object}  v1.BaseResponse                                "Account not yet activated"
// @Failure      422      {object}  v1.BaseResponse                                "Validation error"
// @Router       /auth/login [post]
func (h Handler) Login(c fiber.Ctx) error {
	const (
		controllerName = "auth"
		funcName       = "Login"
		fileName       = "auth.login.go"
	)
	ctx := c.Context()
	var req requests.LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		logger.WarnWithContext(ctx, "Login: invalid request body", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
		})
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		logger.WarnWithContext(ctx, "Login: validation error", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"request": logger.Fields{
				"email":        req.Email,
				"has_password": req.Password != "",
			},
		})
		return v1.RespondWithError(c, err)
	}

	// Клиентийн хаяг — BFF (Next.js) X-Client-IP-ээр дамжуулдаг (nginx → web →
	// api). Байхгүй бол шууд холболтын IP. Lockout-ийг (email, IP)-ээр түлхүүрлэнэ.
	clientIP := c.Get("X-Client-IP")
	if clientIP == "" {
		clientIP = c.IP()
	}
	result, err := h.usecase.Login(ctx, authuc.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
		IP:       clientIP,
	})
	if err != nil {
		ev := auditFromFiber(c)
		ev.Type = audit.EventLoginFailure
		ev.Email = req.Email
		ev.Reason = err.Error()
		audit.Record(ev)
		logger.ErrorWithContext(ctx, "Login failed in controller", logger.Fields{
			"controller": controllerName,
			"method":     funcName,
			"file":       fileName,
			"error":      err.Error(),
			"email":      req.Email,
		})
		return v1.RespondWithError(c, err)
	}

	ev := auditFromFiber(c)
	ev.Type = audit.EventLoginSuccess
	ev.Success = true
	ev.UserID = result.User.ID
	ev.Email = result.User.Email
	audit.Record(ev)

	return v1.NewSuccessResponse(c, http.StatusOK, "login success", responses.FromLoginResponse(result))
}
