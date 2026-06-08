// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// bpm.federation.go нь delegatedTask-ийн engine-федераци гүүр: алхмыг peer руу
// шилжүүлэх (delegateStep), дэд гүйлт эхлүүлэх (StartDelegated), callback-аар эх
// гүйлтийг үргэлжлүүлэх (ResumeDelegated). Гарын үсэг/тээвэр нь federation
// usecase (FedSender)-д; энд зөвхөн engine-ийн логик.
package bpm

import (
	"context"
	"encoding/json"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/domain"
	"geregetemplateai/pkg/logger"
)

// delegationRequestBody нь delegation.request мессежийн агуулга.
type delegationRequestBody struct {
	ProcessKey     string                 `json:"process_key"`
	Variables      map[string]interface{} `json:"variables"`
	ParentInstance string                 `json:"parent_instance"`
}

// delegationCallbackBody нь delegation.callback мессежийн агуулга.
type delegationCallbackBody struct {
	ParentInstance string                 `json:"parent_instance"`
	Variables      map[string]interface{} `json:"variables"`
	Status         string                 `json:"status"`
}

// delegateStep нь delegatedTask дээр peer руу delegation.request илгээж
// instance-ийг 'waiting' болгоно (callback хүлээнэ).
func (u *usecase) delegateStep(ctx context.Context, instance domain.BPMProcessInstance, nodeID string, pp parsedProcess) (RunResponse, error) {
	dc := pp.delegates[nodeID]
	if u.fedSender == nil {
		return u.failInstance(ctx, instance, "federation is not configured for delegatedTask")
	}
	vars, _ := parseObject(instance.Variables)
	body, _ := json.Marshal(delegationRequestBody{
		ProcessKey:     dc.processKey,
		Variables:      vars,
		ParentInstance: instance.ID,
	})
	if err := u.fedSender.SendByKey(ctx, dc.peer, "delegation.request", body); err != nil {
		return u.failInstance(ctx, instance, "delegate send: "+err.Error())
	}
	instance.Status = domain.BPMInstanceWaiting
	instance.CurrentNodeID = nodeID
	updated, err := u.repo.UpdateInstance(ctx, &instance)
	if err != nil {
		return RunResponse{}, mapRepoError(err, "update instance")
	}
	u.recordEvent(ctx, instance.ID, instance.UserID, domain.BPMEventServiceCalled, nodeID, "delegated to "+dc.peer)
	return RunResponse{Instance: updated, Task: nil}, nil
}

// sendDelegationCallback нь дэд гүйлт (origin_peer-тэй) дуусахад эх node руу
// үр дүнг буцаана. Best-effort (outbox найдвартай хүргэнэ).
func (u *usecase) sendDelegationCallback(ctx context.Context, instance domain.BPMProcessInstance, status string) {
	if instance.OriginPeer == "" || u.fedSender == nil {
		return
	}
	vars, _ := parseObject(instance.Variables)
	body, _ := json.Marshal(delegationCallbackBody{
		ParentInstance: instance.ParentInstanceID,
		Variables:      vars,
		Status:         status,
	})
	if err := u.fedSender.SendByKey(ctx, instance.OriginPeer, "delegation.callback", body); err != nil {
		logger.ErrorWithContext(ctx, "delegation callback send failed", logger.Fields{
			"instance": instance.ID, "origin_peer": instance.OriginPeer, "error": err.Error(),
		})
	}
}

// StartDelegated нь peer-ээс ирсэн delegation.request-ийн дагуу локал процессын
// (нэрээр тааруулсан) дэд гүйлт эхлүүлж явуулна (federation.DelegationHandler).
func (u *usecase) StartDelegated(ctx context.Context, processKey string, vars json.RawMessage, originPeer, parentInstance string) error {
	def, err := u.repo.GetDefinitionByName(ctx, processKey)
	if err != nil {
		return err
	}
	pp, err := parseDefinition(def.Definition)
	if err != nil {
		return err
	}
	startVars := "{}"
	if len(vars) > 0 {
		startVars = string(vars)
	}
	instance, err := u.repo.CreateInstance(ctx, &domain.BPMProcessInstance{
		DefinitionID:       def.ID,
		UserID:             def.UserID,
		Status:             domain.BPMInstanceRunning,
		CurrentNodeID:      "",
		DefinitionSnapshot: def.Definition,
		ParentInstanceID:   parentInstance,
		OriginPeer:         originPeer,
		Variables:          startVars,
	})
	if err != nil {
		return mapRepoError(err, "create delegated instance")
	}
	u.recordEvent(ctx, instance.ID, instance.UserID, domain.BPMEventInstanceStarted, pp.startID, "delegated from "+originPeer)
	v, _ := parseObject(instance.Variables)
	stop, advErr := u.advance(ctx, pp, pp.startID, v, instance.ID, instance.UserID)
	instance.Variables = marshalVars(v)
	// settle нь дуусгах/waiting/user-task-г шийднэ; дуусвал finishInstance
	// нь callback-ийг автоматаар илгээнэ.
	_, serr := u.settle(ctx, instance, pp, stop, advErr)
	return serr
}

// ResumeDelegated нь peer-ийн callback-аар хүлээж буй эх гүйлтийг хувьсагч
// нэгтгэн delegatedTask-аас цааш үргэлжлүүлнэ.
func (u *usecase) ResumeDelegated(ctx context.Context, parentInstance string, vars json.RawMessage, status string) error {
	instance, err := u.repo.GetInstance(ctx, parentInstance)
	if err != nil {
		return err
	}
	if instance.Status != domain.BPMInstanceWaiting {
		// Аль хэдийн үргэлжилсэн/дууссан — давхар callback-ийг алгасна (idempotency).
		return nil
	}
	if status == domain.BPMInstanceFailed {
		_, ferr := u.failInstance(ctx, instance, "delegated step failed on peer")
		return ferr
	}
	merged, _ := parseObject(instance.Variables)
	if incoming, perr := parseObject(string(vars)); perr == nil {
		for k, val := range incoming {
			merged[k] = val
		}
	}
	bpmnSrc := instance.DefinitionSnapshot
	if bpmnSrc == "" {
		def, derr := u.repo.GetDefinition(ctx, instance.DefinitionID)
		if derr != nil {
			return mapRepoError(derr, "get process")
		}
		bpmnSrc = def.Definition
	}
	pp, err := parseDefinition(bpmnSrc)
	if err != nil {
		return err
	}
	instance.Status = domain.BPMInstanceRunning
	stop, advErr := u.advance(ctx, pp, instance.CurrentNodeID, merged, instance.ID, instance.UserID)
	instance.Variables = marshalVars(merged)
	_, serr := u.settle(ctx, instance, pp, stop, advErr)
	return serr
}

// failInstance нь instance-ийг failed болгож алдааг бүртгэнэ.
func (u *usecase) failInstance(ctx context.Context, instance domain.BPMProcessInstance, reason string) (RunResponse, error) {
	instance.Status = domain.BPMInstanceFailed
	instance.CurrentNodeID = ""
	updated, err := u.repo.UpdateInstance(ctx, &instance)
	if err != nil {
		return RunResponse{}, mapRepoError(err, "update instance")
	}
	u.recordEvent(ctx, instance.ID, instance.UserID, domain.BPMEventInstanceFailed, "", reason)
	return RunResponse{Instance: updated, Task: nil}, apperror.BadRequest(reason)
}
