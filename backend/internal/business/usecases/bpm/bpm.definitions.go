// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package bpm

import (
	"context"

	"geregetemplateai/internal/business/domain"
)

func (u *usecase) CreateProcess(ctx context.Context, req SaveProcessRequest) (ProcessResponse, error) {
	if err := u.validateDefinition(req.Definition); err != nil {
		return ProcessResponse{}, err
	}
	status := normalizeStatus(req.Status)

	process, err := u.repo.CreateDefinition(ctx, &domain.BPMProcessDefinition{
		UserID:      req.UserID,
		OrgID:       req.OrgID,
		Name:        req.Name,
		Description: req.Description,
		Definition:  req.Definition,
		Status:      status,
	})
	if err != nil {
		return ProcessResponse{}, mapRepoError(err, "create process")
	}
	return ProcessResponse{Process: process}, nil
}

func (u *usecase) UpdateProcess(ctx context.Context, req UpdateProcessRequest) (ProcessResponse, error) {
	if err := u.validateDefinition(req.Definition); err != nil {
		return ProcessResponse{}, err
	}
	// Эзэмшил/оршихуйг RLS + explicit Get-ээр шалгана (өөр хэрэглэгчийн
	// процессыг засах оролдлого NotFound буцна).
	if _, err := u.repo.GetDefinition(ctx, req.ID); err != nil {
		return ProcessResponse{}, mapRepoError(err, "get process")
	}

	process, err := u.repo.UpdateDefinition(ctx, &domain.BPMProcessDefinition{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Definition:  req.Definition,
		Status:      normalizeStatus(req.Status),
	})
	if err != nil {
		return ProcessResponse{}, mapRepoError(err, "update process")
	}
	return ProcessResponse{Process: process}, nil
}

func (u *usecase) GetProcess(ctx context.Context, req GetProcessRequest) (ProcessResponse, error) {
	process, err := u.repo.GetDefinition(ctx, req.ID)
	if err != nil {
		return ProcessResponse{}, mapRepoError(err, "get process")
	}
	return ProcessResponse{Process: process}, nil
}

func (u *usecase) ListProcesses(ctx context.Context, req ListProcessesRequest) (ListProcessesResponse, error) {
	processes, err := u.repo.ListDefinitions(ctx, req.UserID, req.Offset, req.Limit)
	if err != nil {
		return ListProcessesResponse{}, mapRepoError(err, "list processes")
	}
	return ListProcessesResponse{Processes: processes}, nil
}

func (u *usecase) DeleteProcess(ctx context.Context, req GetProcessRequest) error {
	if err := u.repo.DeleteDefinition(ctx, req.ID); err != nil {
		return mapRepoError(err, "delete process")
	}
	return nil
}

// normalizeStatus нь хоосон/буруу төлвийг "draft" болгоно.
func normalizeStatus(status string) string {
	if status == domain.BPMStatusPublished {
		return domain.BPMStatusPublished
	}
	return domain.BPMStatusDraft
}
