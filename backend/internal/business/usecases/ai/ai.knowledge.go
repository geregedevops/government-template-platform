// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package ai

import (
	"context"

	"geregetemplateai/internal/business/domain"
)

func (u *usecase) CreateKnowledge(ctx context.Context, req SaveKnowledgeRequest) (KnowledgeResponse, error) {
	k, err := u.repo.CreateKnowledge(ctx, &domain.AIKnowledge{
		UserID:  req.UserID,
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		return KnowledgeResponse{}, mapRepoError(err, "create knowledge")
	}
	return KnowledgeResponse{Knowledge: k}, nil
}

func (u *usecase) UpdateKnowledge(ctx context.Context, req UpdateKnowledgeRequest) (KnowledgeResponse, error) {
	// Эзэмшил/оршихуйг RLS + explicit Get-ээр шалгана.
	if _, err := u.repo.GetKnowledge(ctx, req.ID); err != nil {
		return KnowledgeResponse{}, mapRepoError(err, "get knowledge")
	}
	k, err := u.repo.UpdateKnowledge(ctx, &domain.AIKnowledge{
		ID:      req.ID,
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		return KnowledgeResponse{}, mapRepoError(err, "update knowledge")
	}
	return KnowledgeResponse{Knowledge: k}, nil
}

func (u *usecase) ListKnowledge(ctx context.Context, req ListKnowledgeRequest) (ListKnowledgeResponse, error) {
	items, err := u.repo.ListKnowledge(ctx, req.UserID, req.Offset, req.Limit)
	if err != nil {
		return ListKnowledgeResponse{}, mapRepoError(err, "list knowledge")
	}
	return ListKnowledgeResponse{Items: items}, nil
}

// ListAllKnowledge нь бүх хэрэглэгчийн мэдлэгийг буцаана. RLS-ийн ачаар admin
// бүгдийг, энгийн хэрэглэгч зөвхөн өөрийнхөө мэдлэгийг авна.
func (u *usecase) ListAllKnowledge(ctx context.Context, req ListKnowledgeRequest) (ListKnowledgeResponse, error) {
	items, err := u.repo.ListAllKnowledge(ctx, req.Offset, req.Limit)
	if err != nil {
		return ListKnowledgeResponse{}, mapRepoError(err, "list all knowledge")
	}
	return ListKnowledgeResponse{Items: items}, nil
}

func (u *usecase) DeleteKnowledge(ctx context.Context, req DeleteKnowledgeRequest) error {
	if err := u.repo.DeleteKnowledge(ctx, req.ID); err != nil {
		return mapRepoError(err, "delete knowledge")
	}
	return nil
}
