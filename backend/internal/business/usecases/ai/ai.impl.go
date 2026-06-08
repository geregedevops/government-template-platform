// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package ai

import (
	"fmt"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/ports"
	repointerface "geregetemplateai/internal/datasources/repositories/interface"
	"geregetemplateai/pkg/aiclient"
)

// Config нь AI usecase-ийн тохируулга. Утгууд нь internal/config-оос
// server.go-ийн угсралтаар дамжин ирдэг — usecase нь viper/env-ийн
// талаар мэддэггүй.
type Config struct {
	// Enabled нь ANTHROPIC_API_KEY тохируулагдсан эсэх. false үед бүх
	// дуудлага apperror.Unavailable (503) буцаана — verify клиентийн
	// "чимээгүй буруу тохиргоо үлдээхгүй" зарчим.
	Enabled bool
	// Model нь usage метерингд бичигдэх модель нэр.
	Model string
	// DailyRequestLimit нь нэг хэрэглэгчийн өдөрт хийж болох AI чат
	// хүсэлтийн дээд тоо (Redis тоолуур). 0 буюу сөрөг бол хязгааргүй.
	DailyRequestLimit int
	// HistoryLimit нь Claude руу дамжуулах түүхийн мессежийн дээд тоо.
	HistoryLimit int
}

type usecase struct {
	repo     repointerface.AIRepository
	streamer aiclient.Streamer
	cache    ports.Cache
	cfg      Config
}

// NewUsecase нь AI usecase-ийг үүсгэнэ. streamer нь aiclient.Client
// (production) эсвэл mock (test) байж болно.
func NewUsecase(repo repointerface.AIRepository, streamer aiclient.Streamer, cache ports.Cache, cfg Config) Usecase {
	if cfg.HistoryLimit <= 0 {
		cfg.HistoryLimit = 20
	}
	return &usecase{
		repo:     repo,
		streamer: streamer,
		cache:    cache,
		cfg:      cfg,
	}
}

// mapRepoError — users usecase-тэй ижил: DomainError-уудыг хадгалж,
// түүхий алдааг ерөнхий дотоод алдаагаар боно.
func mapRepoError(err error, op string) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(*apperror.DomainError); ok {
		return err
	}
	return apperror.InternalCause(fmt.Errorf("%s: %w", op, err))
}
