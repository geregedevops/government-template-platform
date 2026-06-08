// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package organization нь байгууллагын модны use case давхарга.
package organization

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/domain"
	repointerface "geregetemplateai/internal/datasources/repositories/interface"
)

// Usecase нь байгууллагын мод (CRUD).
type Usecase interface {
	List(ctx context.Context) (ListResponse, error)
	Create(ctx context.Context, req SaveRequest) (Response, error)
	Update(ctx context.Context, req SaveRequest) (Response, error)
	Delete(ctx context.Context, id string) error
}

type (
	SaveRequest struct {
		ID       string
		ParentID string
		Name     string
		Kind     string
	}
	Response struct {
		Org domain.Organization
	}
	ListResponse struct {
		Orgs []domain.Organization
	}
)

type usecase struct {
	repo repointerface.OrganizationRepository
}

func NewUsecase(repo repointerface.OrganizationRepository) Usecase {
	return &usecase{repo: repo}
}

var validKinds = map[string]bool{
	domain.OrgKindRoot: true, domain.OrgKindMinistry: true,
	domain.OrgKindAgency: true, domain.OrgKindSOE: true,
}

func (u *usecase) List(ctx context.Context) (ListResponse, error) {
	orgs, err := u.repo.List(ctx)
	if err != nil {
		return ListResponse{}, mapRepoError(err, "list organizations")
	}
	return ListResponse{Orgs: orgs}, nil
}

func (u *usecase) Create(ctx context.Context, req SaveRequest) (Response, error) {
	if strings.TrimSpace(req.Name) == "" {
		return Response{}, apperror.BadRequest("organization name is required")
	}
	if strings.TrimSpace(req.ParentID) == "" {
		return Response{}, apperror.BadRequest("parent organization is required")
	}
	kind := req.Kind
	if kind == "" {
		kind = domain.OrgKindAgency
	}
	if kind == domain.OrgKindRoot || !validKinds[kind] {
		// root-ийг зөвхөн seed үүсгэнэ; шинээр root нэмэхгүй.
		return Response{}, apperror.BadRequest("invalid organization kind")
	}
	org, err := u.repo.Create(ctx, &domain.Organization{
		ParentID: req.ParentID, Name: req.Name, Kind: kind,
	})
	if err != nil {
		return Response{}, mapRepoError(err, "create organization")
	}
	return Response{Org: org}, nil
}

func (u *usecase) Update(ctx context.Context, req SaveRequest) (Response, error) {
	if strings.TrimSpace(req.ID) == "" {
		return Response{}, apperror.BadRequest("organization id is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return Response{}, apperror.BadRequest("organization name is required")
	}
	kind := req.Kind
	if kind != "" && !validKinds[kind] {
		return Response{}, apperror.BadRequest("invalid organization kind")
	}
	if kind == "" {
		kind = domain.OrgKindAgency
	}
	org, err := u.repo.Update(ctx, req.ID, req.Name, kind)
	if err != nil {
		return Response{}, mapRepoError(err, "update organization")
	}
	return Response{Org: org}, nil
}

func (u *usecase) Delete(ctx context.Context, id string) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		return mapRepoError(err, "delete organization")
	}
	return nil
}

func mapRepoError(err error, op string) error {
	if err == nil {
		return nil
	}
	var de *apperror.DomainError
	if errors.As(err, &de) {
		return err
	}
	return apperror.InternalCause(fmt.Errorf("%s: %w", op, err))
}
