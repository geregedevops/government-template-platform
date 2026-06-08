// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package responses

import (
	"time"

	"geregetemplateai/internal/business/domain"
)

// AIConversationResponse нь харилцан ярианы DTO юм.
type AIConversationResponse struct {
	Id        string     `json:"id"`
	Title     string     `json:"title"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func FromAIConversation(c domain.AIConversation) AIConversationResponse {
	return AIConversationResponse{
		Id:        c.ID,
		Title:     c.Title,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func ToAIConversationList(convs []domain.AIConversation) []AIConversationResponse {
	out := make([]AIConversationResponse, 0, len(convs))
	for _, c := range convs {
		out = append(out, FromAIConversation(c))
	}
	return out
}

// AIMessageResponse нь нэг мессежийн DTO юм. user_id-г санаатай
// гаргахгүй — дуудагч өөрөө эзэмшигч нь.
type AIMessageResponse struct {
	Id             string    `json:"id"`
	ConversationId string    `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

func FromAIMessage(m domain.AIMessage) AIMessageResponse {
	return AIMessageResponse{
		Id:             m.ID,
		ConversationId: m.ConversationID,
		Role:           m.Role,
		Content:        m.Content,
		CreatedAt:      m.CreatedAt,
	}
}

func ToAIMessageList(msgs []domain.AIMessage) []AIMessageResponse {
	out := make([]AIMessageResponse, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, FromAIMessage(m))
	}
	return out
}

// AIKnowledgeResponse нь мэдлэгийн бичлэгийн DTO. OwnerEmail нь зөвхөн admin-ийн
// "бүх мэдлэг" жагсаалтад дүүрнэ (хэний бичлэг болохыг харуулна).
type AIKnowledgeResponse struct {
	Id         string     `json:"id"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	OwnerEmail string     `json:"owner_email,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
}

func FromAIKnowledge(k domain.AIKnowledge) AIKnowledgeResponse {
	return AIKnowledgeResponse{
		Id:         k.ID,
		Title:      k.Title,
		Content:    k.Content,
		OwnerEmail: k.OwnerEmail,
		CreatedAt:  k.CreatedAt,
		UpdatedAt:  k.UpdatedAt,
	}
}

func ToAIKnowledgeList(items []domain.AIKnowledge) []AIKnowledgeResponse {
	out := make([]AIKnowledgeResponse, 0, len(items))
	for _, k := range items {
		out = append(out, FromAIKnowledge(k))
	}
	return out
}
