// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package _interface

import (
	"context"

	"geregetemplateai/internal/business/domain"
)

// BPMRepository нь процессын тодорхойлолт, гүйлт (instance), даалгавар (task)-
// ийн gateway юм. AIRepository-тэй ижил: бүх query нь RLS-тэй транзакцид
// ажилладаг тул энгийн хэрэглэгч зөвхөн өөрийн мөрүүдэд хүрнэ.
type BPMRepository interface {
	// --- Process definitions (CRUD) ---
	CreateDefinition(ctx context.Context, in *domain.BPMProcessDefinition) (domain.BPMProcessDefinition, error)
	GetDefinition(ctx context.Context, id string) (domain.BPMProcessDefinition, error)
	// GetDefinitionByName нь нэрээр (delegatedTask-ийн зорилтот процесс) хамгийн
	// сүүлд шинэчлэгдсэн тодорхойлолтыг олно (service/admin RLS-ээр).
	GetDefinitionByName(ctx context.Context, name string) (domain.BPMProcessDefinition, error)
	ListDefinitions(ctx context.Context, userID string, offset, limit int) ([]domain.BPMProcessDefinition, error)
	UpdateDefinition(ctx context.Context, in *domain.BPMProcessDefinition) (domain.BPMProcessDefinition, error)
	DeleteDefinition(ctx context.Context, id string) error

	// --- Instances & tasks (runtime) ---
	CreateInstance(ctx context.Context, in *domain.BPMProcessInstance) (domain.BPMProcessInstance, error)
	GetInstance(ctx context.Context, id string) (domain.BPMProcessInstance, error)
	// ListInstances нь нэг процессын гүйлтүүдийг шинэ нь эхэндээ байхаар буцаана.
	ListInstances(ctx context.Context, definitionID string, offset, limit int) ([]domain.BPMProcessInstance, error)
	UpdateInstance(ctx context.Context, in *domain.BPMProcessInstance) (domain.BPMProcessInstance, error)
	CreateTask(ctx context.Context, in *domain.BPMTask) (domain.BPMTask, error)
	GetTask(ctx context.Context, id string) (domain.BPMTask, error)
	// GetOpenTaskByInstance нь instance-ийн идэвхтэй (open) даалгаврыг буцаана;
	// байхгүй бол apperror.NotFound.
	GetOpenTaskByInstance(ctx context.Context, instanceID string) (domain.BPMTask, error)
	// CompleteTask нь даалгаврыг submission-той хамт дуусгаж тэмдэглэнэ.
	CompleteTask(ctx context.Context, id, submission string) (domain.BPMTask, error)

	// --- Audit log (events) ---
	CreateEvent(ctx context.Context, in *domain.BPMEvent) error
	ListEvents(ctx context.Context, instanceID string, limit int) ([]domain.BPMEvent, error)

	// --- Shared form library (олон процесс дунд хуваалцах) ---
	CreateForm(ctx context.Context, in *domain.BPMForm) (domain.BPMForm, error)
	GetForm(ctx context.Context, id string) (domain.BPMForm, error)
	ListForms(ctx context.Context, userID string, offset, limit int) ([]domain.BPMForm, error)
	UpdateForm(ctx context.Context, in *domain.BPMForm) (domain.BPMForm, error)
	DeleteForm(ctx context.Context, id string) error
}
