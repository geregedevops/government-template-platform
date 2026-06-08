// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package bpm

import (
	"fmt"
	"time"
)

// Redis key-ийн угтвар — ai.redis_keys.go-той ижил төвлөрсөн загвар.
const prefixGenerateDailyCount = "bpm_generate_daily_count:"

// GenerateDailyCountKey нь нэг хэрэглэгчийн өнөөдрийн AI-аар процесс үүсгэх
// хүсэлтийн тоолуурын түлхүүр. Өдөр UTC-аар эргэнэ (түлхүүр өдрөөр нэрлэгддэг).
func GenerateDailyCountKey(userID string, now time.Time) string {
	return fmt.Sprintf("%s%s:%s", prefixGenerateDailyCount, userID, now.UTC().Format("2006-01-02"))
}
