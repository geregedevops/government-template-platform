// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package requests

// AIChatRequest нь POST /ai/chat-ийн body юм. Мессежийн дээд урт нь
// токен костын хяналт + abuse хамгаалалт (нэг хүсэлтээр роман явуулахгүй).
type AIChatRequest struct {
	ConversationID string `json:"conversation_id" validate:"omitempty,uuid4"`
	Message        string `json:"message" validate:"required,min=1,max=4000"`
}

// AIKnowledgeRequest нь мэдлэгийн бичлэг үүсгэх/засах body.
type AIKnowledgeRequest struct {
	Title   string `json:"title" validate:"omitempty,max=200"`
	Content string `json:"content" validate:"required,min=1,max=8000"`
}
