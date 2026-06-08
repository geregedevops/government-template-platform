// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package bpm

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
	"time"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/domain"
	"geregetemplateai/pkg/logger"
)

// GenerateProcess нь текст тайлбараас Claude-аар процессын JSON spec гаргуулж,
// түүнийг стандарт .bpmn (маягт/service/нөхцөл embed + DI зураглал) болгон
// хөрвүүлж хадгална. Хэрэглэгч modeler-т нээж засаж болно.
func (u *usecase) GenerateProcess(ctx context.Context, req GenerateProcessRequest) (ProcessResponse, error) {
	if !u.cfg.AIEnabled || u.generator == nil {
		return ProcessResponse{}, apperror.Unavailable("ai service is not configured")
	}
	desc := strings.TrimSpace(req.Description)
	if desc == "" {
		return ProcessResponse{}, apperror.BadRequest("description is required")
	}

	// Хэрэглэгч тус бүрийн өдрийн лимит (AI кост хяналт). ai.chat.go-той ижил:
	// эхний хүсэлтэд 25 цагийн TTL; Redis саатвал нээлттэй бүтэлгүйтэж зөвшөөрнө.
	if u.cfg.GenerateDailyLimit > 0 && u.cache != nil {
		key := GenerateDailyCountKey(req.UserID, time.Now())
		count, err := u.cache.Incr(ctx, key)
		if err != nil {
			logger.WarnWithContext(ctx, "bpm generate daily counter unavailable, allowing request", logger.Fields{
				"usecase": "bpm", "method": "GenerateProcess", "error": err.Error(),
			})
		} else {
			if count == 1 {
				if err := u.cache.Expire(ctx, key, 25*time.Hour); err != nil {
					logger.WarnWithContext(ctx, "bpm generate daily counter: failed to set TTL", logger.Fields{
						"usecase": "bpm", "method": "GenerateProcess", "error": err.Error(),
					})
				}
			}
			if count > int64(u.cfg.GenerateDailyLimit) {
				return ProcessResponse{}, apperror.Forbidden("ai daily request limit exceeded")
			}
		}
	}

	raw, err := u.generator.Generate(ctx, generateSystemPrompt(req.Lang), desc)
	if err != nil {
		return ProcessResponse{}, apperror.Wrap(apperror.Unavailable("ai generation failed"), err)
	}

	spec, err := parseSpec(raw)
	if err != nil {
		snippet := raw
		if len(snippet) > 1200 {
			snippet = snippet[:1200]
		}
		logger.WarnWithContext(ctx, "bpm generate: spec parse failed", logger.Fields{
			"usecase": "bpm", "method": "GenerateProcess", "error": err.Error(),
			"raw_len": len(raw), "raw_head": snippet,
		})
		return ProcessResponse{}, err
	}
	bpmn, err := buildBPMN(spec)
	if err != nil {
		return ProcessResponse{}, err
	}
	// Үүсгэсэн BPMN өөрөө хүчинтэй эсэхийг engine-ийн parser-аар шалгана.
	if _, err := parseDefinition(bpmn); err != nil {
		return ProcessResponse{}, apperror.BadRequest("generated process is invalid")
	}

	name := strings.TrimSpace(spec.Name)
	if name == "" {
		name = "AI process"
	}
	process, err := u.repo.CreateDefinition(ctx, &domain.BPMProcessDefinition{
		UserID:      req.UserID,
		OrgID:       req.OrgID,
		Name:        name,
		Description: desc,
		Definition:  bpmn,
		Status:      domain.BPMStatusDraft,
	})
	if err != nil {
		return ProcessResponse{}, mapRepoError(err, "create process")
	}
	return ProcessResponse{Process: process}, nil
}

// --- AI spec ----------------------------------------------------------------

type genSpec struct {
	Name  string    `json:"name"`
	Nodes []genNode `json:"nodes"`
	Edges []genEdge `json:"edges"`
}

type genNode struct {
	ID     string     `json:"id"`
	Type   string     `json:"type"` // start|end|form|service|gateway
	Label  string     `json:"label"`
	Fields []genField `json:"fields,omitempty"`
	Method string     `json:"method,omitempty"`
	URL    string     `json:"url,omitempty"`
	// Body нь JSON object эсвэл string аль аль нь байж болно (Claude ихэвчлэн
	// object буцаадаг) тул RawMessage-ээр хүлээж авна.
	Body      json.RawMessage `json:"body,omitempty"`
	ResultVar string          `json:"resultVar,omitempty"`
}

// bodyString нь genNode.Body-г taskHeaders-ийн body утга (JSON template string)
// болгоно. Claude string өгсөн бол задлан авна; object/array бол түүхийгээр нь.
func bodyString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

type genField struct {
	Key      string   `json:"key"`
	Type     string   `json:"type"`
	Label    string   `json:"label"`
	Required bool     `json:"required,omitempty"`
	Options  []string `json:"options,omitempty"`
}

type genEdge struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Condition string `json:"condition,omitempty"`
}

// parseSpec нь Claude-ийн хариунаас JSON spec-ийг гаргаж задална (markdown
// ```json``` хашилт байвал арилгана).
func parseSpec(raw string) (genSpec, error) {
	var spec genSpec
	s := strings.TrimSpace(raw)
	// markdown fence-ийг арилгах.
	if i := strings.Index(s, "```"); i >= 0 {
		s = s[i+3:]
		s = strings.TrimPrefix(s, "json")
		if j := strings.LastIndex(s, "```"); j >= 0 {
			s = s[:j]
		}
	}
	// эхний '{' ба сүүлийн '}' хооронд авах (нэмэлт текст байвал).
	if a, b := strings.Index(s, "{"), strings.LastIndex(s, "}"); a >= 0 && b > a {
		s = s[a : b+1]
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(s)), &spec); err != nil {
		return spec, apperror.BadRequest("ai returned invalid spec")
	}
	if len(spec.Nodes) == 0 {
		return spec, apperror.BadRequest("ai returned an empty process")
	}
	return spec, nil
}

// generateSystemPrompt нь Claude-д ЗӨВХӨН тодорхой схемийн JSON гаргахыг
// зааварлана.
func generateSystemPrompt(lang string) string {
	langLine := "Use the language of the user's description for all labels and field labels."
	if lang == "mn" {
		langLine = "Шошго (label)-уудыг хэрэглэгчийн тайлбарын хэлээр (Монголоор) бич."
	}
	return `You design business processes as BPMN. The user describes a process in natural language.
Return ONLY a JSON object (no prose, no markdown) with this exact shape:

{
  "name": "short process name",
  "nodes": [
    {"id":"start","type":"start"},
    {"id":"apply","type":"form","label":"Application","fields":[
      {"key":"fullName","type":"text","label":"Full name","required":true},
      {"key":"amount","type":"number","label":"Amount"}
    ]},
    {"id":"check","type":"service","label":"Check","method":"GET","url":"https://api.example.com/check?amount=${amount}","resultVar":"result"},
    {"id":"decide","type":"gateway","label":"Approved?"},
    {"id":"approved","type":"form","label":"Approval","fields":[{"key":"note","type":"textarea","label":"Note"}]},
    {"id":"end","type":"end"}
  ],
  "edges": [
    {"from":"start","to":"apply"},
    {"from":"apply","to":"check"},
    {"from":"check","to":"decide"},
    {"from":"decide","to":"approved","condition":"amount >= 1000"},
    {"from":"decide","to":"end"}
  ]
}

Rules:
- Node "type" is one of: start, end, form, service, gateway.
- Exactly one "start" node; at least one "end" node.
- "form" nodes collect input from a person; give them "fields". field "type" is one of: text, textarea, number, email, select, checkbox, date. For "select", add "options": ["A","B"].
- "service" nodes call an external API; set method (GET/POST/...), url, optional body (JSON), and resultVar to store the response. You may use ${variableKey} placeholders in url/body referencing earlier field keys or service resultVars.
- "gateway" nodes branch. Each outgoing edge from a gateway should have a "condition" like "amount >= 1000" or "status == \"ok\""; leave ONE outgoing edge without a condition as the default branch.
- Every node must be reachable from start and lead toward an end. Use clear short ids (letters, digits, underscores).
- ` + langLine + `
Return only the JSON.`
}

// --- BPMN бүтээгч (spec -> .bpmn + DI зураглал) ------------------------------

const formKeyGenPrefix = "camunda-forms:bpmn:"

type elemBox struct {
	w, h int
	cx   int // төвийн x
	top  int // дээд y
}

// buildBPMN нь spec-ийг стандарт BPMN 2.0 XML болгож (форм/service/нөхцөл
// embed + энгийн босоо DI зураглалтай) хөрвүүлнэ.
func buildBPMN(spec genSpec) (string, error) {
	if len(spec.Nodes) == 0 {
		return "", apperror.BadRequest("ai returned an empty process")
	}
	ids := map[string]genNode{}
	for _, n := range spec.Nodes {
		if n.ID == "" {
			return "", apperror.BadRequest("generated node has no id")
		}
		ids[n.ID] = n
	}
	// edge-ийн заагчийг шалгах.
	for _, e := range spec.Edges {
		if _, ok := ids[e.From]; !ok {
			return "", apperror.BadRequest("generated edge references unknown node")
		}
		if _, ok := ids[e.To]; !ok {
			return "", apperror.BadRequest("generated edge references unknown node")
		}
	}

	// incoming/outgoing flow id-ууд.
	flowID := func(i int) string { return fmt.Sprintf("Flow_%d", i+1) }
	outFlows := map[string][]string{} // nodeID -> []flowID (гарах)
	inFlows := map[string][]string{}  // nodeID -> []flowID (орох)
	for i, e := range spec.Edges {
		fid := flowID(i)
		outFlows[e.From] = append(outFlows[e.From], fid)
		inFlows[e.To] = append(inFlows[e.To], fid)
	}

	layout := computeLayout(spec, ids)

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" ` +
		`xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI" ` +
		`xmlns:dc="http://www.omg.org/spec/DD/20100524/DC" ` +
		`xmlns:di="http://www.omg.org/spec/DD/20100524/DI" ` +
		`xmlns:zeebe="http://camunda.org/schema/zeebe/1.0" ` +
		`id="Definitions_gen" targetNamespace="http://bpmn.io/schema/bpmn">` + "\n")
	b.WriteString(`  <bpmn:process id="Process_gen" isExecutable="true">` + "\n")

	// process-level extensionElements: form node бүрийн userTaskForm.
	formNodes := []genNode{}
	for _, n := range spec.Nodes {
		if n.Type == "form" && len(n.Fields) > 0 {
			formNodes = append(formNodes, n)
		}
	}
	if len(formNodes) > 0 {
		b.WriteString(`    <bpmn:extensionElements>` + "\n")
		for _, n := range formNodes {
			b.WriteString(fmt.Sprintf(`      <zeebe:userTaskForm id="UserTaskForm_%s">%s</zeebe:userTaskForm>`+"\n",
				xmlAttr(n.ID), xmlText(formSchema(n.Fields))))
		}
		b.WriteString(`    </bpmn:extensionElements>` + "\n")
	}

	// node элементүүд.
	for _, n := range spec.Nodes {
		writeNode(&b, n, inFlows[n.ID], outFlows[n.ID])
	}

	// sequenceFlow-ууд.
	for i, e := range spec.Edges {
		fid := flowID(i)
		if c := strings.TrimSpace(e.Condition); c != "" {
			b.WriteString(fmt.Sprintf(`    <bpmn:sequenceFlow id="%s" sourceRef="%s" targetRef="%s">`+"\n"+
				`      <bpmn:conditionExpression>${%s}</bpmn:conditionExpression>`+"\n"+
				`    </bpmn:sequenceFlow>`+"\n",
				fid, xmlAttr(e.From), xmlAttr(e.To), xmlText(c)))
		} else {
			b.WriteString(fmt.Sprintf(`    <bpmn:sequenceFlow id="%s" sourceRef="%s" targetRef="%s"/>`+"\n",
				fid, xmlAttr(e.From), xmlAttr(e.To)))
		}
	}
	b.WriteString(`  </bpmn:process>` + "\n")

	// DI зураглал.
	b.WriteString(`  <bpmndi:BPMNDiagram id="BPMNDiagram_gen">` + "\n")
	b.WriteString(`    <bpmndi:BPMNPlane id="BPMNPlane_gen" bpmnElement="Process_gen">` + "\n")
	for _, n := range spec.Nodes {
		box := layout[n.ID]
		b.WriteString(fmt.Sprintf(`      <bpmndi:BPMNShape id="Shape_%s" bpmnElement="%s">`+"\n"+
			`        <dc:Bounds x="%d" y="%d" width="%d" height="%d"/>`+"\n"+
			`      </bpmndi:BPMNShape>`+"\n",
			xmlAttr(n.ID), xmlAttr(n.ID), box.cx-box.w/2, box.top, box.w, box.h))
	}
	for i, e := range spec.Edges {
		s, t := layout[e.From], layout[e.To]
		b.WriteString(fmt.Sprintf(`      <bpmndi:BPMNEdge id="Edge_%s" bpmnElement="%s">`+"\n"+
			`        <di:waypoint x="%d" y="%d"/>`+"\n"+
			`        <di:waypoint x="%d" y="%d"/>`+"\n"+
			`      </bpmndi:BPMNEdge>`+"\n",
			flowID(i), flowID(i), s.cx, s.top+s.h, t.cx, t.top))
	}
	b.WriteString(`    </bpmndi:BPMNPlane>` + "\n")
	b.WriteString(`  </bpmndi:BPMNDiagram>` + "\n")
	b.WriteString(`</bpmn:definitions>` + "\n")
	return b.String(), nil
}

// writeNode нь нэг node-ийн BPMN элементийг бичнэ.
func writeNode(b *strings.Builder, n genNode, in, out []string) {
	inOut := func() string {
		var s strings.Builder
		for _, f := range in {
			s.WriteString(fmt.Sprintf("      <bpmn:incoming>%s</bpmn:incoming>\n", f))
		}
		for _, f := range out {
			s.WriteString(fmt.Sprintf("      <bpmn:outgoing>%s</bpmn:outgoing>\n", f))
		}
		return s.String()
	}
	id, name := xmlAttr(n.ID), xmlAttr(n.Label)
	switch n.Type {
	case "start":
		b.WriteString(fmt.Sprintf("    <bpmn:startEvent id=\"%s\" name=\"%s\">\n%s    </bpmn:startEvent>\n", id, name, inOut()))
	case "end":
		b.WriteString(fmt.Sprintf("    <bpmn:endEvent id=\"%s\" name=\"%s\">\n%s    </bpmn:endEvent>\n", id, name, inOut()))
	case "gateway":
		b.WriteString(fmt.Sprintf("    <bpmn:exclusiveGateway id=\"%s\" name=\"%s\">\n%s    </bpmn:exclusiveGateway>\n", id, name, inOut()))
	case "service":
		b.WriteString(fmt.Sprintf("    <bpmn:serviceTask id=\"%s\" name=\"%s\">\n", id, name))
		b.WriteString("      <bpmn:extensionElements>\n        <zeebe:taskHeaders>\n")
		method := n.Method
		if method == "" {
			method = "GET"
		}
		b.WriteString(fmt.Sprintf("          <zeebe:header key=\"method\" value=\"%s\"/>\n", xmlAttr(method)))
		b.WriteString(fmt.Sprintf("          <zeebe:header key=\"url\" value=\"%s\"/>\n", xmlAttr(n.URL)))
		if bs := bodyString(n.Body); strings.TrimSpace(bs) != "" {
			b.WriteString(fmt.Sprintf("          <zeebe:header key=\"body\" value=\"%s\"/>\n", xmlAttr(bs)))
		}
		if strings.TrimSpace(n.ResultVar) != "" {
			b.WriteString(fmt.Sprintf("          <zeebe:header key=\"resultVariable\" value=\"%s\"/>\n", xmlAttr(n.ResultVar)))
		}
		b.WriteString("        </zeebe:taskHeaders>\n      </bpmn:extensionElements>\n")
		b.WriteString(inOut())
		b.WriteString("    </bpmn:serviceTask>\n")
	default: // form (user task)
		b.WriteString(fmt.Sprintf("    <bpmn:userTask id=\"%s\" name=\"%s\">\n", id, name))
		if len(n.Fields) > 0 {
			b.WriteString(fmt.Sprintf("      <bpmn:extensionElements>\n        <zeebe:formDefinition formKey=\"%s%s\"/>\n      </bpmn:extensionElements>\n",
				formKeyGenPrefix, "UserTaskForm_"+xmlAttr(n.ID)))
		}
		b.WriteString(inOut())
		b.WriteString("    </bpmn:userTask>\n")
	}
}

// formSchema нь spec-ийн талбаруудыг form-js схем (JSON) болгоно.
func formSchema(fields []genField) string {
	comps := make([]map[string]interface{}, 0, len(fields))
	for i, f := range fields {
		c := map[string]interface{}{
			"type":  formJsType(f.Type),
			"key":   f.Key,
			"label": f.Label,
			"id":    "Field_" + f.Key,
			// 2 баган: талбар бүр 16 нэгжийн grid-ийн 8-ыг эзэлж, хос хосоор
			// нэг мөрөнд (textarea-г бүтэн өргөнөөр).
			"layout": fieldLayout(i, f.Type),
		}
		if f.Required {
			c["validate"] = map[string]interface{}{"required": true}
		}
		if f.Type == "select" && len(f.Options) > 0 {
			vals := make([]map[string]string, 0, len(f.Options))
			for _, o := range f.Options {
				vals = append(vals, map[string]string{"label": o, "value": o})
			}
			c["values"] = vals
		}
		comps = append(comps, c)
	}
	schema := map[string]interface{}{"type": "default", "components": comps}
	out, _ := json.Marshal(schema)
	return string(out)
}

// fieldLayout нь form-js-ийн 2 баган зохион байгуулалтыг буцаана. textarea нь
// бүтэн өргөн (16), бусад нь хагас (8) — индексээр хослуулна.
func fieldLayout(i int, fieldType string) map[string]interface{} {
	if fieldType == "textarea" {
		return map[string]interface{}{"row": fmt.Sprintf("row_%d", i), "columns": 16}
	}
	return map[string]interface{}{"row": fmt.Sprintf("row_%d", i/2), "columns": 8}
}

func formJsType(t string) string {
	switch t {
	case "textarea":
		return "textarea"
	case "number":
		return "number"
	case "checkbox":
		return "checkbox"
	case "select":
		return "select"
	case "date":
		return "datetime"
	default:
		return "textfield" // text, email
	}
}

// computeLayout нь node бүрд энгийн босоо зураглалын байрлал тооцно: start-аас
// хол байх зэрэг (rank)-аар доош, нэг зэргийн node-уудыг хэвтээгээр тараана.
func computeLayout(spec genSpec, ids map[string]genNode) map[string]elemBox {
	rank := map[string]int{}
	for id := range ids {
		rank[id] = 0
	}
	// longest-path relaxation (DAG бус ч len(nodes) давталтаар тогтворжино).
	for i := 0; i < len(spec.Nodes); i++ {
		for _, e := range spec.Edges {
			if rank[e.To] < rank[e.From]+1 {
				rank[e.To] = rank[e.From] + 1
			}
		}
	}
	// зэрэг тус бүрийн node-ууд.
	byRank := map[int][]string{}
	maxRank := 0
	for _, n := range spec.Nodes {
		r := rank[n.ID]
		byRank[r] = append(byRank[r], n.ID)
		if r > maxRank {
			maxRank = r
		}
	}

	const (
		centerX = 400
		topY    = 80
		rankGap = 130
		laneGap = 200
	)
	box := map[string]elemBox{}
	for r := 0; r <= maxRank; r++ {
		lane := byRank[r]
		sort.Strings(lane) // тогтвортой дараалал
		for i, id := range lane {
			w, h := sizeOf(ids[id].Type)
			cx := centerX + (i-(len(lane)-1)/2)*laneGap
			top := topY + r*rankGap
			box[id] = elemBox{w: w, h: h, cx: cx, top: top}
		}
	}
	return box
}

func sizeOf(t string) (int, int) {
	switch t {
	case "start", "end":
		return 36, 36
	case "gateway":
		return 50, 50
	default: // form, service
		return 100, 80
	}
}

// xmlAttr / xmlText нь утгуудыг XML-д аюулгүй болгоно.
func xmlAttr(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

func xmlText(s string) string { return xmlAttr(s) }
