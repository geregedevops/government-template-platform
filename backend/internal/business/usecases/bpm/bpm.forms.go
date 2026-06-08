// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package bpm

import (
	"context"
	"strings"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/domain"
)

// CreateForm нь хуваалцсан форм сан руу шинэ форм нэмнэ.
func (u *usecase) CreateForm(ctx context.Context, req SaveFormRequest) (FormResponse, error) {
	if strings.TrimSpace(req.Name) == "" {
		return FormResponse{}, apperror.BadRequest("form name is required")
	}
	schema := strings.TrimSpace(req.Schema)
	if schema == "" {
		schema = "{}"
	}
	f, err := u.repo.CreateForm(ctx, &domain.BPMForm{UserID: req.UserID, Name: req.Name, Schema: schema})
	if err != nil {
		return FormResponse{}, mapRepoError(err, "create form")
	}
	return FormResponse{Form: f}, nil
}

// UpdateForm нь форм засна (latest-wins: бүх процесс шинэ schema-г шууд авна).
func (u *usecase) UpdateForm(ctx context.Context, req SaveFormRequest) (FormResponse, error) {
	if strings.TrimSpace(req.ID) == "" {
		return FormResponse{}, apperror.BadRequest("form id is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return FormResponse{}, apperror.BadRequest("form name is required")
	}
	schema := strings.TrimSpace(req.Schema)
	if schema == "" {
		schema = "{}"
	}
	f, err := u.repo.UpdateForm(ctx, &domain.BPMForm{ID: req.ID, Name: req.Name, Schema: schema})
	if err != nil {
		return FormResponse{}, mapRepoError(err, "update form")
	}
	return FormResponse{Form: f}, nil
}

// ListForms нь хэрэглэгчийн хуваалцсан формуудыг буцаана (modeler-ийн сонголтод).
func (u *usecase) ListForms(ctx context.Context, req ListFormsRequest) (ListFormsResponse, error) {
	items, err := u.repo.ListForms(ctx, req.UserID, req.Offset, req.Limit)
	if err != nil {
		return ListFormsResponse{}, mapRepoError(err, "list forms")
	}
	return ListFormsResponse{Forms: items}, nil
}

// DeleteForm нь форм устгана. (Лавласан процессууд runtime-д хоосон формтой
// үлдэж болзошгүйг дуудагч анхаарна — latest-wins.)
func (u *usecase) DeleteForm(ctx context.Context, req GetFormRequest) error {
	if err := u.repo.DeleteForm(ctx, req.ID); err != nil {
		return mapRepoError(err, "delete form")
	}
	return nil
}
