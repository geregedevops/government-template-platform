// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package ai

import (
	"context"

	"geregetemplateai/internal/apperror"
)

func (u *usecase) ListConversations(ctx context.Context, req ListConversationsRequest) (ListConversationsResponse, error) {
	convs, err := u.repo.ListConversations(ctx, req.UserID, req.Offset, req.Limit)
	if err != nil {
		return ListConversationsResponse{}, mapRepoError(err, "list conversations")
	}
	return ListConversationsResponse{Conversations: convs}, nil
}

func (u *usecase) GetMessages(ctx context.Context, req GetMessagesRequest) (GetMessagesResponse, error) {
	conv, err := u.repo.GetConversation(ctx, req.ConversationID)
	if err != nil {
		return GetMessagesResponse{}, mapRepoError(err, "get conversation")
	}
	if conv.UserID != req.UserID {
		// Бусдын яриаг "байхгүй" гэж харуулна — enumeration хамгаалалт.
		return GetMessagesResponse{}, apperror.NotFound("conversation not found")
	}
	msgs, err := u.repo.ListMessages(ctx, conv.ID, 0)
	if err != nil {
		return GetMessagesResponse{}, mapRepoError(err, "list messages")
	}
	return GetMessagesResponse{Conversation: conv, Messages: msgs}, nil
}
