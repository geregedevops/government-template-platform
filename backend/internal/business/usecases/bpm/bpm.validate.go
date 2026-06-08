// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package bpm

import (
	"encoding/json"
	"encoding/xml"
	"strings"

	"geregetemplateai/internal/apperror"
)

// definition нь цэвэр BPMN 2.0 XML (.bpmn). Доорхи стандарт extension-уудыг
// уншина:
//   • user task маягт: process>extensionElements>zeebe:userTaskForm + userTask
//     >extensionElements>zeebe:formDefinition (formKey)
//   • service task HTTP: serviceTask>extensionElements>zeebe:taskHeaders>header
//     (method/url/body/resultVariable)
//   • exclusive gateway салаа: sequenceFlow>conditionExpression + gateway-ийн
//     default атрибут

// formKeyPrefix нь Camunda-ийн embed хийсэн маягтын formKey-ийн угтвар.
const formKeyPrefix = "camunda-forms:bpmn:"

// sharedFormKeyPrefix нь хуваалцсан (linked) формын formKey угтвар:
// `gerege-forms:<formId>`. Engine нь openTask үед энэ ID-аар bpm_forms сангаас
// хамгийн сүүлийн schema-г уншина (latest-wins).
const sharedFormKeyPrefix = "gerege-forms:"

// --- BPMN 2.0 XML-ийн хэрэгцээт дэд олонлог (encoding/xml) -------------------
// Go-ийн encoding/xml нь элементүүдийг local нэрээр (namespace prefix үл
// харгалзан) тааруулдаг.

type bpmnDefinitions struct {
	XMLName   xml.Name      `xml:"definitions"`
	Processes []bpmnProcess `xml:"process"`
}

type bpmnProcess struct {
	ID                string             `xml:"id,attr"`
	UserTaskForms     []bpmnUserTaskForm `xml:"extensionElements>userTaskForm"`
	StartEvents       []bpmnNode         `xml:"startEvent"`
	EndEvents         []bpmnNode         `xml:"endEvent"`
	UserTasks         []bpmnUserTask     `xml:"userTask"`
	ServiceTasks      []bpmnServiceTask  `xml:"serviceTask"`
	ExclusiveGateways []bpmnGateway      `xml:"exclusiveGateway"`
	SequenceFlows     []bpmnSequenceFlow `xml:"sequenceFlow"`
}

// bpmnUserTaskForm нь process-ийн extensionElements дотор embed хийсэн form-js
// схем.
type bpmnUserTaskForm struct {
	ID      string `xml:"id,attr"`
	Content string `xml:",chardata"`
}

type bpmnNode struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

type bpmnUserTask struct {
	ID             string `xml:"id,attr"`
	Name           string `xml:"name,attr"`
	FormDefinition struct {
		FormKey string `xml:"formKey,attr"`
	} `xml:"extensionElements>formDefinition"`
}

// bpmnServiceTask нь zeebe:taskHeaders-аар HTTP тохиргоог авна.
type bpmnServiceTask struct {
	ID      string       `xml:"id,attr"`
	Name    string       `xml:"name,attr"`
	Headers []bpmnHeader `xml:"extensionElements>taskHeaders>header"`
}

type bpmnHeader struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

type bpmnGateway struct {
	ID      string `xml:"id,attr"`
	Default string `xml:"default,attr"`
}

type bpmnSequenceFlow struct {
	ID        string `xml:"id,attr"`
	SourceRef string `xml:"sourceRef,attr"`
	TargetRef string `xml:"targetRef,attr"`
	Condition string `xml:"conditionExpression"`
}

// Node төрлийн ангилал.
const (
	kindStart     = "start"
	kindEnd       = "end"
	kindUserTask  = "userTask"
	kindService   = "serviceTask"
	kindGateway   = "gateway"
	kindDelegated = "delegatedTask" // алхмыг peer node гүйцэтгэнэ (delegate.peer header)
)

// flow нь нэг sequence flow (салаа сонголтод нөхцөл хэрэгтэй).
type flow struct {
	id        string
	target    string
	condition string // цэвэрлэсэн нөхцөл (хоосон = нөхцөлгүй / default)
}

// serviceCall нь serviceTask-ийн HTTP тохиргоо.
type serviceCall struct {
	method    string
	url       string
	body      string
	resultVar string
	headers   map[string]string
}

// delegateCall нь delegatedTask-ийн тохиргоо (peer руу шилжүүлэх).
type delegateCall struct {
	peer       string // зорилтот node-ийн key (registry-д бүртгэлтэй)
	processKey string // зорилтот node дээрх процессын нэр/key
}

// parsedProcess нь engine-ийн ажиллах хялбаршуулсан төлөв.
type parsedProcess struct {
	startID        string
	nodeType       map[string]string
	outgoing       map[string][]flow
	forms          map[string]json.RawMessage
	sharedForms    map[string]string // nodeID -> хуваалцсан формын ID (gerege-forms:)
	services       map[string]serviceCall
	delegates      map[string]delegateCall // nodeID -> delegatedTask тохиргоо
	gatewayDefault map[string]string       // gatewayID -> default flow id
}

// httpHeaderKeys нь serviceTask-ийн HTTP тохиргооны нөөцлөгдсөн header түлхүүрүүд.
var reservedHeaderKeys = map[string]bool{
	"method": true, "url": true, "body": true, "resultVariable": true,
}

func parseDefinition(bpmn string) (parsedProcess, error) {
	var pp parsedProcess
	if strings.TrimSpace(bpmn) == "" {
		return pp, apperror.BadRequest("invalid process definition")
	}
	var defs bpmnDefinitions
	if err := xml.Unmarshal([]byte(bpmn), &defs); err != nil {
		return pp, apperror.BadRequest("invalid bpmn xml")
	}

	pp.nodeType = map[string]string{}
	pp.outgoing = map[string][]flow{}
	pp.forms = map[string]json.RawMessage{}
	pp.sharedForms = map[string]string{}
	pp.services = map[string]serviceCall{}
	pp.delegates = map[string]delegateCall{}
	pp.gatewayDefault = map[string]string{}

	startCount, endCount := 0, 0
	for _, proc := range defs.Processes {
		formByID := make(map[string]string, len(proc.UserTaskForms))
		for _, f := range proc.UserTaskForms {
			if c := strings.TrimSpace(f.Content); c != "" && json.Valid([]byte(c)) {
				formByID[f.ID] = c
			}
		}

		for _, s := range proc.StartEvents {
			pp.nodeType[s.ID] = kindStart
			pp.startID = s.ID
			startCount++
		}
		for _, e := range proc.EndEvents {
			pp.nodeType[e.ID] = kindEnd
			endCount++
		}
		for _, u := range proc.UserTasks {
			pp.nodeType[u.ID] = kindUserTask
			fk := u.FormDefinition.FormKey
			if id := strings.TrimPrefix(fk, formKeyPrefix); id != "" && id != fk {
				// Embedded форм (энэ процессод шигтгэсэн).
				if content, ok := formByID[id]; ok {
					pp.forms[u.ID] = json.RawMessage(content)
				}
			} else if fid := strings.TrimPrefix(fk, sharedFormKeyPrefix); fid != "" && fid != fk {
				// Хуваалцсан (linked) форм — openTask үед сангаас resolve хийнэ.
				pp.sharedForms[u.ID] = fid
			}
		}
		for _, s := range proc.ServiceTasks {
			// delegate.peer header байвал энэ нь delegatedTask (peer гүйцэтгэнэ),
			// эс бөгөөс энгийн HTTP serviceTask.
			if dc, ok := delegateFromHeaders(s.Headers); ok {
				pp.nodeType[s.ID] = kindDelegated
				pp.delegates[s.ID] = dc
				continue
			}
			pp.nodeType[s.ID] = kindService
			pp.services[s.ID] = serviceFromHeaders(s.Headers)
		}
		for _, g := range proc.ExclusiveGateways {
			pp.nodeType[g.ID] = kindGateway
			pp.gatewayDefault[g.ID] = g.Default
		}
		for _, fl := range proc.SequenceFlows {
			if fl.SourceRef == "" || fl.TargetRef == "" {
				continue
			}
			pp.outgoing[fl.SourceRef] = append(pp.outgoing[fl.SourceRef], flow{
				id:        fl.ID,
				target:    fl.TargetRef,
				condition: cleanCondition(fl.Condition),
			})
		}
	}

	if startCount != 1 {
		return pp, apperror.BadRequest("process must have exactly one start event")
	}
	if endCount == 0 {
		return pp, apperror.BadRequest("process must have at least one end event")
	}
	return pp, nil
}

// delegateFromHeaders нь taskHeaders-аас delegatedTask тохиргоог (delegate.peer
// + delegate.process) гаргана. delegate.peer хоосон бол ok=false (энгийн service).
func delegateFromHeaders(headers []bpmnHeader) (delegateCall, bool) {
	var dc delegateCall
	for _, h := range headers {
		switch h.Key {
		case "delegate.peer":
			dc.peer = strings.TrimSpace(h.Value)
		case "delegate.process":
			dc.processKey = strings.TrimSpace(h.Value)
		}
	}
	if dc.peer == "" {
		return delegateCall{}, false
	}
	return dc, true
}

// serviceFromHeaders нь zeebe:taskHeaders-аас HTTP тохиргоог гаргана. method/
// url/body/resultVariable нь нөөцлөгдсөн; бусад header нь HTTP толгой болно.
func serviceFromHeaders(headers []bpmnHeader) serviceCall {
	sc := serviceCall{headers: map[string]string{}}
	for _, h := range headers {
		switch h.Key {
		case "method":
			sc.method = h.Value
		case "url":
			sc.url = h.Value
		case "body":
			sc.body = h.Value
		case "resultVariable":
			sc.resultVar = h.Value
		default:
			if !reservedHeaderKeys[h.Key] && h.Key != "" {
				sc.headers[h.Key] = h.Value
			}
		}
	}
	return sc
}

// cleanCondition нь BPMN/Zeebe нөхцөлийн илэрхийллээс `${…}` боодол эсвэл
// эхний `=`-ийг хасч, доторх илэрхийллийг буцаана.
func cleanCondition(expr string) string {
	e := strings.TrimSpace(expr)
	if e == "" {
		return ""
	}
	if strings.HasPrefix(e, "${") && strings.HasSuffix(e, "}") {
		e = strings.TrimSpace(e[2 : len(e)-1])
	} else if strings.HasPrefix(e, "=") {
		e = strings.TrimSpace(e[1:])
	}
	return e
}

// validateDefinition нь процессыг хадгалахаас өмнө шалгана.
func (u *usecase) validateDefinition(definition string) error {
	pp, err := parseDefinition(definition)
	if err != nil {
		return err
	}
	// Abuse хамгаалалт: хэт олон node-той процесс (ялангуяа serviceTask)
	// нэг хүсэлтийг олон гадагш HTTP дуудлага болгоно.
	if u.cfg.MaxNodes > 0 && len(pp.nodeType) > u.cfg.MaxNodes {
		return apperror.BadRequest("process has too many nodes")
	}
	return nil
}
