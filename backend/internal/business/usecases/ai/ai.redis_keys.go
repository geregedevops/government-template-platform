// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package ai

import (
	"fmt"
	"time"
)

// Redis key-ийн угтварууд — auth.redis_keys.go-тэй ижил төвлөрсөн загвар.
const prefixDailyCount = "ai_daily_count:"

// DailyCountKey нь нэг хэрэглэгчийн өнөөдрийн AI хүсэлтийн тоолуурын
// түлхүүр юм. Өдөр нь UTC-аар эргэдэг — энгийн, тогтвортой; хэрэглэгчийн
// цагийн бүсээр эргүүлэх нь шаардлагагүй нарийвчлал.
func DailyCountKey(userID string, now time.Time) string {
	return fmt.Sprintf("%s%s:%s", prefixDailyCount, userID, now.UTC().Format("2006-01-02"))
}
