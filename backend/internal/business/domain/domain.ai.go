// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package domain

import "time"

// AI chat-ийн domain entity-үүд. domain.users.go-тэй ижил зарчмаар зөвхөн
// стандарт сангаас хамаарна — HTTP, GORM, Anthropic зэрэг гадаад
// хамаарлууд энд орохгүй.

// AI мессежийн role утгууд — DB CHECK constraint болон Anthropic API-ийн
// role талбартай ЯГ таарах ёстой.
const (
	AIMessageRoleUser      = "user"
	AIMessageRoleAssistant = "assistant"
)

// AIConversation нь нэг хэрэглэгчийн нэг харилцан яриа (thread) юм.
type AIConversation struct {
	ID        string
	UserID    string
	Title     string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// AIMessage нь харилцан ярианы нэг мөр. UserID нь conversation-ийн
// эзэмшигчтэй давхардсан мэт боловч RLS бодлогыг JOIN-гүйгээр шууд
// бичих боломж олгоно (мөр бүр өөрөө эзэмшигчээ мэднэ).
type AIMessage struct {
	ID             string
	ConversationID string
	UserID         string
	Role           string
	Content        string
	CreatedAt      time.Time
}

// AIKnowledge нь хэрэглэгчийн оруулсан мэдлэгийн нэг бичлэг. AI чат нь
// эдгээрийг system prompt-д шигтгэж асуултад хариулахад ашиглана.
type AIKnowledge struct {
	ID        string
	UserID    string
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt *time.Time
	// OwnerEmail нь зөвхөн admin-ийн "бүх мэдлэг" жагсаалтад дүүргэгдэнэ
	// (хэний бичлэг болохыг харуулах). Энгийн query-д хоосон.
	OwnerEmail string
}

// AIUsage нь нэг AI дуудлагын токен зарцуулалтын бичлэг — кост хяналт,
// хэрэглэгч тус бүрийн метеринг, аудитад ашиглагдана.
type AIUsage struct {
	ID             string
	UserID         string
	ConversationID string
	Model          string
	InputTokens    int
	OutputTokens   int
	CreatedAt      time.Time
}
