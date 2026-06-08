// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package bpm

import (
	"context"
	"encoding/json"
	"fmt"

	"geregetemplateai/internal/apperror"
	"geregetemplateai/internal/business/ports"
	repointerface "geregetemplateai/internal/datasources/repositories/interface"
)

// Config нь BPM usecase-ийн тохируулга. Утгууд нь server.go-ийн угсралтаар
// дамжин ирдэг — usecase нь viper/env-ийн талаар мэддэггүй.
type Config struct {
	// MaxNodes нь нэг процесст зөвшөөрөх node-ийн дээд тоо (abuse хамгаалалт,
	// 0/сөрөг бол өгөгдмөл хэрэглэнэ).
	MaxNodes int
	// AIEnabled нь Claude (генератор) тохируулагдсан эсэх. false үед
	// GenerateProcess нь apperror.Unavailable (503) буцаана.
	AIEnabled bool
	// GenerateDailyLimit нь нэг хэрэглэгчийн өдөрт AI-аар процесс үүсгэх дээд
	// тоо (Redis тоолуур). 0/сөрөг бол хязгааргүй.
	GenerateDailyLimit int
}

const defaultMaxNodes = 200

// Connector нь serviceTask-аас гадаад REST API руу хийх HTTP дуудлагын гарах
// хил (outbound boundary). bpmconnector.Client (production, SSRF-хамгаалалттай)
// эсвэл mock (test) хэрэгжүүлнэ.
type Connector interface {
	Do(ctx context.Context, method, url string, headers map[string]string, body string) (int, []byte, error)
}

// Generator нь текст тайлбараас процессын JSON spec гаргах AI-ийн гарах хил.
// server.go нь Claude (aiclient)-ийн адаптерийг дамжуулна; тестэд mock.
type Generator interface {
	Generate(ctx context.Context, system, userMessage string) (string, error)
}

// FedSender нь delegatedTask дээр өөр node руу гарын үсэгтэй мессеж илгээх
// гарах хил (федерацийн usecase хэрэгжүүлнэ). nil бол федераци идэвхгүй.
type FedSender interface {
	SendByKey(ctx context.Context, peerKey, typ string, body json.RawMessage) error
}

type usecase struct {
	repo      repointerface.BPMRepository
	connector Connector
	generator Generator
	cache     ports.Cache
	cfg       Config
	fedSender FedSender
}

// SetFedSender нь федерацийн илгээгчийг (delegatedTask-д) тарина. server.go
// нь bpm↔federation мөчлөгийг таслахын тулд аль алийг үүсгэсний дараа дуудна.
func (u *usecase) SetFedSender(s FedSender) { u.fedSender = s }

// NewUsecase нь BPM usecase-ийг үүсгэнэ. cache нь AI-аар үүсгэх өдрийн лимитэд
// (Redis) ашиглагдана; nil бол лимит шалгахгүй.
func NewUsecase(repo repointerface.BPMRepository, connector Connector, generator Generator, cache ports.Cache, cfg Config) Usecase {
	if cfg.MaxNodes <= 0 {
		cfg.MaxNodes = defaultMaxNodes
	}
	return &usecase{repo: repo, connector: connector, generator: generator, cache: cache, cfg: cfg}
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
