// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package federation нь федерацийн HTTP handler-ууд: peer registry (admin),
// inbound (peer-ийн гарын үсгээр баталгаажих, нэвтрэлтгүй), ping (туршилт).
package federation

import (
	"net/http"

	"geregetemplateai/internal/business/domain"
	feduc "geregetemplateai/internal/business/usecases/federation"
	"geregetemplateai/internal/http/datatransfers/requests"
	"geregetemplateai/internal/http/datatransfers/responses"
	v1 "geregetemplateai/internal/http/handlers/v1"
	"geregetemplateai/pkg/validators"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type Handler struct {
	usecase feduc.Usecase
}

func NewHandler(usecase feduc.Usecase) Handler {
	return Handler{usecase: usecase}
}

// Status godoc
// @Summary  Федерацийн төлөв (энэ node)
// @Tags     federation
// @Produce  json
// @Security BearerAuth
// @Success  200 {object} v1.BaseResponse{data=responses.FedStatusResponse}
// @Router   /fed/status [get]
func (h Handler) Status(c fiber.Ctx) error {
	return v1.NewSuccessResponse(c, http.StatusOK, "ok", responses.FedStatusResponse{
		Configured: h.usecase.Configured(),
		NodeId:     h.usecase.NodeID(),
	})
}

// ListPeers godoc
// @Summary  Гишүүн node-уудыг жагсаах
// @Tags     federation
// @Produce  json
// @Security BearerAuth
// @Success  200 {object} v1.BaseResponse{data=[]responses.FedPeerResponse}
// @Router   /fed/peers [get]
func (h Handler) ListPeers(c fiber.Ctx) error {
	peers, err := h.usecase.ListPeers(c.Context())
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "peers fetched successfully", responses.ToFedPeerList(peers))
}

// CreatePeer godoc
// @Summary  Гишүүн node бүртгэх
// @Tags     federation
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    request body requests.FedPeerCreateRequest true "Peer"
// @Success  201 {object} v1.BaseResponse{data=responses.FedPeerResponse}
// @Router   /fed/peers [post]
func (h Handler) CreatePeer(c fiber.Ctx) error {
	var req requests.FedPeerCreateRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	peer, err := h.usecase.RegisterPeer(c.Context(), domain.FedPeer{
		Key: req.Key, Name: req.Name, OrgID: req.OrgID,
		BaseURL: req.BaseURL, JWKSURL: req.JWKSURL, Status: req.Status,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusCreated, "peer registered successfully", responses.FromFedPeer(peer))
}

// UpdatePeer godoc
// @Summary  Гишүүн node засах
// @Tags     federation
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id      path string                       true "Peer ID"
// @Param    request body requests.FedPeerUpdateRequest true "Peer"
// @Success  200 {object} v1.BaseResponse{data=responses.FedPeerResponse}
// @Router   /fed/peers/{id} [put]
func (h Handler) UpdatePeer(c fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid peer id")
	}
	var req requests.FedPeerUpdateRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	peer, err := h.usecase.UpdatePeer(c.Context(), domain.FedPeer{
		ID: id, Name: req.Name, OrgID: req.OrgID,
		BaseURL: req.BaseURL, JWKSURL: req.JWKSURL, Status: req.Status,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "peer updated successfully", responses.FromFedPeer(peer))
}

// DeletePeer godoc
// @Summary  Гишүүн node устгах
// @Tags     federation
// @Produce  json
// @Security BearerAuth
// @Param    id path string true "Peer ID"
// @Success  200 {object} v1.BaseResponse
// @Router   /fed/peers/{id} [delete]
func (h Handler) DeletePeer(c fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid peer id")
	}
	if err := h.usecase.DeletePeer(c.Context(), id); err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "peer deleted successfully", nil)
}

// Ping godoc
// @Summary  Peer руу гарын үсэгтэй ping илгээж round-trip шалгах
// @Tags     federation
// @Produce  json
// @Security BearerAuth
// @Param    id path string true "Peer ID"
// @Success  200 {object} v1.BaseResponse
// @Router   /fed/peers/{id}/ping [post]
func (h Handler) Ping(c fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid peer id")
	}
	if err := h.usecase.Send(c.Context(), id, "ping", nil); err != nil {
		return v1.RespondWithError(c, err)
	}
	// Туршилтад нэн даруй хүргэхийн тулд outbox-ийг нэг удаа боловсруулна.
	sent := h.usecase.ProcessOutbox(c.Context())
	return v1.NewSuccessResponse(c, http.StatusOK, "ping enqueued", fiber.Map{"delivered": sent})
}

// Flush godoc
// @Summary  Гарах дарааллыг (outbox) нэн даруй боловсруулах
// @Tags     federation
// @Produce  json
// @Security BearerAuth
// @Success  200 {object} v1.BaseResponse
// @Router   /fed/flush [post]
//
// Flush нь delegation request→callback каскадыг нэг дуудлагаар (хэд хэдэн алхам)
// шавхана — мониторинг/туршилтад тус болно.
func (h Handler) Flush(c fiber.Ctx) error {
	total := 0
	for i := 0; i < 5; i++ {
		n := h.usecase.ProcessOutbox(c.Context())
		total += n
		if n == 0 {
			break
		}
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "outbox flushed", fiber.Map{"delivered": total})
}

// Inbound godoc
// @Summary  Орох федерацийн мессеж (peer гарын үсгээр баталгаажна)
// @Tags     federation
// @Accept   plain
// @Produce  json
// @Success  200 {object} v1.BaseResponse
// @Router   /fed/inbound [post]
//
// Inbound нь нэвтрэлтгүй — итгэлийг peer-ийн ES256 гарын үсгээр (registry-ийн
// jwks_url) тогтооно.
func (h Handler) Inbound(c fiber.Ctx) error {
	body := c.Body() // Fiber буфферлэсэн биеийг буцаана (JWS текст)
	res, herr := h.usecase.HandleInbound(c.Context(), body)
	if herr != nil {
		return v1.RespondWithError(c, herr)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "ok", res)
}
