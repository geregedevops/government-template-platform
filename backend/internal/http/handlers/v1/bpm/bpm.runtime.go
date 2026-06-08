// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package bpm

import (
	"net/http"

	bpmuc "geregetemplateai/internal/business/usecases/bpm"
	"geregetemplateai/internal/http/auth"
	"geregetemplateai/internal/http/datatransfers/requests"
	"geregetemplateai/internal/http/datatransfers/responses"
	v1 "geregetemplateai/internal/http/handlers/v1"
	"geregetemplateai/pkg/validators"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// StartInstance godoc
// @Summary      BPM процессын гүйлт эхлүүлэх
// @Tags         bpm
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Process ID (uuid)"
// @Success      201  {object}  v1.BaseResponse{data=responses.BPMRunResponse}
// @Failure      401  {object}  v1.BaseResponse
// @Failure      404  {object}  v1.BaseResponse
// @Router       /bpm/processes/{id}/start [post]
func (h Handler) StartInstance(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "process not found")
	}
	res, err := h.usecase.StartInstance(c.Context(), bpmuc.StartInstanceRequest{
		UserID:       user.ID,
		DefinitionID: id,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusCreated, "instance started successfully",
		responses.FromBPMRun(res.Instance, res.Task))
}

// GetActiveTask godoc
// @Summary      Гүйлтийн идэвхтэй даалгаврыг авах (рендерлэх дэлгэц)
// @Tags         bpm
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Instance ID (uuid)"
// @Success      200  {object}  v1.BaseResponse{data=responses.BPMRunResponse}
// @Failure      401  {object}  v1.BaseResponse
// @Failure      404  {object}  v1.BaseResponse
// @Router       /bpm/instances/{id}/task [get]
func (h Handler) GetActiveTask(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "instance not found")
	}
	res, err := h.usecase.GetActiveTask(c.Context(), bpmuc.GetActiveTaskRequest{
		UserID:     user.ID,
		InstanceID: id,
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "task fetched successfully",
		responses.FromBPMRun(res.Instance, res.Task))
}

// ListInstances godoc
// @Summary      Процессын гүйлтүүдийг (instances) жагсаах
// @Tags         bpm
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Process ID (uuid)"
// @Success      200  {object}  v1.BaseResponse{data=[]responses.BPMInstanceResponse}
// @Failure      401  {object}  v1.BaseResponse
// @Failure      404  {object}  v1.BaseResponse
// @Router       /bpm/processes/{id}/instances [get]
func (h Handler) ListInstances(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "process not found")
	}
	res, err := h.usecase.ListInstances(c.Context(), bpmuc.ListInstancesRequest{
		UserID:       user.ID,
		DefinitionID: id,
		Offset:       fiber.Query[int](c, "offset", 0),
		Limit:        fiber.Query[int](c, "limit", 50),
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "instances fetched successfully",
		responses.ToBPMInstanceList(res.Instances))
}

// ListEvents godoc
// @Summary      Гүйлтийн audit timeline (events)
// @Tags         bpm
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Instance ID (uuid)"
// @Success      200  {object}  v1.BaseResponse{data=[]responses.BPMEventResponse}
// @Failure      401  {object}  v1.BaseResponse
// @Failure      404  {object}  v1.BaseResponse
// @Router       /bpm/instances/{id}/events [get]
func (h Handler) ListEvents(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "instance not found")
	}
	res, err := h.usecase.ListEvents(c.Context(), bpmuc.ListEventsRequest{UserID: user.ID, InstanceID: id})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "events fetched successfully",
		responses.ToBPMEventList(res.Events))
}

// SubmitTask godoc
// @Summary      Form даалгаврыг бөглөж дараагийн алхам руу шилжих
// @Tags         bpm
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  string                        true  "Task ID (uuid)"
// @Param        request  body  requests.BPMSubmitTaskRequest  true  "Submission"
// @Success      200  {object}  v1.BaseResponse{data=responses.BPMRunResponse}
// @Failure      401  {object}  v1.BaseResponse
// @Failure      404  {object}  v1.BaseResponse
// @Failure      409  {object}  v1.BaseResponse
// @Failure      422  {object}  v1.BaseResponse
// @Router       /bpm/tasks/{id}/submit [post]
func (h Handler) SubmitTask(c fiber.Ctx) error {
	user, err := auth.CurrentUserFromContext(c)
	if err != nil {
		return v1.NewAbortResponse(c, "invalid token")
	}
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return v1.NewErrorResponse(c, http.StatusNotFound, "task not found")
	}
	var req requests.BPMSubmitTaskRequest
	if err := c.Bind().Body(&req); err != nil {
		return v1.NewErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := validators.ValidatePayloads(req); err != nil {
		return v1.RespondWithError(c, err)
	}
	res, err := h.usecase.SubmitTask(c.Context(), bpmuc.SubmitTaskRequest{
		UserID: user.ID,
		TaskID: id,
		Data:   string(req.Data),
	})
	if err != nil {
		return v1.RespondWithError(c, err)
	}
	return v1.NewSuccessResponse(c, http.StatusOK, "task submitted successfully",
		responses.FromBPMRun(res.Instance, res.Task))
}
