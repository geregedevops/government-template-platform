// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package bpm нь Business Process Management-ийн business логикийг агуулна:
// процессын тодорхойлолтыг үүсгэх/засах/хадгалах (CRUD), графикийг шалгах
// (validate), гүйлт (instance) эхлүүлж хэрэглэгчийн даалгавар (form task)-ийг
// бөглүүлэх. ai/voice usecase-тэй ижил Request/Response struct загвар.
//
// Энэ foundation slice нь form node-уудыг шугаман дагуулна; apiCall/decision
// node-ийн жинхэнэ гүйцэтгэл дараагийн increment-д нэмэгдэнэ.
package bpm

import (
	"context"
	"encoding/json"

	"geregetemplateai/internal/business/domain"
)

// Usecase нь оролтын хил (input boundary) юм.
type Usecase interface {
	// --- Process definitions (CRUD) ---
	CreateProcess(ctx context.Context, req SaveProcessRequest) (ProcessResponse, error)
	// GenerateProcess нь текст тайлбараас Claude-аар BPMN процесс үүсгэж
	// (маягт/service/нөхцөлтэй), хадгалаад буцаана.
	GenerateProcess(ctx context.Context, req GenerateProcessRequest) (ProcessResponse, error)
	UpdateProcess(ctx context.Context, req UpdateProcessRequest) (ProcessResponse, error)
	GetProcess(ctx context.Context, req GetProcessRequest) (ProcessResponse, error)
	ListProcesses(ctx context.Context, req ListProcessesRequest) (ListProcessesResponse, error)
	DeleteProcess(ctx context.Context, req GetProcessRequest) error

	// --- Runtime ---
	// StartInstance нь процессын шинэ гүйлт эхлүүлж, эхний form node дээр
	// идэвхтэй task нээнэ (эсвэл form байхгүй бол шууд дуусгана).
	StartInstance(ctx context.Context, req StartInstanceRequest) (RunResponse, error)
	// GetActiveTask нь instance-ийн идэвхтэй даалгаврыг (рендерлэх дэлгэц)
	// буцаана.
	GetActiveTask(ctx context.Context, req GetActiveTaskRequest) (RunResponse, error)
	// ListInstances нь нэг процессын гүйлтүүдийг (мониторинг) буцаана.
	ListInstances(ctx context.Context, req ListInstancesRequest) (ListInstancesResponse, error)
	// ListEvents нь нэг гүйлтийн audit log-ийн timeline-г буцаана.
	ListEvents(ctx context.Context, req ListEventsRequest) (ListEventsResponse, error)
	// SubmitTask нь даалгаврын хариуг хадгалж, дараагийн form node руу
	// шилжинэ (эсвэл процессыг дуусгана).
	SubmitTask(ctx context.Context, req SubmitTaskRequest) (RunResponse, error)

	// --- Shared form library (олон процесс дунд хуваалцах) ---
	CreateForm(ctx context.Context, req SaveFormRequest) (FormResponse, error)
	UpdateForm(ctx context.Context, req SaveFormRequest) (FormResponse, error)
	ListForms(ctx context.Context, req ListFormsRequest) (ListFormsResponse, error)
	DeleteForm(ctx context.Context, req GetFormRequest) error

	// --- Федераци (delegatedTask; ROADMAP Үе 1) ---
	SetFedSender(s FedSender)
	StartDelegated(ctx context.Context, processKey string, vars json.RawMessage, originPeer, parentInstance string) error
	ResumeDelegated(ctx context.Context, parentInstance string, vars json.RawMessage, status string) error
}

type (
	SaveProcessRequest struct {
		UserID      string
		OrgID       string // үүсгэгчийн байгууллага (org-scoped хандалт)
		Name        string
		Description string
		Definition  string // ProcessGraph JSON
		Status      string // "draft" | "published"; хоосон бол "draft"
	}
	UpdateProcessRequest struct {
		UserID      string
		ID          string
		Name        string
		Description string
		Definition  string
		Status      string
	}
	GetProcessRequest struct {
		UserID string
		ID     string
	}
	GenerateProcessRequest struct {
		UserID      string
		OrgID       string // үүсгэгчийн байгууллага
		Description string
		Lang        string // "mn" | "en" — prompt-д дамжина
	}
	ListProcessesRequest struct {
		UserID string
		Offset int
		Limit  int
	}
	ListProcessesResponse struct {
		Processes []domain.BPMProcessDefinition
	}
	ProcessResponse struct {
		Process domain.BPMProcessDefinition
	}

	StartInstanceRequest struct {
		UserID       string
		DefinitionID string
	}
	GetActiveTaskRequest struct {
		UserID     string
		InstanceID string
	}
	ListInstancesRequest struct {
		UserID       string
		DefinitionID string
		Offset       int
		Limit        int
	}
	ListInstancesResponse struct {
		Instances []domain.BPMProcessInstance
	}
	ListEventsRequest struct {
		UserID     string
		InstanceID string
	}
	ListEventsResponse struct {
		Events []domain.BPMEvent
	}
	SubmitTaskRequest struct {
		UserID string
		TaskID string
		Data   string // хэрэглэгчийн бөглөсөн хариу (JSON object)
	}
	// RunResponse нь гүйлтийн нэг алхмын төлөв. Task нь nil бол instance
	// дууссан (өөр бөглөх дэлгэц алга).
	RunResponse struct {
		Instance domain.BPMProcessInstance
		Task     *domain.BPMTask
	}

	// SaveFormRequest нь хуваалцсан форм үүсгэх/засах. ID хоосон ⇒ үүсгэнэ.
	SaveFormRequest struct {
		UserID string
		ID     string
		Name   string
		Schema string // form-js схем (JSON)
	}
	GetFormRequest struct {
		UserID string
		ID     string
	}
	ListFormsRequest struct {
		UserID string
		Offset int
		Limit  int
	}
	ListFormsResponse struct {
		Forms []domain.BPMForm
	}
	FormResponse struct {
		Form domain.BPMForm
	}
)
