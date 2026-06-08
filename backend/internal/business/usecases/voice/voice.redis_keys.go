// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package voice

import (
	"fmt"
	"time"
)

// Redis key-ийн угтвар — ai.redis_keys.go-тэй ижил төвлөрсөн загвар.
const prefixDailyCount = "voice_daily_count:"

// DailyCountKey нь нэг хэрэглэгчийн өнөөдрийн дуу хоолойн орчуулгын
// тоолуурын түлхүүр юм. Өдөр нь UTC-аар эргэдэг.
func DailyCountKey(userID string, now time.Time) string {
	return fmt.Sprintf("%s%s:%s", prefixDailyCount, userID, now.UTC().Format("2006-01-02"))
}
