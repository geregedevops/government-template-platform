// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package _interface

import (
	"context"

	"geregetemplateai/internal/business/domain"
)

// AIRepository нь AI харилцан яриа, мессеж, токен зарцуулалтын gateway юм.
// Бүх query нь RLS-тэй транзакцид ажилладаг тул энгийн хэрэглэгч зөвхөн
// өөрийн мөрүүдэд хүрнэ — usecase давхаргын WHERE болзлууд дээр нэмэлт
// хамгаалалтын давхарга болно.
type AIRepository interface {
	// CreateConversation нь шинэ харилцан яриа үүсгэж, DB-ийн үүсгэсэн
	// ID-тэй мөрийг буцаана.
	CreateConversation(ctx context.Context, userID, title string) (domain.AIConversation, error)
	// GetConversation нь ID-ээр харилцан яриаг буцаана. Байхгүй (эсвэл
	// RLS-ээр харагдахгүй) үед apperror.NotFound буцна.
	GetConversation(ctx context.Context, id string) (domain.AIConversation, error)
	// ListConversations нь хэрэглэгчийн харилцан яриануудыг шинэ нь
	// эхэндээ байхаар буцаана. Limit нь сервер талд хатуу хязгаарлагдана.
	ListConversations(ctx context.Context, userID string, offset, limit int) ([]domain.AIConversation, error)
	// StoreMessage нь мессеж нэмж, conversation-ийн updated_at-г сэргээнэ.
	StoreMessage(ctx context.Context, in *domain.AIMessage) (domain.AIMessage, error)
	// ListMessages нь харилцан ярианы мессежүүдийг хуучнаас нь шинэ рүү
	// дарааллаар буцаана (Anthropic-д шууд дамжуулах дараалал). limit <= 0
	// бол бүх мессежийг буцаана (түүхэн харагдац); эерэг утга нь СҮҮЛИЙН N-ийг
	// авна.
	ListMessages(ctx context.Context, conversationID string, limit int) ([]domain.AIMessage, error)
	// RecordUsage нь нэг AI дуудлагын токен зарцуулалтыг бичнэ.
	RecordUsage(ctx context.Context, in *domain.AIUsage) error

	// --- Knowledge base (CRUD) ---
	CreateKnowledge(ctx context.Context, in *domain.AIKnowledge) (domain.AIKnowledge, error)
	GetKnowledge(ctx context.Context, id string) (domain.AIKnowledge, error)
	ListKnowledge(ctx context.Context, userID string, offset, limit int) ([]domain.AIKnowledge, error)
	// ListAllKnowledge нь бүх хэрэглэгчийн мэдлэгийг (admin RLS) эзэмшигчийн
	// имэйлтэй нь буцаана.
	ListAllKnowledge(ctx context.Context, offset, limit int) ([]domain.AIKnowledge, error)
	UpdateKnowledge(ctx context.Context, in *domain.AIKnowledge) (domain.AIKnowledge, error)
	DeleteKnowledge(ctx context.Context, id string) error
}
