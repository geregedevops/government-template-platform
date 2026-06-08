// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package responses

import (
	"encoding/json"
	"time"

	"geregetemplateai/internal/business/domain"
)

// BPMProcessResponse нь процессын тодорхойлолтын DTO. BPMN нь цэвэр BPMN 2.0
// XML (.bpmn) файл (маягтууд дотроо embed хийгдсэн).
type BPMProcessResponse struct {
	Id          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	BPMN        string     `json:"bpmn"`
	Status      string     `json:"status"`
	Version     int        `json:"version"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

func FromBPMProcess(p domain.BPMProcessDefinition) BPMProcessResponse {
	return BPMProcessResponse{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		BPMN:        p.Definition,
		Status:      p.Status,
		Version:     p.Version,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func ToBPMProcessList(items []domain.BPMProcessDefinition) []BPMProcessResponse {
	out := make([]BPMProcessResponse, 0, len(items))
	for _, p := range items {
		out = append(out, FromBPMProcess(p))
	}
	return out
}

// BPMFormResponse нь хуваалцсан формын DTO. Schema нь form-js схем (JSON object).
type BPMFormResponse struct {
	Id        string          `json:"id"`
	Name      string          `json:"name"`
	Schema    json.RawMessage `json:"schema"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt *time.Time      `json:"updated_at"`
}

func FromBPMForm(f domain.BPMForm) BPMFormResponse {
	return BPMFormResponse{
		Id:        f.ID,
		Name:      f.Name,
		Schema:    rawOrEmpty(f.Schema),
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}
}

func ToBPMFormList(items []domain.BPMForm) []BPMFormResponse {
	out := make([]BPMFormResponse, 0, len(items))
	for _, f := range items {
		out = append(out, FromBPMForm(f))
	}
	return out
}

// BPMInstanceResponse нь нэг гүйлтийн төлөв.
type BPMInstanceResponse struct {
	Id            string          `json:"id"`
	DefinitionId  string          `json:"definition_id"`
	Status        string          `json:"status"`
	CurrentNodeId string          `json:"current_node_id"`
	Variables     json.RawMessage `json:"variables"`
	CreatedAt     time.Time       `json:"created_at"`
	CompletedAt   *time.Time      `json:"completed_at"`
}

func FromBPMInstance(i domain.BPMProcessInstance) BPMInstanceResponse {
	return BPMInstanceResponse{
		Id:            i.ID,
		DefinitionId:  i.DefinitionID,
		Status:        i.Status,
		CurrentNodeId: i.CurrentNodeID,
		Variables:     rawOrEmpty(i.Variables),
		CreatedAt:     i.CreatedAt,
		CompletedAt:   i.CompletedAt,
	}
}

func ToBPMInstanceList(items []domain.BPMProcessInstance) []BPMInstanceResponse {
	out := make([]BPMInstanceResponse, 0, len(items))
	for _, i := range items {
		out = append(out, FromBPMInstance(i))
	}
	return out
}

// BPMTaskResponse нь рендерлэх нэг даалгавар (form node-ийн дэлгэц).
type BPMTaskResponse struct {
	Id     string          `json:"id"`
	NodeId string          `json:"node_id"`
	Status string          `json:"status"`
	Form   json.RawMessage `json:"form"`
}

func FromBPMTask(t domain.BPMTask) BPMTaskResponse {
	return BPMTaskResponse{
		Id:     t.ID,
		NodeId: t.NodeID,
		Status: t.Status,
		Form:   rawOrEmpty(t.Form),
	}
}

// BPMEventResponse нь audit log-ийн нэг бичлэг.
type BPMEventResponse struct {
	Id        string    `json:"id"`
	Type      string    `json:"type"`
	NodeId    string    `json:"node_id"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

func ToBPMEventList(items []domain.BPMEvent) []BPMEventResponse {
	out := make([]BPMEventResponse, 0, len(items))
	for _, e := range items {
		out = append(out, BPMEventResponse{
			Id:        e.ID,
			Type:      e.Type,
			NodeId:    e.NodeID,
			Detail:    e.Detail,
			CreatedAt: e.CreatedAt,
		})
	}
	return out
}

// BPMRunResponse нь гүйлтийн нэг алхмын хариу — instance + (байвал) идэвхтэй
// даалгавар. Task nil бол гүйлт дууссан.
type BPMRunResponse struct {
	Instance BPMInstanceResponse `json:"instance"`
	Task     *BPMTaskResponse    `json:"task"`
	Done     bool                `json:"done"`
}

func FromBPMRun(instance domain.BPMProcessInstance, task *domain.BPMTask) BPMRunResponse {
	res := BPMRunResponse{
		Instance: FromBPMInstance(instance),
		Done:     task == nil,
	}
	if task != nil {
		t := FromBPMTask(*task)
		res.Task = &t
	}
	return res
}

// rawOrEmpty нь хоосон JSON мөрийг хүчинтэй "{}" болгож хамгаална.
func rawOrEmpty(s string) json.RawMessage {
	if s == "" {
		return json.RawMessage("{}")
	}
	return json.RawMessage(s)
}
