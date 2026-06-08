// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package ai

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/domain"
	"geregetemplateai/pkg/aiclient"
	"geregetemplateai/pkg/logger"
)

// titleMaxRunes нь автоматаар үүсгэх ярианы гарчгийн дээд урт.
const titleMaxRunes = 60

// Chat нь нэг AI чат хүсэлтийг бүрэн гүйцэтгэнэ:
//
//  1. өдрийн хязгаар шалгах (Redis тоолуур, fail-open биш — Redis алдаа
//     дээр хүсэлтийг нэвтрүүлнэ, учир нь чат нь аюулгүй байдлын хил биш);
//  2. харилцан яриа олох/үүсгэх (эзэмшил шалгана);
//  3. хэрэглэгчийн мессежийг хадгалах;
//  4. түүхийг ачаалж Claude-г streaming-ээр дуудах;
//  5. бүрэн хариуг хадгалж, токен зарцуулалт бичих.
//
// Streaming үед хэсэгчилсэн алдаа гарвал (Claude тасрах г.м.) аль хэдийн
// гарсан текстийг хадгалахыг оролдоно — хэрэглэгчийн харсан зүйл түүхэнд
// үлдэнэ.
func (u *usecase) Chat(ctx context.Context, req ChatRequest, onDelta func(delta string) error) (ChatResponse, error) {
	if !u.cfg.Enabled {
		return ChatResponse{}, apperror.Unavailable("ai service is not configured")
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return ChatResponse{}, apperror.BadRequest("message is required")
	}

	// (1) Өдрийн хязгаар. Incr нь түлхүүр байхгүй үед 1-ээс эхэлдэг тул
	// эхний хүсэлт дээр TTL тавина (өдрийн төгсгөл хүртэл биш — 25 цаг нь
	// хангалттай: түлхүүр өдрөөр нэрлэгддэг тул дараагийн өдөр шинэ түлхүүр).
	if u.cfg.DailyRequestLimit > 0 && u.cache != nil {
		key := DailyCountKey(req.UserID, time.Now())
		count, err := u.cache.Incr(ctx, key)
		if err != nil {
			// Redis-ийн түр саатал чатыг бүхэлд нь унагах ёсгүй — нээлттэй
			// бүтэлгүйтнэ (auth middleware-ийн cutoff шалгалттай ижил шийдвэр).
			logger.WarnWithContext(ctx, "ai daily counter unavailable, allowing request", logger.Fields{
				"usecase": "ai",
				"method":  "Chat",
				"error":   err.Error(),
			})
		} else {
			if count == 1 {
				// TTL суулгаж чадахгүй бол түлхүүр мөнхөрч хэрэглэгчийг
				// байнга хязгаарлах эрсдэлтэй тул алдааг чимээгүй залгилгүй
				// log хийнэ (Incr-тэй адил Redis саатлыг үл унагана).
				if err := u.cache.Expire(ctx, key, 25*time.Hour); err != nil {
					logger.WarnWithContext(ctx, "ai daily counter: failed to set TTL", logger.Fields{
						"usecase": "ai",
						"method":  "Chat",
						"error":   err.Error(),
					})
				}
			}
			if count > int64(u.cfg.DailyRequestLimit) {
				return ChatResponse{}, apperror.Forbidden("ai daily request limit exceeded")
			}
		}
	}

	// (2) Харилцан яриа.
	var conv domain.AIConversation
	var err error
	if req.ConversationID == "" {
		conv, err = u.repo.CreateConversation(ctx, req.UserID, makeTitle(message))
		if err != nil {
			return ChatResponse{}, mapRepoError(err, "create conversation")
		}
	} else {
		conv, err = u.repo.GetConversation(ctx, req.ConversationID)
		if err != nil {
			return ChatResponse{}, mapRepoError(err, "get conversation")
		}
		// RLS энгийн хэрэглэгчийг аль хэдийн хязгаарладаг ч admin context
		// болон ирээдүйн өөрчлөлтөөс хамгаалж эзэмшлийг ил шалгана.
		if conv.UserID != req.UserID {
			return ChatResponse{}, apperror.NotFound("conversation not found")
		}
	}

	// (3) Түүхийг ХЭРЭГЛЭГЧИЙН ШИНЭ МЕССЕЖЭЭС ӨМНӨ ачаална — давхардал
	// орохгүй; дараа нь шинэ мессежийг төгсгөлд нь залгана.
	history, err := u.repo.ListMessages(ctx, conv.ID, u.cfg.HistoryLimit)
	if err != nil {
		return ChatResponse{}, mapRepoError(err, "list messages")
	}

	userMsg, err := u.repo.StoreMessage(ctx, &domain.AIMessage{
		ConversationID: conv.ID,
		UserID:         req.UserID,
		Role:           domain.AIMessageRoleUser,
		Content:        message,
	})
	if err != nil {
		return ChatResponse{}, mapRepoError(err, "store user message")
	}

	// (4) Claude streaming дуудлага. Хэрэглэгчийн мэдлэгийн бичлэгүүдийг
	// system prompt-д шигтгэж, туслах тэдгээрийг ашиглан хариулна. Мэдлэг
	// татах алдаа чатыг унагах ёсгүй — хоосон жагсаалтаар үргэлжилнэ.
	knowledge, kErr := u.repo.ListKnowledge(ctx, req.UserID, 0, 0)
	if kErr != nil {
		knowledge = nil
	}

	msgs := make([]aiclient.Message, 0, len(history)+1)
	for _, h := range history {
		msgs = append(msgs, aiclient.Message{Role: h.Role, Content: h.Content})
	}
	msgs = append(msgs, aiclient.Message{Role: domain.AIMessageRoleUser, Content: message})

	full, usage, streamErr := u.streamer.StreamMessage(ctx, buildSystemPrompt(req.Lang, knowledge), msgs, onDelta)

	// (5) Хариу хадгалах + метеринг. Хэсэгчилсэн хариу ч хадгалагдана.
	var reply domain.AIMessage
	if strings.TrimSpace(full) != "" {
		reply, err = u.repo.StoreMessage(ctx, &domain.AIMessage{
			ConversationID: conv.ID,
			UserID:         req.UserID,
			Role:           domain.AIMessageRoleAssistant,
			Content:        full,
		})
		if err != nil {
			// Хариу хэрэглэгч рүү аль хэдийн урссан — хадгалалтын алдааг
			// логдоод цааш явна (хэрэглэгчид streaming-ийн дараа 500 өгөх нь
			// илүү муу туршлага).
			logger.ErrorWithContext(ctx, "failed to persist assistant reply", logger.Fields{
				"usecase":         "ai",
				"method":          "Chat",
				"conversation_id": conv.ID,
				"error":           err.Error(),
			})
		}
	}
	if usage.InputTokens > 0 || usage.OutputTokens > 0 {
		if err := u.repo.RecordUsage(ctx, &domain.AIUsage{
			UserID:         req.UserID,
			ConversationID: conv.ID,
			Model:          u.cfg.Model,
			InputTokens:    usage.InputTokens,
			OutputTokens:   usage.OutputTokens,
		}); err != nil {
			logger.ErrorWithContext(ctx, "failed to record ai usage", logger.Fields{
				"usecase":         "ai",
				"method":          "Chat",
				"conversation_id": conv.ID,
				"error":           err.Error(),
			})
		}
	}

	if streamErr != nil {
		// Context цуцлалт (клиент салсан) ба provider алдааг ялгахгүйгээр
		// дотоод алдаа болгон боно — handler энэ үед SSE error event бичнэ.
		if ctx.Err() != nil {
			return ChatResponse{}, apperror.InternalCause(ctx.Err())
		}
		return ChatResponse{}, apperror.InternalCause(streamErr)
	}

	return ChatResponse{
		Conversation: conv,
		UserMessage:  userMsg,
		Reply:        reply,
		InputTokens:  usage.InputTokens,
		OutputTokens: usage.OutputTokens,
	}, nil
}

// makeTitle нь эхний мессежээс ярианы гарчиг үүсгэнэ (rune-аар тайрна —
// кирилл текст байтаар тайрахад эвдэрдэг).
func makeTitle(message string) string {
	title := strings.Join(strings.Fields(message), " ")
	if utf8.RuneCountInString(title) <= titleMaxRunes {
		return title
	}
	runes := []rune(title)
	return string(runes[:titleMaxRunes]) + "…"
}
