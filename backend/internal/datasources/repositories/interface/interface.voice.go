// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package _interface

import (
	"context"

	"geregetemplateai/internal/business/domain"
)

// VoiceRepository нь дуу хоолойн орчуулга, токен зарцуулалтын gateway юм.
// AIRepository-тэй ижил: бүх query RLS-тэй транзакцид ажилладаг тул энгийн
// хэрэглэгч зөвхөн өөрийн мөрүүдэд хүрнэ.
type VoiceRepository interface {
	// CreateTranslation нь шинэ орчуулгын бичлэг үүсгэж, DB-ийн үүсгэсэн
	// ID-тэй мөрийг буцаана.
	CreateTranslation(ctx context.Context, in *domain.VoiceTranslation) (domain.VoiceTranslation, error)
	// ListTranslations нь хэрэглэгчийн орчуулгуудыг шинэ нь эхэндээ байхаар
	// буцаана. Limit нь сервер талд хатуу хязгаарлагдана.
	ListTranslations(ctx context.Context, userID string, offset, limit int) ([]domain.VoiceTranslation, error)
	// RecordUsage нь нэг дуу хоолойн дуудлагын токен зарцуулалтыг бичнэ.
	RecordUsage(ctx context.Context, in *domain.VoiceUsage) error
}
