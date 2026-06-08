// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package records

import (
	"time"

	"geregetemplateai/internal/business/domain"
)

// AIConversations нь ai_conversations хүснэгтийн GORM model юм.
type AIConversations struct {
	Id        string     `gorm:"column:id;primaryKey"`
	UserId    string     `gorm:"column:user_id"`
	Title     string     `gorm:"column:title"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
}

func (AIConversations) TableName() string { return "ai_conversations" }

func (r AIConversations) ToV1Domain() domain.AIConversation {
	return domain.AIConversation{
		ID:        r.Id,
		UserID:    r.UserId,
		Title:     r.Title,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// AIMessages нь ai_messages хүснэгтийн GORM model юм.
type AIMessages struct {
	Id             string    `gorm:"column:id;primaryKey"`
	ConversationId string    `gorm:"column:conversation_id"`
	UserId         string    `gorm:"column:user_id"`
	Role           string    `gorm:"column:role"`
	Content        string    `gorm:"column:content"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (AIMessages) TableName() string { return "ai_messages" }

func (r AIMessages) ToV1Domain() domain.AIMessage {
	return domain.AIMessage{
		ID:             r.Id,
		ConversationID: r.ConversationId,
		UserID:         r.UserId,
		Role:           r.Role,
		Content:        r.Content,
		CreatedAt:      r.CreatedAt,
	}
}

// AIUsageRecords нь ai_usage хүснэгтийн GORM model юм.
type AIUsageRecords struct {
	Id             string    `gorm:"column:id;primaryKey"`
	UserId         string    `gorm:"column:user_id"`
	ConversationId string    `gorm:"column:conversation_id"`
	Model          string    `gorm:"column:model"`
	InputTokens    int       `gorm:"column:input_tokens"`
	OutputTokens   int       `gorm:"column:output_tokens"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (AIUsageRecords) TableName() string { return "ai_usage" }

// AIKnowledgeRecords нь ai_knowledge хүснэгтийн GORM model юм.
type AIKnowledgeRecords struct {
	Id        string     `gorm:"column:id;primaryKey"`
	UserId    string     `gorm:"column:user_id"`
	Title     string     `gorm:"column:title"`
	Content   string     `gorm:"column:content"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
}

func (AIKnowledgeRecords) TableName() string { return "ai_knowledge" }

func (r AIKnowledgeRecords) ToV1Domain() domain.AIKnowledge {
	return domain.AIKnowledge{
		ID:        r.Id,
		UserID:    r.UserId,
		Title:     r.Title,
		Content:   r.Content,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
