// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package voice

import (
	"fmt"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/ports"
	repointerface "geregetemplateai/internal/datasources/repositories/interface"
	"geregetemplateai/pkg/geminiclient"
)

// Config нь voice usecase-ийн тохируулга. Утгууд нь internal/config-оос
// server.go-ийн угсралтаар дамжин ирдэг.
type Config struct {
	// Enabled нь GEMINI_API_KEY тохируулагдсан эсэх. false үед бүх дуудлага
	// apperror.Unavailable (503) буцаана — "чимээгүй буруу тохиргоо
	// үлдээхгүй" зарчим.
	Enabled bool
	// Model нь usage метерингд бичигдэх модель нэр.
	Model string
	// DailyRequestLimit нь нэг хэрэглэгчийн өдөрт хийж болох орчуулгын дээд
	// тоо (Redis тоолуур). 0 буюу сөрөг бол хязгааргүй.
	DailyRequestLimit int
	// MaxAudioBytes нь нэг хүсэлтэд зөвшөөрөх түүхий аудионы дээд хэмжээ.
	// 0 буюу сөрөг бол default (640 KiB). Base64 болоход ~853 KiB болж,
	// JSON-ийн ачаалалтай нийлэхэд глобал 1 MiB body cap дотор багтана.
	MaxAudioBytes int
}

const defaultMaxAudioBytes = 640 * 1024

type usecase struct {
	repo   repointerface.VoiceRepository
	voicer geminiclient.Voicer
	cache  ports.Cache
	cfg    Config
}

// NewUsecase нь voice usecase-ийг үүсгэнэ. voicer нь geminiclient.Client
// (production) эсвэл mock (test) байж болно.
func NewUsecase(repo repointerface.VoiceRepository, voicer geminiclient.Voicer, cache ports.Cache, cfg Config) Usecase {
	if cfg.MaxAudioBytes <= 0 {
		cfg.MaxAudioBytes = defaultMaxAudioBytes
	}
	return &usecase{
		repo:   repo,
		voicer: voicer,
		cache:  cache,
		cfg:    cfg,
	}
}

// mapRepoError — ai usecase-тэй ижил: DomainError-уудыг хадгалж, түүхий
// алдааг ерөнхий дотоод алдаагаар боно.
func mapRepoError(err error, op string) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(*apperror.DomainError); ok {
		return err
	}
	return apperror.InternalCause(fmt.Errorf("%s: %w", op, err))
}
