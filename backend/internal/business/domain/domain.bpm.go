// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package domain

import "time"

// BPM (Business Process Management)-ийн domain entity-үүд. domain.ai.go-тэй
// ижил зарчмаар зөвхөн стандарт сангаас хамаарна — HTTP, GORM, гадаад
// клиент энд орохгүй.
//
// Процесс нь дэлхийн стандарт, зөөвөрлөх боломжтой нэг **BPMN 2.0 (.bpmn)
// файл** болж хадгалагдана. User task-ийн дэлгэцүүд (form-js / Camunda Forms
// схем) нь тус .bpmn-ийн ДОТОР Camunda-ийн стандарт extension element-ээр
// embed хийгдсэн байна:
//
//	<bpmn:process …>
//	  <bpmn:extensionElements>
//	    <zeebe:userTaskForm id="UserTaskForm_X">{ …form-js JSON… }</zeebe:userTaskForm>
//	  </bpmn:extensionElements>
//	  <bpmn:userTask id="X">
//	    <bpmn:extensionElements>
//	      <zeebe:formDefinition formKey="camunda-forms:bpmn:UserTaskForm_X"/>
//	    </bpmn:extensionElements>
//	  </bpmn:userTask>
//	</bpmn:process>
//
// Ингэснээр хадгалсан зүйл нь Camunda Modeler-ээр нээгдэх жинхэнэ .bpmn файл.
// Entity-үүд XML-ийг түүхий мөр (string)-ээр хадгалж, задлах/шалгахыг usecase
// давхарга (encoding/xml) хийнэ.

// Процессын тодорхойлолтын төлөв.
const (
	BPMStatusDraft     = "draft"
	BPMStatusPublished = "published"
)

// Instance-ийн төлөв.
const (
	BPMInstanceRunning   = "running"
	BPMInstanceWaiting   = "waiting" // delegatedTask дээр peer-ийн callback хүлээж буй
	BPMInstanceCompleted = "completed"
	BPMInstanceCancelled = "cancelled"
	BPMInstanceFailed    = "failed"
)

// Task-ийн төлөв.
const (
	BPMTaskOpen      = "open"
	BPMTaskCompleted = "completed"
)

// Audit event-ийн төрлүүд (bpm_events).
const (
	BPMEventInstanceStarted   = "instance_started"
	BPMEventTaskOpened        = "task_opened"
	BPMEventTaskCompleted     = "task_completed"
	BPMEventServiceCalled     = "service_called"
	BPMEventServiceFailed     = "service_failed"
	BPMEventGatewayRouted     = "gateway_routed"
	BPMEventInstanceCompleted = "instance_completed"
	BPMEventInstanceFailed    = "instance_failed"
)

// BPMEvent нь нэг гүйлтийн төлөв өөрчлөлтийн бүртгэл (audit log). Дараалал нь
// created_at-аар тодорхойлогдоно; append-only.
type BPMEvent struct {
	ID         string
	InstanceID string
	UserID     string
	Type       string
	NodeID     string // холбогдох BPMN элементийн id (хэрэв байгаа бол)
	Detail     string // нэмэлт мэдээлэл (алдааны текст, сонгосон салаа г.м.)
	CreatedAt  time.Time
}

// BPMProcessDefinition нь хадгалагдсан процессын тодорхойлолт. Definition нь
// дээр тайлбарласан { bpmn, forms } JSON баримт.
type BPMProcessDefinition struct {
	ID          string
	UserID      string
	OrgID       string // харьяалагдах байгууллага (org-scoped хандалт)
	Name        string
	Description string
	Definition  string // { "bpmn": "...", "forms": { ... } } JSON
	Status      string
	Version     int
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}

// BPMForm нь олон процесс дунд хуваалцаж ашиглах форм (form library).
// Процессын userTask-ийн formKey `gerege-forms:<ID>`-ээр лавлана; engine нь
// openTask үед хамгийн сүүлийн Schema-г уншина (latest-wins).
type BPMForm struct {
	ID        string
	UserID    string
	Name      string
	Schema    string // form-js схем (JSON)
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// BPMProcessInstance нь процессын нэг гүйлт.
type BPMProcessInstance struct {
	ID            string
	DefinitionID  string
	UserID        string
	Status        string
	CurrentNodeID string // BPMN элементийн id (token-ий байрлал)
	// DefinitionSnapshot нь эхлэх агшны .bpmn — гүйлт энэ хувилбараар явна
	// (тодорхойлолтыг засахад хөндөгдөхгүй). Хоосон бол хуучин гүйлт.
	DefinitionSnapshot string
	// Федерацийн корреляци: энэ нь өөр node-оос (origin_peer) шилжүүлсэн дэд
	// гүйлт бол ParentInstanceID нь эх node дахь instance, OriginPeer нь
	// callback илгээх node. Хоосон бол энгийн (локал) гүйлт.
	ParentInstanceID string
	OriginPeer       string
	Variables        string // цуглуулсан хувьсагчид (JSON object)
	CreatedAt        time.Time
	UpdatedAt        *time.Time
	CompletedAt      *time.Time
}

// BPMTask нь хэрэглэгчийн бөглөх ёстой нэг даалгавар (BPMN user task-ийн дэлгэц).
type BPMTask struct {
	ID          string
	InstanceID  string
	UserID      string
	NodeID      string // BPMN userTask-ийн id
	Status      string
	Form        string  // тухайн userTask-ийн form-js схем (JSON snapshot)
	Submission  *string // бөглөсөн хариу (JSON); дуустал nil
	CreatedAt   time.Time
	CompletedAt *time.Time
}
