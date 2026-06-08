// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package ai нь платформын AI туслахын (Claude) business логикийг
// агуулна: харилцан яриа удирдах, мессежийн түүх ачаалах, Anthropic руу
// streaming дуудлага хийх, өдрийн хязгаар сахиулах, токен зарцуулалт
// бичих. Gemini-д суурилсан дуу хоолойн (STT/TTS, MN↔EN орчуулга)
// үйлчилгээ нь зэрэгцээ `voice` package-д хэрэгжсэн.
package ai

import (
	"context"

	"geregetemplateai/internal/business/domain"
)

// Usecase нь оролтын хил (input boundary) юм. users.Usecase-тэй ижил
// Request/Response struct загвар.
type Usecase interface {
	// Chat нь хэрэглэгчийн мессежийг харилцан ярианд нэмж, Claude-аас
	// streaming хариу авна. Текст хэсэг бүр onDelta callback-аар дамжина
	// (handler нь SSE болгон бичнэ); бүрэн хариу хадгалагдсаны дараа
	// ChatResponse буцна. ConversationID хоосон бол шинэ яриа үүсгэнэ.
	Chat(ctx context.Context, req ChatRequest, onDelta func(delta string) error) (ChatResponse, error)
	// ListConversations нь хэрэглэгчийн харилцан яриануудыг буцаана.
	ListConversations(ctx context.Context, req ListConversationsRequest) (ListConversationsResponse, error)
	// GetMessages нь нэг харилцан ярианы мессежүүдийг буцаана. Эзэмшигч
	// биш хэрэглэгчид NotFound буцна (RLS + ил шалгалт).
	GetMessages(ctx context.Context, req GetMessagesRequest) (GetMessagesResponse, error)

	// --- Knowledge base (CRUD) — чат эдгээрийг system prompt-д шигтгэнэ ---
	CreateKnowledge(ctx context.Context, req SaveKnowledgeRequest) (KnowledgeResponse, error)
	UpdateKnowledge(ctx context.Context, req UpdateKnowledgeRequest) (KnowledgeResponse, error)
	ListKnowledge(ctx context.Context, req ListKnowledgeRequest) (ListKnowledgeResponse, error)
	// ListAllKnowledge нь бүх хэрэглэгчийн мэдлэгийг буцаана (admin RLS).
	ListAllKnowledge(ctx context.Context, req ListKnowledgeRequest) (ListKnowledgeResponse, error)
	DeleteKnowledge(ctx context.Context, req DeleteKnowledgeRequest) error
}

type (
	ChatRequest struct {
		UserID         string
		ConversationID string // хоосон бол шинэ яриа үүсгэнэ
		Message        string
		Lang           string // "mn" | "en" — system prompt-д дамжина
	}
	ChatResponse struct {
		Conversation domain.AIConversation
		UserMessage  domain.AIMessage
		Reply        domain.AIMessage
		InputTokens  int
		OutputTokens int
	}

	ListConversationsRequest struct {
		UserID string
		Offset int
		Limit  int
	}
	ListConversationsResponse struct {
		Conversations []domain.AIConversation
	}

	GetMessagesRequest struct {
		UserID         string
		ConversationID string
	}
	GetMessagesResponse struct {
		Conversation domain.AIConversation
		Messages     []domain.AIMessage
	}

	SaveKnowledgeRequest struct {
		UserID  string
		Title   string
		Content string
	}
	UpdateKnowledgeRequest struct {
		UserID  string
		ID      string
		Title   string
		Content string
	}
	ListKnowledgeRequest struct {
		UserID string
		Offset int
		Limit  int
	}
	ListKnowledgeResponse struct {
		Items []domain.AIKnowledge
	}
	DeleteKnowledgeRequest struct {
		UserID string
		ID     string
	}
	KnowledgeResponse struct {
		Knowledge domain.AIKnowledge
	}
)
