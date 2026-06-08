// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package records

import (
	"time"

	"geregetemplateai/internal/business/domain"
)

// JSONB баганууд record давхаргад түүхий JSON мөр (string) болж хадгалагдана.
// Уншихад Postgres jsonb-г текст болгон буцаадаг; бичихдээ raw SQL дотор
// `?::jsonb` cast хийнэ (bpm.*.go-г үз). Энэ нь datatypes.JSON шиг нэмэлт
// хамаарал авахгүйгээр энгийн record↔domain буулгалт өгнө.

// BPMProcessDefinitions нь bpm_process_definitions хүснэгтийн GORM model.
type BPMProcessDefinitions struct {
	Id          string     `gorm:"column:id;primaryKey"`
	UserId      string     `gorm:"column:user_id"`
	OrgId       string     `gorm:"column:org_id;type:uuid"`
	Name        string     `gorm:"column:name"`
	Description string     `gorm:"column:description"`
	Definition  string     `gorm:"column:definition;type:text"` // цэвэр BPMN 2.0 XML (.bpmn)
	Status      string     `gorm:"column:status"`
	Version     int        `gorm:"column:version"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at"`
}

func (BPMProcessDefinitions) TableName() string { return "bpm_process_definitions" }

func (r BPMProcessDefinitions) ToV1Domain() domain.BPMProcessDefinition {
	return domain.BPMProcessDefinition{
		ID:          r.Id,
		UserID:      r.UserId,
		OrgID:       r.OrgId,
		Name:        r.Name,
		Description: r.Description,
		Definition:  r.Definition,
		Status:      r.Status,
		Version:     r.Version,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// BPMForms нь bpm_forms (хуваалцсан форм сан) хүснэгтийн GORM model.
type BPMForms struct {
	Id        string     `gorm:"column:id;primaryKey"`
	UserId    string     `gorm:"column:user_id"`
	Name      string     `gorm:"column:name"`
	Schema    string     `gorm:"column:schema;type:text"` // form-js схем (JSON)
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
}

func (BPMForms) TableName() string { return "bpm_forms" }

func (r BPMForms) ToV1Domain() domain.BPMForm {
	return domain.BPMForm{
		ID:        r.Id,
		UserID:    r.UserId,
		Name:      r.Name,
		Schema:    r.Schema,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// BPMProcessInstances нь bpm_process_instances хүснэгтийн GORM model.
type BPMProcessInstances struct {
	Id                 string     `gorm:"column:id;primaryKey"`
	DefinitionId       string     `gorm:"column:definition_id"`
	UserId             string     `gorm:"column:user_id"`
	Status             string     `gorm:"column:status"`
	CurrentNodeId      string     `gorm:"column:current_node_id"`
	DefinitionSnapshot string     `gorm:"column:definition_snapshot;type:text"`
	ParentInstanceId   *string    `gorm:"column:parent_instance_id"`
	OriginPeer         string     `gorm:"column:origin_peer"`
	Variables          string     `gorm:"column:variables;type:jsonb"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedAt          *time.Time `gorm:"column:updated_at"`
	CompletedAt        *time.Time `gorm:"column:completed_at"`
}

func (BPMProcessInstances) TableName() string { return "bpm_process_instances" }

func (r BPMProcessInstances) ToV1Domain() domain.BPMProcessInstance {
	return domain.BPMProcessInstance{
		ID:                 r.Id,
		DefinitionID:       r.DefinitionId,
		UserID:             r.UserId,
		Status:             r.Status,
		CurrentNodeID:      r.CurrentNodeId,
		DefinitionSnapshot: r.DefinitionSnapshot,
		ParentInstanceID:   deref(r.ParentInstanceId),
		OriginPeer:         r.OriginPeer,
		Variables:          r.Variables,
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
		CompletedAt:        r.CompletedAt,
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// BPMTasks нь bpm_tasks хүснэгтийн GORM model.
type BPMTasks struct {
	Id          string     `gorm:"column:id;primaryKey"`
	InstanceId  string     `gorm:"column:instance_id"`
	UserId      string     `gorm:"column:user_id"`
	NodeId      string     `gorm:"column:node_id"`
	Status      string     `gorm:"column:status"`
	Form        string     `gorm:"column:form;type:jsonb"`
	Submission  *string    `gorm:"column:submission;type:jsonb"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	CompletedAt *time.Time `gorm:"column:completed_at"`
}

func (BPMTasks) TableName() string { return "bpm_tasks" }

// BPMEvents нь bpm_events хүснэгтийн GORM model (audit log).
type BPMEvents struct {
	Id         string    `gorm:"column:id;primaryKey"`
	InstanceId string    `gorm:"column:instance_id"`
	UserId     string    `gorm:"column:user_id"`
	Type       string    `gorm:"column:type"`
	NodeId     string    `gorm:"column:node_id"`
	Detail     string    `gorm:"column:detail"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (BPMEvents) TableName() string { return "bpm_events" }

func (r BPMEvents) ToV1Domain() domain.BPMEvent {
	return domain.BPMEvent{
		ID:         r.Id,
		InstanceID: r.InstanceId,
		UserID:     r.UserId,
		Type:       r.Type,
		NodeID:     r.NodeId,
		Detail:     r.Detail,
		CreatedAt:  r.CreatedAt,
	}
}

func (r BPMTasks) ToV1Domain() domain.BPMTask {
	return domain.BPMTask{
		ID:          r.Id,
		InstanceID:  r.InstanceId,
		UserID:      r.UserId,
		NodeID:      r.NodeId,
		Status:      r.Status,
		Form:        r.Form,
		Submission:  r.Submission,
		CreatedAt:   r.CreatedAt,
		CompletedAt: r.CompletedAt,
	}
}
