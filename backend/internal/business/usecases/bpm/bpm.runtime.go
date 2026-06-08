// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package bpm

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/domain"
	"geregetemplateai/pkg/logger"
)

// recordEvent нь audit log-д нэг бичлэг нэмнэ. Бүтэлгүйтэл нь процессыг
// унагах ёсгүй (audit нь туслах) тул алдааг log хийгээд залгина.
func (u *usecase) recordEvent(ctx context.Context, instanceID, userID, evType, nodeID, detail string) {
	if err := u.repo.CreateEvent(ctx, &domain.BPMEvent{
		InstanceID: instanceID,
		UserID:     userID,
		Type:       evType,
		NodeID:     nodeID,
		Detail:     detail,
	}); err != nil {
		logger.WarnWithContext(ctx, "bpm: failed to record audit event", logger.Fields{
			"usecase": "bpm", "instance_id": instanceID, "event": evType, "error": err.Error(),
		})
	}
}

// StartInstance нь процессын шинэ гүйлт эхлүүлж, эхний user task дээр идэвхтэй
// даалгавар нээнэ. start event-ээс эхний user task хүртэлх замд тааралдсан
// serviceTask-уудыг ГҮЙЦЭТГЭЖ (HTTP), exclusiveGateway-уудыг нөхцөлөөр салаалж
// дамжина.
func (u *usecase) StartInstance(ctx context.Context, req StartInstanceRequest) (RunResponse, error) {
	def, err := u.repo.GetDefinition(ctx, req.DefinitionID)
	if err != nil {
		return RunResponse{}, mapRepoError(err, "get process")
	}
	pp, err := parseDefinition(def.Definition)
	if err != nil {
		return RunResponse{}, err
	}

	instance, err := u.repo.CreateInstance(ctx, &domain.BPMProcessInstance{
		DefinitionID:  def.ID,
		UserID:        req.UserID,
		Status:        domain.BPMInstanceRunning,
		CurrentNodeID: "",
		// Эхлэх агшны .bpmn-ийг snapshot болгоно — гүйлт энэ хувилбараар явна,
		// тодорхойлолтыг засахад хөндөгдөхгүй.
		DefinitionSnapshot: def.Definition,
		Variables:          "{}",
	})
	if err != nil {
		return RunResponse{}, mapRepoError(err, "create instance")
	}
	u.recordEvent(ctx, instance.ID, instance.UserID, domain.BPMEventInstanceStarted, pp.startID, "")

	vars := map[string]interface{}{}
	stop, advErr := u.advance(ctx, pp, pp.startID, vars, instance.ID, instance.UserID)
	instance.Variables = marshalVars(vars)
	return u.settle(ctx, instance, pp, stop, advErr)
}

// GetActiveTask нь instance-ийн идэвхтэй (open) даалгаврыг буцаана.
func (u *usecase) GetActiveTask(ctx context.Context, req GetActiveTaskRequest) (RunResponse, error) {
	instance, err := u.repo.GetInstance(ctx, req.InstanceID)
	if err != nil {
		return RunResponse{}, mapRepoError(err, "get instance")
	}
	task, err := u.repo.GetOpenTaskByInstance(ctx, req.InstanceID)
	if err != nil {
		var domErr *apperror.DomainError
		if errors.As(err, &domErr) && domErr.Type == apperror.ErrTypeNotFound {
			return RunResponse{Instance: instance, Task: nil}, nil
		}
		return RunResponse{}, mapRepoError(err, "get task")
	}
	return RunResponse{Instance: instance, Task: &task}, nil
}

// ListInstances нь нэг процессын гүйлтүүдийг буцаана (мониторинг). Процессын
// эзэмшлийг GetDefinition (RLS)-ээр шалгана — өөр хэрэглэгчийнх бол NotFound.
func (u *usecase) ListInstances(ctx context.Context, req ListInstancesRequest) (ListInstancesResponse, error) {
	if _, err := u.repo.GetDefinition(ctx, req.DefinitionID); err != nil {
		return ListInstancesResponse{}, mapRepoError(err, "get process")
	}
	instances, err := u.repo.ListInstances(ctx, req.DefinitionID, req.Offset, req.Limit)
	if err != nil {
		return ListInstancesResponse{}, mapRepoError(err, "list instances")
	}
	return ListInstancesResponse{Instances: instances}, nil
}

// ListEvents нь нэг гүйлтийн audit timeline-г буцаана. Эзэмшлийг GetInstance
// (RLS)-ээр шалгана.
func (u *usecase) ListEvents(ctx context.Context, req ListEventsRequest) (ListEventsResponse, error) {
	if _, err := u.repo.GetInstance(ctx, req.InstanceID); err != nil {
		return ListEventsResponse{}, mapRepoError(err, "get instance")
	}
	events, err := u.repo.ListEvents(ctx, req.InstanceID, 0)
	if err != nil {
		return ListEventsResponse{}, mapRepoError(err, "list events")
	}
	return ListEventsResponse{Events: events}, nil
}

// SubmitTask нь даалгаврын хариуг хадгалж, дараагийн user task руу шилжинэ —
// замдаа serviceTask гүйцэтгэж, gateway-аар салаална.
func (u *usecase) SubmitTask(ctx context.Context, req SubmitTaskRequest) (RunResponse, error) {
	task, err := u.repo.GetTask(ctx, req.TaskID)
	if err != nil {
		return RunResponse{}, mapRepoError(err, "get task")
	}
	if task.Status != domain.BPMTaskOpen {
		return RunResponse{}, apperror.Conflict("task already completed")
	}

	data, err := parseObject(req.Data)
	if err != nil {
		return RunResponse{}, err
	}

	instance, err := u.repo.GetInstance(ctx, task.InstanceID)
	if err != nil {
		return RunResponse{}, mapRepoError(err, "get instance")
	}
	// Хувилбаржуулалт: гүйлт нь эхэлсэн агшны snapshot-оор явна (тодорхойлолтыг
	// дунд нь засахад хөндөгдөхгүй). Хоосон бол (хуучин гүйлт) одоогийн
	// тодорхойлолтыг fallback болгоно.
	bpmnSrc := instance.DefinitionSnapshot
	if strings.TrimSpace(bpmnSrc) == "" {
		def, defErr := u.repo.GetDefinition(ctx, instance.DefinitionID)
		if defErr != nil {
			return RunResponse{}, mapRepoError(defErr, "get process")
		}
		bpmnSrc = def.Definition
	}
	pp, err := parseDefinition(bpmnSrc)
	if err != nil {
		return RunResponse{}, err
	}

	vars, _ := parseObject(instance.Variables)
	for k, v := range data {
		vars[k] = v
	}

	if _, err := u.repo.CompleteTask(ctx, task.ID, req.Data); err != nil {
		return RunResponse{}, mapRepoError(err, "complete task")
	}
	u.recordEvent(ctx, instance.ID, instance.UserID, domain.BPMEventTaskCompleted, task.NodeID, "")

	stop, advErr := u.advance(ctx, pp, task.NodeID, vars, instance.ID, instance.UserID)
	instance.Variables = marshalVars(vars)
	return u.settle(ctx, instance, pp, stop, advErr)
}

// settle нь advance-ийн үр дүнгээр instance-ийг эцэслэнэ: алдаа → failed,
// зогсолтгүй → completed, user task → шинэ task нээх.
func (u *usecase) settle(ctx context.Context, instance domain.BPMProcessInstance, pp parsedProcess, stop string, advErr error) (RunResponse, error) {
	if advErr != nil {
		instance.Status = domain.BPMInstanceFailed
		instance.CurrentNodeID = ""
		updated, err := u.repo.UpdateInstance(ctx, &instance)
		if err != nil {
			return RunResponse{}, mapRepoError(err, "update instance")
		}
		u.recordEvent(ctx, instance.ID, instance.UserID, domain.BPMEventInstanceFailed, "", advErr.Error())
		return RunResponse{Instance: updated, Task: nil}, advErr
	}
	if stop == "" {
		return u.finishInstance(ctx, instance)
	}
	// delegatedTask — peer руу шилжүүлж 'waiting' болгоно (callback хүлээнэ).
	if pp.nodeType[stop] == kindDelegated {
		return u.delegateStep(ctx, instance, stop, pp)
	}
	instance.CurrentNodeID = stop
	updated, err := u.repo.UpdateInstance(ctx, &instance)
	if err != nil {
		return RunResponse{}, mapRepoError(err, "update instance")
	}
	task, err := u.openTask(ctx, updated, stop, pp)
	if err != nil {
		return RunResponse{}, err
	}
	return RunResponse{Instance: updated, Task: &task}, nil
}

// advance нь fromID-аас гарч flow-уудыг дагаж дараагийн user task хүртэл шилжинэ.
// serviceTask бүрийг гүйцэтгэж (vars-ыг шинэчилнэ), exclusiveGateway дээр нөхцөл
// тооцож салаална. user task тааралдвал түүний id, end / гарах flow байхгүй бол
// "" буцаана. vars-ыг газар дээр нь (in place) шинэчилнэ.
func (u *usecase) advance(ctx context.Context, pp parsedProcess, fromID string, vars map[string]interface{}, instanceID, userID string) (string, error) {
	visited := map[string]bool{}
	current := fromID
	for {
		flows := pp.outgoing[current]
		if len(flows) == 0 {
			return "", nil // гарах flow алга → дуусав
		}

		var next string
		if pp.nodeType[current] == kindGateway {
			next = chooseBranch(pp, current, flows, vars)
			if next == "" {
				return "", apperror.BadRequest("no gateway branch matched")
			}
			u.recordEvent(ctx, instanceID, userID, domain.BPMEventGatewayRouted, current, next)
		} else {
			next = flows[0].target
		}

		if visited[next] {
			return "", nil // мөчлөг илэрлээ → зогсооно
		}
		visited[next] = true

		switch pp.nodeType[next] {
		case kindUserTask:
			return next, nil
		case kindDelegated:
			// peer гүйцэтгэх алхам — энд зогсож, settle нь delegation.request
			// илгээж instance-ийг 'waiting' болгоно.
			return next, nil
		case kindEnd:
			return "", nil
		case kindService:
			if err := u.runService(ctx, pp.services[next], vars); err != nil {
				u.recordEvent(ctx, instanceID, userID, domain.BPMEventServiceFailed, next, err.Error())
				return "", err
			}
			u.recordEvent(ctx, instanceID, userID, domain.BPMEventServiceCalled, next, "")
			current = next
		default:
			// start (дахин таарвал) / бусад → үргэлжилнэ.
			current = next
		}
	}
}

// chooseBranch нь exclusive gateway-ийн салаануудаас нөхцөлд тохирохыг сонгоно.
// Эхлээд default бус салаануудыг (дараалсан) шалгаж, эхний тохирохыг авна; алга
// бол default салаа; тэр ч байхгүй бол "" (тохирох салаа алга).
func chooseBranch(pp parsedProcess, gatewayID string, flows []flow, vars map[string]interface{}) string {
	def := pp.gatewayDefault[gatewayID]
	for _, f := range flows {
		if def != "" && f.id == def {
			continue
		}
		if f.condition == "" || evalCondition(f.condition, vars) {
			return f.target
		}
	}
	if def != "" {
		for _, f := range flows {
			if f.id == def {
				return f.target
			}
		}
	}
	return ""
}

// runService нь serviceTask-ийн HTTP дуудлагыг гүйцэтгэнэ: url/body дотор
// `${var}` орлуулж, connector-оор дуудаж, хариуг resultVariable-д хадгална.
func (u *usecase) runService(ctx context.Context, svc serviceCall, vars map[string]interface{}) error {
	if u.connector == nil || svc.url == "" {
		return apperror.BadRequest("service task is not configured")
	}
	url := substitute(svc.url, vars)
	body := substitute(svc.body, vars)

	status, resp, err := u.connector.Do(ctx, svc.method, url, svc.headers, body)
	if err != nil {
		return apperror.Wrap(apperror.BadRequest("service task failed"), err)
	}
	if status >= 400 {
		return apperror.BadRequest("service task failed")
	}
	if svc.resultVar != "" {
		var parsed interface{}
		if json.Unmarshal(resp, &parsed) == nil {
			vars[svc.resultVar] = parsed
		} else {
			vars[svc.resultVar] = string(resp)
		}
	}
	return nil
}

// --- helpers ----------------------------------------------------------------

func (u *usecase) finishInstance(ctx context.Context, instance domain.BPMProcessInstance) (RunResponse, error) {
	instance.Status = domain.BPMInstanceCompleted
	instance.CurrentNodeID = ""
	updated, err := u.repo.UpdateInstance(ctx, &instance)
	if err != nil {
		return RunResponse{}, mapRepoError(err, "update instance")
	}
	u.recordEvent(ctx, instance.ID, instance.UserID, domain.BPMEventInstanceCompleted, "", "")
	// Энэ нь peer-ээс шилжсэн дэд гүйлт бол эх node руу delegation.callback
	// илгээж хувьсагчдыг буцаана (settle-ийн дараах гэхэд instance бүрэн).
	u.sendDelegationCallback(ctx, updated, domain.BPMInstanceCompleted)
	return RunResponse{Instance: updated, Task: nil}, nil
}

// openTask нь user task-д идэвхтэй даалгавар нээж, embed хийсэн form-js схемийн
// snapshot-ийг хадгална.
func (u *usecase) openTask(ctx context.Context, instance domain.BPMProcessInstance, nodeID string, pp parsedProcess) (domain.BPMTask, error) {
	formJSON := "{}"
	if raw, ok := pp.forms[nodeID]; ok && len(raw) > 0 {
		formJSON = string(raw)
	} else if fid, ok := pp.sharedForms[nodeID]; ok {
		// Хуваалцсан форм — сангаас хамгийн сүүлийн schema-г уншина (latest-wins).
		// Олдохгүй бол хоосон формоор үргэлжилнэ (процесс зогсохгүй).
		if f, gErr := u.repo.GetForm(ctx, fid); gErr == nil && strings.TrimSpace(f.Schema) != "" {
			formJSON = f.Schema
		} else if gErr != nil {
			logger.WarnWithContext(ctx, "bpm: shared form not resolved (using empty form)", logger.Fields{
				"usecase": "bpm", "method": "openTask", "node_id": nodeID, "form_id": fid, "error": gErr.Error(),
			})
		}
	}
	task, err := u.repo.CreateTask(ctx, &domain.BPMTask{
		InstanceID: instance.ID,
		UserID:     instance.UserID,
		NodeID:     nodeID,
		Form:       formJSON,
	})
	if err != nil {
		return domain.BPMTask{}, mapRepoError(err, "create task")
	}
	u.recordEvent(ctx, instance.ID, instance.UserID, domain.BPMEventTaskOpened, nodeID, "")
	return task, nil
}

// parseObject нь JSON object-ийг map болгож задална.
func parseObject(s string) (map[string]interface{}, error) {
	out := map[string]interface{}{}
	if s == "" {
		return out, nil
	}
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, apperror.BadRequest("invalid submission data")
	}
	return out, nil
}

// marshalVars нь хувьсагчдын map-ийг JSON object мөр болгоно.
func marshalVars(vars map[string]interface{}) string {
	b, err := json.Marshal(vars)
	if err != nil {
		return "{}"
	}
	return string(b)
}
