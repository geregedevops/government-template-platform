// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package records

import (
	"time"

	"geregetemplateai/internal/business/domain"
)

// VoiceTranslations нь voice_translations хүснэгтийн GORM model юм.
type VoiceTranslations struct {
	Id             string    `gorm:"column:id;primaryKey"`
	UserId         string    `gorm:"column:user_id"`
	SourceLang     string    `gorm:"column:source_lang"`
	TargetLang     string    `gorm:"column:target_lang"`
	SourceText     string    `gorm:"column:source_text"`
	TranslatedText string    `gorm:"column:translated_text"`
	Model          string    `gorm:"column:model"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (VoiceTranslations) TableName() string { return "voice_translations" }

func (r VoiceTranslations) ToV1Domain() domain.VoiceTranslation {
	return domain.VoiceTranslation{
		ID:             r.Id,
		UserID:         r.UserId,
		SourceLang:     r.SourceLang,
		TargetLang:     r.TargetLang,
		SourceText:     r.SourceText,
		TranslatedText: r.TranslatedText,
		Model:          r.Model,
		CreatedAt:      r.CreatedAt,
	}
}

// VoiceUsageRecords нь voice_usage хүснэгтийн GORM model юм.
type VoiceUsageRecords struct {
	Id            string    `gorm:"column:id;primaryKey"`
	UserId        string    `gorm:"column:user_id"`
	TranslationId string    `gorm:"column:translation_id"`
	Model         string    `gorm:"column:model"`
	InputTokens   int       `gorm:"column:input_tokens"`
	OutputTokens  int       `gorm:"column:output_tokens"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (VoiceUsageRecords) TableName() string { return "voice_usage" }

func (r VoiceUsageRecords) ToV1Domain() domain.VoiceUsage {
	return domain.VoiceUsage{
		ID:            r.Id,
		UserID:        r.UserId,
		TranslationID: r.TranslationId,
		Model:         r.Model,
		InputTokens:   r.InputTokens,
		OutputTokens:  r.OutputTokens,
		CreatedAt:     r.CreatedAt,
	}
}
