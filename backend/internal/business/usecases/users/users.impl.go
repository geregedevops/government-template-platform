// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package users

import (
	"context"
	"fmt"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/datasources/caches"
	repointerface "geregetemplateai/internal/datasources/repositories/interface"
	"golang.org/x/sync/singleflight"
)

// TokenRevoker нь хэрэглэгчийн идэвхтэй access токенуудыг хүчингүй болгох port.
// auth usecase хангадаг; эрх (role) солигдоход дуудаж, шинэ эрхийг нэн даруй
// хүчинтэй болгоно (хэрэглэгчийг refresh хийлгэнэ). Construction cycle-ээс
// зайлсхийхийн тулд SetTokenRevoker-оор дараа нь холбоно.
type TokenRevoker interface {
	RevokeUserTokens(ctx context.Context, userID string) error
}

// Config нь usecase-ийн domain давхарга руу дамжуулдаг тохируулах боломжтой
// утгуудыг агуулна. Domain өөрөө bcryptCost-ийг параметрээр авдаг тул тохиргооны
// асуудлуудаас ангид хэвээр үлдэж чадна; usecase нь config-ийн талаар мэддэг
// хил юм.
type Config struct {
	BcryptCost int
}

// usecase нь хамаарлууд болон method хоорондын төлөвийг агуулдаг. Нэг зан
// төлөв өөрчлөгдөхөд PR-ийн diff нарийн (surgical) хэвээр үлдэхийн тулд method
// бүр өөрийн файлд байрладаг.
type usecase struct {
	repo           repointerface.UserRepository
	ristrettoCache caches.RistrettoCache
	cfg            Config
	revoker        TokenRevoker // optional; SetTokenRevoker-оор холбоно

	// userByEmailGroup нь ижил email-ийн зэрэгцээ кэш алдалтуудыг (cache miss)
	// нэгтгэдэг тул олон зэрэг хүсэлт (thundering herd) N зэрэгцээ DB дуудлага
	// болон тархахгүй. Group нь нормчилсон email-ээр түлхүүрлэгдэнэ.
	userByEmailGroup singleflight.Group
}

// NewUsecase нь User CRUD use case-ийг үүсгэнэ. Энэ нь auth-тай холбоотой ямар
// нэг хамтрагчаас (JWT, Redis, mailer байхгүй) хамаардаггүй — энэ нь User vs
// Auth хуваагдлын гол утга юм.
func NewUsecase(repo repointerface.UserRepository, ristrettoCache caches.RistrettoCache, cfg Config) Usecase {
	return &usecase{
		repo:           repo,
		ristrettoCache: ristrettoCache,
		cfg:            cfg,
	}
}

// SetTokenRevoker нь auth usecase-ийг (TokenRevoker) дараа холбоно — users нь
// auth-аас өмнө үүсдэг тул constructor cycle-ээс зайлсхийнэ.
func (u *usecase) SetTokenRevoker(r TokenRevoker) { u.revoker = r }

// mapRepoError нь repository-ээс буцсан DomainError төрлүүдийг хадгалж, харин
// түүхий алдаануудыг форматтай дотоод алдаагаар боодог. Үүнгүйгээр дээд урсгал
// дахь errors.As(err, *DomainError) амжилтгүй болно.
func mapRepoError(err error, op string) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(*apperror.DomainError); ok {
		return err
	}
	return apperror.InternalCause(fmt.Errorf("%s: %w", op, err))
}
