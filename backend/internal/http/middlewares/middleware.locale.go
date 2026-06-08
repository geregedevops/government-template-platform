// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package middlewares

import (
	V1Handler "geregetemplateai/internal/http/handlers/v1"
	"geregetemplateai/internal/i18n"
	"github.com/gofiber/fiber/v3"
)

// LocaleMiddleware нь Accept-Language толгойг уншиж, шийдвэрлэсэн хэлийг
// хоёр газар суулгана:
//
//  1. Fiber Locals (V1Handler.LocaleLocalsKey) — хариу бичигчид
//     (handler.base_response.go) мессежийг орчуулахад ашиглана;
//  2. request context (i18n.With) — usecase давхарга хэрэглэгчийн хэлийг
//     мэдэх шаардлагатай үед (жишээ нь AI assistant-ийн system prompt-д
//     "хэрэглэгчийн хэлээр хариул" гэж дамжуулах) ашиглана.
//
// Толгой байхгүй / танигдахгүй үед i18n.DefaultLang (en) — одоогийн
// клиентүүдийн зан төлөв өөрчлөгдөхгүй.
func LocaleMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		lang := i18n.ParseAcceptLanguage(c.Get(fiber.HeaderAcceptLanguage))
		c.Locals(V1Handler.LocaleLocalsKey, lang)
		c.SetContext(i18n.With(c.Context(), lang))
		return c.Next()
	}
}
