// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package bpm

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeBPMRepo нь BPMRepository-ийн санах ойн (in-memory) хуурамч хувилбар —
// usecase-ийн engine логикийг (BPMN parse + start/submit advance) DB-гүйгээр
// шалгахад хангалттай. Production query-уудыг integration тест шалгана.
type fakeBPMRepo struct {
	defs      map[string]domain.BPMProcessDefinition
	instances map[string]domain.BPMProcessInstance
	tasks     map[string]domain.BPMTask
	forms     map[string]domain.BPMForm
	seq       int
}

func newFakeRepo() *fakeBPMRepo {
	return &fakeBPMRepo{
		defs:      map[string]domain.BPMProcessDefinition{},
		instances: map[string]domain.BPMProcessInstance{},
		tasks:     map[string]domain.BPMTask{},
		forms:     map[string]domain.BPMForm{},
	}
}

func (f *fakeBPMRepo) nextID(prefix string) string {
	f.seq++
	return fmt.Sprintf("%s-%d", prefix, f.seq)
}

func (f *fakeBPMRepo) CreateDefinition(_ context.Context, in *domain.BPMProcessDefinition) (domain.BPMProcessDefinition, error) {
	d := *in
	d.ID = f.nextID("def")
	d.Version = 1
	f.defs[d.ID] = d
	return d, nil
}

func (f *fakeBPMRepo) GetDefinition(_ context.Context, id string) (domain.BPMProcessDefinition, error) {
	d, ok := f.defs[id]
	if !ok {
		return domain.BPMProcessDefinition{}, apperror.NotFound("process not found")
	}
	return d, nil
}

func (f *fakeBPMRepo) GetDefinitionByName(_ context.Context, name string) (domain.BPMProcessDefinition, error) {
	for _, d := range f.defs {
		if d.Name == name {
			return d, nil
		}
	}
	return domain.BPMProcessDefinition{}, apperror.NotFound("process not found")
}

func (f *fakeBPMRepo) ListDefinitions(_ context.Context, userID string, _, _ int) ([]domain.BPMProcessDefinition, error) {
	var out []domain.BPMProcessDefinition
	for _, d := range f.defs {
		if d.UserID == userID {
			out = append(out, d)
		}
	}
	return out, nil
}

func (f *fakeBPMRepo) UpdateDefinition(_ context.Context, in *domain.BPMProcessDefinition) (domain.BPMProcessDefinition, error) {
	existing, ok := f.defs[in.ID]
	if !ok {
		return domain.BPMProcessDefinition{}, apperror.NotFound("process not found")
	}
	existing.Name = in.Name
	existing.Description = in.Description
	existing.Definition = in.Definition
	existing.Status = in.Status
	f.defs[in.ID] = existing
	return existing, nil
}

func (f *fakeBPMRepo) DeleteDefinition(_ context.Context, id string) error {
	if _, ok := f.defs[id]; !ok {
		return apperror.NotFound("process not found")
	}
	delete(f.defs, id)
	return nil
}

func (f *fakeBPMRepo) CreateInstance(_ context.Context, in *domain.BPMProcessInstance) (domain.BPMProcessInstance, error) {
	i := *in
	i.ID = f.nextID("inst")
	f.instances[i.ID] = i
	return i, nil
}

func (f *fakeBPMRepo) GetInstance(_ context.Context, id string) (domain.BPMProcessInstance, error) {
	i, ok := f.instances[id]
	if !ok {
		return domain.BPMProcessInstance{}, apperror.NotFound("instance not found")
	}
	return i, nil
}

func (f *fakeBPMRepo) ListInstances(_ context.Context, definitionID string, _, _ int) ([]domain.BPMProcessInstance, error) {
	var out []domain.BPMProcessInstance
	for _, i := range f.instances {
		if i.DefinitionID == definitionID {
			out = append(out, i)
		}
	}
	return out, nil
}

func (f *fakeBPMRepo) UpdateInstance(_ context.Context, in *domain.BPMProcessInstance) (domain.BPMProcessInstance, error) {
	if _, ok := f.instances[in.ID]; !ok {
		return domain.BPMProcessInstance{}, apperror.NotFound("instance not found")
	}
	f.instances[in.ID] = *in
	return *in, nil
}

func (f *fakeBPMRepo) CreateTask(_ context.Context, in *domain.BPMTask) (domain.BPMTask, error) {
	t := *in
	t.ID = f.nextID("task")
	t.Status = domain.BPMTaskOpen
	f.tasks[t.ID] = t
	return t, nil
}

func (f *fakeBPMRepo) GetTask(_ context.Context, id string) (domain.BPMTask, error) {
	t, ok := f.tasks[id]
	if !ok {
		return domain.BPMTask{}, apperror.NotFound("task not found")
	}
	return t, nil
}

func (f *fakeBPMRepo) GetOpenTaskByInstance(_ context.Context, instanceID string) (domain.BPMTask, error) {
	for _, t := range f.tasks {
		if t.InstanceID == instanceID && t.Status == domain.BPMTaskOpen {
			return t, nil
		}
	}
	return domain.BPMTask{}, apperror.NotFound("task not found")
}

func (f *fakeBPMRepo) CreateEvent(_ context.Context, _ *domain.BPMEvent) error { return nil }

func (f *fakeBPMRepo) ListEvents(_ context.Context, _ string, _ int) ([]domain.BPMEvent, error) {
	return nil, nil
}

func (f *fakeBPMRepo) CreateForm(_ context.Context, in *domain.BPMForm) (domain.BPMForm, error) {
	in.ID = f.nextID("form")
	f.forms[in.ID] = *in
	return *in, nil
}

func (f *fakeBPMRepo) GetForm(_ context.Context, id string) (domain.BPMForm, error) {
	form, ok := f.forms[id]
	if !ok {
		return domain.BPMForm{}, apperror.NotFound("form not found")
	}
	return form, nil
}

func (f *fakeBPMRepo) ListForms(_ context.Context, _ string, _, _ int) ([]domain.BPMForm, error) {
	out := make([]domain.BPMForm, 0, len(f.forms))
	for _, form := range f.forms {
		out = append(out, form)
	}
	return out, nil
}

func (f *fakeBPMRepo) UpdateForm(_ context.Context, in *domain.BPMForm) (domain.BPMForm, error) {
	if _, ok := f.forms[in.ID]; !ok {
		return domain.BPMForm{}, apperror.NotFound("form not found")
	}
	f.forms[in.ID] = *in
	return *in, nil
}

func (f *fakeBPMRepo) DeleteForm(_ context.Context, id string) error {
	delete(f.forms, id)
	return nil
}

func (f *fakeBPMRepo) CompleteTask(_ context.Context, id, submission string) (domain.BPMTask, error) {
	t, ok := f.tasks[id]
	if !ok || t.Status != domain.BPMTaskOpen {
		return domain.BPMTask{}, apperror.Conflict("task already completed")
	}
	t.Status = domain.BPMTaskCompleted
	sub := submission
	t.Submission = &sub
	f.tasks[id] = t
	return t, nil
}

// fakeConnector нь Connector-ийн хуурамч хувилбар — serviceTask гүйцэтгэлийг
// бодит HTTP-гүйгээр шалгана.
type fakeConnector struct {
	status   int
	body     string
	err      error
	calls    int
	lastURL  string
	lastBody string
}

func (f *fakeConnector) Do(_ context.Context, _, url string, _ map[string]string, body string) (int, []byte, error) {
	f.calls++
	f.lastURL = url
	f.lastBody = body
	if f.err != nil {
		return 0, nil, f.err
	}
	st := f.status
	if st == 0 {
		st = 200
	}
	return st, []byte(f.body), nil
}

// twoTaskBPMN нь embed хийсэн form-той стандарт .bpmn:
// start → Task_apply (form-той) → Task_review → end.
const twoTaskBPMN = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" xmlns:zeebe="http://camunda.org/schema/zeebe/1.0" id="Defs_1">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:extensionElements>
      <zeebe:userTaskForm id="Form_apply">{"type":"default","components":[{"type":"textfield","key":"name","label":"Name"}]}</zeebe:userTaskForm>
    </bpmn:extensionElements>
    <bpmn:startEvent id="Start_1"><bpmn:outgoing>F1</bpmn:outgoing></bpmn:startEvent>
    <bpmn:userTask id="Task_apply" name="Apply">
      <bpmn:extensionElements><zeebe:formDefinition formKey="camunda-forms:bpmn:Form_apply"/></bpmn:extensionElements>
      <bpmn:incoming>F1</bpmn:incoming><bpmn:outgoing>F2</bpmn:outgoing>
    </bpmn:userTask>
    <bpmn:userTask id="Task_review" name="Review">
      <bpmn:incoming>F2</bpmn:incoming><bpmn:outgoing>F3</bpmn:outgoing>
    </bpmn:userTask>
    <bpmn:endEvent id="End_1"><bpmn:incoming>F3</bpmn:incoming></bpmn:endEvent>
    <bpmn:sequenceFlow id="F1" sourceRef="Start_1" targetRef="Task_apply"/>
    <bpmn:sequenceFlow id="F2" sourceRef="Task_apply" targetRef="Task_review"/>
    <bpmn:sequenceFlow id="F3" sourceRef="Task_review" targetRef="End_1"/>
  </bpmn:process>
</bpmn:definitions>`

func TestCreateProcess_RejectsInvalidDefinition(t *testing.T) {
	uc := NewUsecase(newFakeRepo(), &fakeConnector{}, nil, nil, Config{})

	cases := map[string]string{
		"empty":          "",
		"not xml":        "not xml at all",
		"no start event": `<bpmn:definitions xmlns:bpmn="x"><bpmn:process id="p"><bpmn:endEvent id="e"/></bpmn:process></bpmn:definitions>`,
		"no end event":   `<bpmn:definitions xmlns:bpmn="x"><bpmn:process id="p"><bpmn:startEvent id="s"/></bpmn:process></bpmn:definitions>`,
		"two starts": `<bpmn:definitions xmlns:bpmn="x"><bpmn:process id="p">` +
			`<bpmn:startEvent id="s1"/><bpmn:startEvent id="s2"/><bpmn:endEvent id="e"/></bpmn:process></bpmn:definitions>`,
	}
	for name, def := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := uc.CreateProcess(context.Background(), SaveProcessRequest{UserID: "u1", Name: "X", Definition: def})
			assert.Error(t, err, "expected %s to be rejected", name)
		})
	}
}

func TestCreateProcess_AcceptsValidBPMN(t *testing.T) {
	uc := NewUsecase(newFakeRepo(), &fakeConnector{}, nil, nil, Config{})
	res, err := uc.CreateProcess(context.Background(), SaveProcessRequest{UserID: "u1", Name: "Onboarding", Definition: twoTaskBPMN})
	require.NoError(t, err)
	assert.NotEmpty(t, res.Process.ID)
	assert.Equal(t, domain.BPMStatusDraft, res.Process.Status)
}

// TestRun_StartSubmitAdvancesThroughUserTasks нь engine-ийн гол замыг түгжинэ:
// стандарт .bpmn start → Task_apply → Task_review → end. Start нь эхний
// userTask дээр зогсож, submit бүр дараагийн userTask руу шилжиж, эцэст нь
// instance дуусна. openTask нь .bpmn дотроос embed хийсэн form-js схемийг
// гаргаж snapshot болгоно.
func TestRun_StartSubmitAdvancesThroughUserTasks(t *testing.T) {
	uc := NewUsecase(newFakeRepo(), &fakeConnector{}, nil, nil, Config{})
	created, err := uc.CreateProcess(context.Background(), SaveProcessRequest{UserID: "u1", Name: "Flow", Definition: twoTaskBPMN})
	require.NoError(t, err)

	// Start: эхний userTask (Task_apply) дээр зогсоно; form snapshot хадгалагдсан.
	run, err := uc.StartInstance(context.Background(), StartInstanceRequest{UserID: "u1", DefinitionID: created.Process.ID})
	require.NoError(t, err)
	require.NotNil(t, run.Task)
	assert.Equal(t, "Task_apply", run.Task.NodeID)
	assert.Contains(t, run.Task.Form, "textfield")
	assert.Equal(t, domain.BPMInstanceRunning, run.Instance.Status)

	// Submit → Task_review руу шилжинэ.
	run2, err := uc.SubmitTask(context.Background(), SubmitTaskRequest{UserID: "u1", TaskID: run.Task.ID, Data: `{"name":"Bat"}`})
	require.NoError(t, err)
	require.NotNil(t, run2.Task)
	assert.Equal(t, "Task_review", run2.Task.NodeID)

	// Task_review submit → end → instance дуусна (Task == nil).
	run3, err := uc.SubmitTask(context.Background(), SubmitTaskRequest{UserID: "u1", TaskID: run2.Task.ID, Data: `{"approved":true}`})
	require.NoError(t, err)
	assert.Nil(t, run3.Task)
	assert.Equal(t, domain.BPMInstanceCompleted, run3.Instance.Status)

	// Хувьсагчид хоёр submit-ийн өгөгдлийг хадгалсан байх ёстой.
	assert.Contains(t, run3.Instance.Variables, "Bat")
	assert.Contains(t, run3.Instance.Variables, "approved")
}

// serviceGatewayBPMN: start → Svc (HTTP, score-г тогтооно) → Gate
// (score>=650 ? Approve : Reject) → end. serviceTask гүйцэтгэл + exclusive
// gateway-ийн нөхцөлт салаалалтыг шалгана.
const serviceGatewayBPMN = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" xmlns:zeebe="http://camunda.org/schema/zeebe/1.0" id="D">
  <bpmn:process id="P" isExecutable="true">
    <bpmn:startEvent id="S"><bpmn:outgoing>f0</bpmn:outgoing></bpmn:startEvent>
    <bpmn:serviceTask id="Svc" name="Score">
      <bpmn:extensionElements>
        <zeebe:taskHeaders>
          <zeebe:header key="method" value="POST"/>
          <zeebe:header key="url" value="https://api.example.com/score?amount=${amount}"/>
          <zeebe:header key="body" value="{}"/>
          <zeebe:header key="resultVariable" value="score"/>
        </zeebe:taskHeaders>
      </bpmn:extensionElements>
      <bpmn:incoming>f0</bpmn:incoming><bpmn:outgoing>f1</bpmn:outgoing>
    </bpmn:serviceTask>
    <bpmn:exclusiveGateway id="Gate" default="fElse">
      <bpmn:incoming>f1</bpmn:incoming><bpmn:outgoing>fApprove</bpmn:outgoing><bpmn:outgoing>fElse</bpmn:outgoing>
    </bpmn:exclusiveGateway>
    <bpmn:userTask id="Approve" name="Approve"><bpmn:incoming>fApprove</bpmn:incoming><bpmn:outgoing>fa</bpmn:outgoing></bpmn:userTask>
    <bpmn:userTask id="Reject" name="Reject"><bpmn:incoming>fElse</bpmn:incoming><bpmn:outgoing>fr</bpmn:outgoing></bpmn:userTask>
    <bpmn:endEvent id="E1"><bpmn:incoming>fa</bpmn:incoming></bpmn:endEvent>
    <bpmn:endEvent id="E2"><bpmn:incoming>fr</bpmn:incoming></bpmn:endEvent>
    <bpmn:sequenceFlow id="f0" sourceRef="S" targetRef="Svc"/>
    <bpmn:sequenceFlow id="f1" sourceRef="Svc" targetRef="Gate"/>
    <bpmn:sequenceFlow id="fApprove" sourceRef="Gate" targetRef="Approve"><bpmn:conditionExpression>${score >= 650}</bpmn:conditionExpression></bpmn:sequenceFlow>
    <bpmn:sequenceFlow id="fElse" sourceRef="Gate" targetRef="Reject"/>
    <bpmn:sequenceFlow id="fa" sourceRef="Approve" targetRef="E1"/>
    <bpmn:sequenceFlow id="fr" sourceRef="Reject" targetRef="E2"/>
  </bpmn:process>
</bpmn:definitions>`

func TestRun_ServiceTaskExecutesAndGatewayRoutes(t *testing.T) {
	// score=720 → Approve салаа.
	conn := &fakeConnector{status: 200, body: "720"}
	uc := NewUsecase(newFakeRepo(), conn, nil, nil, Config{})
	created, err := uc.CreateProcess(context.Background(), SaveProcessRequest{UserID: "u1", Name: "Loan", Definition: serviceGatewayBPMN})
	require.NoError(t, err)

	run, err := uc.StartInstance(context.Background(), StartInstanceRequest{UserID: "u1", DefinitionID: created.Process.ID})
	require.NoError(t, err)
	assert.Equal(t, 1, conn.calls, "service task must call the connector once")
	require.NotNil(t, run.Task)
	assert.Equal(t, "Approve", run.Task.NodeID, "gateway must route to Approve when score>=650")
	assert.Contains(t, run.Instance.Variables, "720", "service result stored in variables")
}

func TestRun_GatewayDefaultBranch(t *testing.T) {
	// score=500 → нөхцөл биелэхгүй → default (Reject) салаа.
	conn := &fakeConnector{status: 200, body: "500"}
	uc := NewUsecase(newFakeRepo(), conn, nil, nil, Config{})
	created, _ := uc.CreateProcess(context.Background(), SaveProcessRequest{UserID: "u1", Name: "Loan", Definition: serviceGatewayBPMN})
	run, err := uc.StartInstance(context.Background(), StartInstanceRequest{UserID: "u1", DefinitionID: created.Process.ID})
	require.NoError(t, err)
	require.NotNil(t, run.Task)
	assert.Equal(t, "Reject", run.Task.NodeID, "gateway must fall back to default when no condition matches")
}

func TestRun_ServiceTaskFailureMarksInstanceFailed(t *testing.T) {
	// HTTP 500 → service task failed → instance failed + алдаа.
	conn := &fakeConnector{status: 500, body: "oops"}
	uc := NewUsecase(newFakeRepo(), conn, nil, nil, Config{})
	created, _ := uc.CreateProcess(context.Background(), SaveProcessRequest{UserID: "u1", Name: "Loan", Definition: serviceGatewayBPMN})
	run, err := uc.StartInstance(context.Background(), StartInstanceRequest{UserID: "u1", DefinitionID: created.Process.ID})
	assert.Error(t, err, "service failure must surface as an error")
	assert.Equal(t, domain.BPMInstanceFailed, run.Instance.Status)
}

// fakeGenerator нь Generator-ийн хуурамч хувилбар — урьдчилан тогтоосон spec
// JSON буцаана.
type fakeGenerator struct{ spec string }

func (g fakeGenerator) Generate(_ context.Context, _, _ string) (string, error) {
	return g.spec, nil
}

func TestGenerateProcess_BuildsValidBPMN(t *testing.T) {
	// Claude-ийн буцаах spec (markdown fence-тэй — задлагч арилгана).
	spec := "```json\n" + `{
	  "name": "Loan",
	  "nodes": [
	    {"id":"start","type":"start"},
	    {"id":"apply","type":"form","label":"Apply","fields":[{"key":"amount","type":"number","label":"Amount"}]},
	    {"id":"score","type":"service","label":"Score","method":"GET","url":"https://api.example.com/s?a=${amount}","resultVar":"score"},
	    {"id":"gate","type":"gateway","label":"OK?"},
	    {"id":"approve","type":"form","label":"Approve","fields":[{"key":"note","type":"textarea","label":"Note"}]},
	    {"id":"end","type":"end"}
	  ],
	  "edges": [
	    {"from":"start","to":"apply"},
	    {"from":"apply","to":"score"},
	    {"from":"score","to":"gate"},
	    {"from":"gate","to":"approve","condition":"score >= 650"},
	    {"from":"gate","to":"end"}
	  ]
	}` + "\n```"
	uc := NewUsecase(newFakeRepo(), &fakeConnector{}, fakeGenerator{spec: spec}, nil, Config{AIEnabled: true})

	res, err := uc.GenerateProcess(context.Background(), GenerateProcessRequest{UserID: "u1", Description: "loan approval", Lang: "en"})
	require.NoError(t, err)
	assert.Equal(t, "Loan", res.Process.Name)
	// Үүсгэсэн нь хүчинтэй .bpmn — embed хийсэн форм/таскheader/нөхцөлтэй + DI.
	b := res.Process.Definition
	assert.True(t, strings.HasPrefix(strings.TrimSpace(b), "<?xml"))
	for _, want := range []string{"bpmn:userTask", "bpmn:serviceTask", "bpmn:exclusiveGateway",
		"zeebe:userTaskForm", "zeebe:taskHeaders", "conditionExpression", "BPMNShape", "BPMNEdge"} {
		assert.Contains(t, b, want, "generated bpmn must contain %s", want)
	}

	// Үүсгэсэн процесс engine дээр ажиллах ёстой: start → apply (form).
	run, err := uc.StartInstance(context.Background(), StartInstanceRequest{UserID: "u1", DefinitionID: res.Process.ID})
	require.NoError(t, err)
	require.NotNil(t, run.Task)
	assert.Equal(t, "apply", run.Task.NodeID)
}

func TestGenerateProcess_DisabledWhenNoAI(t *testing.T) {
	uc := NewUsecase(newFakeRepo(), &fakeConnector{}, nil, nil, Config{})
	_, err := uc.GenerateProcess(context.Background(), GenerateProcessRequest{UserID: "u1", Description: "x"})
	assert.Error(t, err, "generate must fail when AI is not configured")
}

// fakeCache нь ports.Cache-ийн хамгийн бага хувилбар — Incr нь key тус бүрийн
// тоог хадгална (өдрийн лимит тест).
type fakeCache struct{ counts map[string]int64 }

func newFakeCache() *fakeCache { return &fakeCache{counts: map[string]int64{}} }

func (c *fakeCache) Incr(_ context.Context, key string) (int64, error) {
	c.counts[key]++
	return c.counts[key], nil
}
func (c *fakeCache) Expire(context.Context, string, time.Duration) error                  { return nil }
func (c *fakeCache) Set(context.Context, string, interface{}) error                       { return nil }
func (c *fakeCache) SetWithTTL(context.Context, string, interface{}, time.Duration) error { return nil }
func (c *fakeCache) Get(context.Context, string) (string, error)                          { return "", nil }
func (c *fakeCache) GetDel(context.Context, string) (string, error)                       { return "", nil }
func (c *fakeCache) Del(context.Context, string) error                                    { return nil }
func (c *fakeCache) PTTL(context.Context, string) (time.Duration, error)                  { return 0, nil }

func TestGenerateProcess_DailyLimit(t *testing.T) {
	spec := `{"name":"X","nodes":[{"id":"s","type":"start"},{"id":"e","type":"end"}],"edges":[{"from":"s","to":"e"}]}`
	uc := NewUsecase(newFakeRepo(), &fakeConnector{}, fakeGenerator{spec: spec}, newFakeCache(),
		Config{AIEnabled: true, GenerateDailyLimit: 2})

	for i := 0; i < 2; i++ {
		_, err := uc.GenerateProcess(context.Background(), GenerateProcessRequest{UserID: "u1", Description: "d"})
		require.NoError(t, err, "call %d within limit", i+1)
	}
	_, err := uc.GenerateProcess(context.Background(), GenerateProcessRequest{UserID: "u1", Description: "d"})
	assert.Error(t, err, "3rd call must exceed the daily limit of 2")
}
