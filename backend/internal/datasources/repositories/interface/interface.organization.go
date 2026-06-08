// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package _interface

import (
	"context"

	"geregetemplateai/internal/business/domain"
)

// OrganizationRepository нь байгууллагын модны gateway (ltree материалжсан зам).
type OrganizationRepository interface {
	// Create нь parent доор шинэ байгууллага үүсгэж, path-ийг тооцоолно
	// (parent.path + label). ParentID хоосон бол root түвшний болно.
	Create(ctx context.Context, in *domain.Organization) (domain.Organization, error)
	Get(ctx context.Context, id string) (domain.Organization, error)
	// List нь харагдах бүх байгууллагыг path-аар эрэмбэлж буцаана (мод барих).
	List(ctx context.Context) ([]domain.Organization, error)
	// Update нь зөвхөн name/kind-ийг шинэчилнэ (path/hierarchy хөдөлгөхгүй).
	Update(ctx context.Context, id, name, kind string) (domain.Organization, error)
	Delete(ctx context.Context, id string) error
}
