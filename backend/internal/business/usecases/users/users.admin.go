// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/domain"
	repointerface "geregetemplateai/internal/datasources/repositories/interface"
	"geregetemplateai/pkg/logger"
)

// ListUsers нь бүх хэрэглэгчдийг буцаана. Admin token-ы RLS (app.user_role =
// 'admin') нь users_select policy-ээр бүх мөрийг харахыг зөвшөөрнө.
func (uc *usecase) ListUsers(ctx context.Context, req ListUsersRequest) (ListUsersResponse, error) {
	out, err := uc.repo.List(ctx, repointerface.UserListFilter{}, req.Offset, req.Limit)
	if err != nil {
		return ListUsersResponse{}, mapRepoError(err, "list users")
	}
	return ListUsersResponse{Users: out}, nil
}

// AdminCreateUser нь admin-аар идэвхтэй (active) хэрэглэгч үүсгэнэ — OTP
// идэвхжүүлэлтгүйгээр. Store-той ижил build/hash логик (domain.NewUser).
func (uc *usecase) AdminCreateUser(ctx context.Context, req AdminCreateUserRequest) (StoreResponse, error) {
	roleID := req.RoleID
	if roleID <= 0 {
		roleID = domain.RoleUser
	}
	user, buildErr := domain.NewUser(req.Username, req.Email, req.Password, roleID, uc.cfg.BcryptCost)
	if buildErr != nil {
		if errors.Is(buildErr, domain.ErrEmptyUsername) ||
			errors.Is(buildErr, domain.ErrEmptyEmail) ||
			errors.Is(buildErr, domain.ErrInvalidEmail) ||
			errors.Is(buildErr, domain.ErrEmptyPassword) {
			return StoreResponse{}, apperror.BadRequest(buildErr.Error())
		}
		return StoreResponse{}, apperror.InternalCause(fmt.Errorf("build user: %w", buildErr))
	}

	stored, repoErr := uc.repo.Store(ctx, user)
	if repoErr != nil {
		return StoreResponse{}, mapRepoError(repoErr, "create user")
	}

	// Admin-аар үүсгэсэн хэрэглэгчийг шууд идэвхжүүлнэ (OTP-гүй нэвтэрнэ).
	stored.Active = true
	if err := uc.repo.ChangeActiveUser(ctx, &stored); err != nil {
		logger.WarnWithContext(ctx, "admin create: failed to activate new user", logger.Fields{
			"usecase": "users", "method": "AdminCreateUser", "error": err.Error(), "user_id": stored.ID,
		})
	}
	return StoreResponse{User: stored}, nil
}

// UpdateRole нь хэрэглэгчийн эрхийг сольно (динамик role_id; admin UI нь
// бодит эрхүүдээс сонгуулна).
func (uc *usecase) UpdateRole(ctx context.Context, req UpdateRoleRequest) error {
	if req.RoleID <= 0 {
		return apperror.BadRequest("invalid role")
	}
	if err := uc.repo.UpdateRole(ctx, req.ID, req.RoleID); err != nil {
		return mapRepoError(err, "update role")
	}
	// Эрх солигдмогц нэн даруй хүчинтэй болгохын тулд тухайн хэрэглэгчийн
	// одоогийн access токенуудыг хүчингүй болгоно (refresh хийлгэнэ). Алдвал
	// non-fatal — токен дуусахад ямар ч байсан шинэ эрх хүчинтэй болно.
	if uc.revoker != nil {
		if err := uc.revoker.RevokeUserTokens(ctx, req.ID); err != nil {
			logger.WarnWithContext(ctx, "update role: failed to revoke tokens (non-fatal)", logger.Fields{
				"usecase": "users", "method": "UpdateRole", "error": err.Error(), "user_id": req.ID,
			})
		}
	}
	return nil
}

// UpdateOrg нь хэрэглэгчийг өөр байгууллагад шилжүүлнэ (admin удирдлага).
func (uc *usecase) UpdateOrg(ctx context.Context, req UpdateOrgRequest) error {
	if strings.TrimSpace(req.OrgID) == "" {
		return apperror.BadRequest("organization is required")
	}
	if err := uc.repo.UpdateOrg(ctx, req.ID, req.OrgID); err != nil {
		return mapRepoError(err, "update org")
	}
	return nil
}

// DeleteUser нь хэрэглэгчийг зөөлөн устгана.
func (uc *usecase) DeleteUser(ctx context.Context, req DeleteUserRequest) error {
	if err := uc.repo.SoftDelete(ctx, req.ID); err != nil {
		return mapRepoError(err, "delete user")
	}
	return nil
}
