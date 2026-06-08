// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package i18n нь API-ийн хэрэглэгчид харагдах мессежүүдийг (амжилт,
// алдаа) Accept-Language толгойн дагуу Монгол / Англи хэлээр буцаах
// хөрвүүлгийн хамгийн доод түвшний (leaf) package юм. rls package-тай
// адил зөвхөн стандарт сангаас хамаардаг тул HTTP, business, datasource
// давхаргууд import cycle үүсгэхгүйгээр хуваалцаж чадна.
//
// Загвар: код доторх каноник мессежүүд англиар бичигдсэн хэвээр үлдэнэ
// (одоо байгаа клиент/тестүүдийг эвдэхгүй); хэрэглэгч Accept-Language-аар
// mn хүссэн үед хариу бичих мөчид (handler.base_response.go) каталогоос
// орчуулга хайж буцаана. Орчуулга олдохгүй мессеж англиараа дамжина —
// fail-open, мэдээлэл алдагдахгүй.
package i18n

import (
	"context"
	"strings"
)

// Lang нь дэмжигдсэн хэлний код юм.
type Lang string

const (
	LangEN Lang = "en"
	LangMN Lang = "mn"
)

// DefaultLang нь Accept-Language байхгүй / танигдахгүй үеийн өгөгдмөл.
// Англи хэл нь одоогийн API-ийн каноник хэл тул өгөгдмөлөөр үлдээв —
// ингэснээр толгой илгээдэггүй одоогийн клиентүүдийн зан төлөв өөрчлөгдөхгүй.
const DefaultLang = LangEN

type ctxKey struct{}

// With нь хэлийг context-д суулгана — usecase давхарга (жишээ нь AI
// system prompt) хэрэглэгчийн хэлийг мэдэх шаардлагатай үед ашиглана.
func With(ctx context.Context, lang Lang) context.Context {
	return context.WithValue(ctx, ctxKey{}, lang)
}

// FromContext нь суулгасан хэлийг буцаана; байхгүй бол DefaultLang.
func FromContext(ctx context.Context) Lang {
	if lang, ok := ctx.Value(ctxKey{}).(Lang); ok {
		return lang
	}
	return DefaultLang
}

// ParseAcceptLanguage нь Accept-Language толгойг хялбаршуулсан байдлаар
// уншиж дэмжигдсэн хэл рүү буулгана. Бүрэн RFC 9110 q-factor sort хийхгүй —
// зөвхөн хоёр хэл дэмждэг тул жагсаалтад эхэлж таарсан дэмжигдсэн хэлийг
// сонгох нь хангалттай ("mn-MN,mn;q=0.9,en;q=0.8" → mn).
func ParseAcceptLanguage(header string) Lang {
	if header == "" {
		return DefaultLang
	}
	for _, part := range strings.Split(header, ",") {
		tag := strings.TrimSpace(part)
		// ";q=…" жинг хас — дараалал нь хүслийн эрэмбийг аль хэдийн илэрхийлдэг.
		if i := strings.IndexByte(tag, ';'); i >= 0 {
			tag = tag[:i]
		}
		tag = strings.ToLower(strings.TrimSpace(tag))
		switch {
		case tag == "mn" || strings.HasPrefix(tag, "mn-"):
			return LangMN
		case tag == "en" || strings.HasPrefix(tag, "en-"):
			return LangEN
		}
	}
	return DefaultLang
}

// T нь каноник (англи) мессежийг хүссэн хэл рүү хөрвүүлнэ. Орчуулга
// байхгүй эсвэл хэл нь англи бол мессеж өөрчлөгдөхгүй буцна.
func T(lang Lang, msg string) string {
	if lang != LangMN {
		return msg
	}
	if mn, ok := catalogMN[msg]; ok {
		return mn
	}
	return msg
}

// catalogMN нь каноник англи мессеж → монгол орчуулгын каталог.
// Шинэ хэрэглэгчид-харагдах мессеж нэмэхдээ энд орчуулгыг нь хамт нэм;
// мартсан тохиолдолд мессеж англиараа гарна (аюулгүй fallback).
var catalogMN = map[string]string{
	// --- нийтлэг ---
	"internal server error": "Серверийн дотоод алдаа гарлаа",
	"validation failed":     "Талбарын шалгалт амжилтгүй боллоо",
	"invalid request body":  "Хүсэлтийн бие буруу байна",
	"too many requests":     "Хэт олон хүсэлт илгээлээ, түр хүлээгээд дахин оролдоно уу",

	// --- auth middleware ---
	"missing authorization header":          "Authorization толгой дутуу байна",
	"invalid header format":                 "Authorization толгойн формат буруу байна",
	"token must content bearer":             "Bearer токен шаардлагатай",
	"invalid token":                         "Токен хүчингүй байна",
	"token has been revoked":                "Токен хүчингүй болгогдсон байна",
	"you don't have access for this action": "Танд энэ үйлдлийг хийх эрх байхгүй",

	// --- auth / users usecase алдаанууд ---
	"account already activated":                                "Бүртгэл аль хэдийн идэвхжсэн байна",
	"invalid otp code":                                         "Баталгаажуулах код буруу байна",
	"new password is required":                                 "Шинэ нууц үг шаардлагатай",
	"otp code expired or not found":                            "Баталгаажуулах кодын хугацаа дууссан эсвэл олдсонгүй",
	"otp code has been sent":                                   "Баталгаажуулах код илгээгдлээ",
	"reset token is required":                                  "Сэргээх токен шаардлагатай",
	"user id required":                                         "Хэрэглэгчийн ID шаардлагатай",
	"username or email already exists":                         "Хэрэглэгчийн нэр эсвэл имэйл аль хэдийн бүртгэлтэй байна",
	"account is not activated":                                 "Бүртгэл идэвхжээгүй байна",
	"too many failed login attempts, please try again later":   "Нэвтрэх оролдлого хэт олон удаа амжилтгүй боллоо, түр хүлээгээд дахин оролдоно уу",
	"too many invalid otp attempts, please request a new code": "Буруу код хэт олон удаа орууллаа, шинэ код авна уу",
	"too many password reset requests, please try again later": "Нууц үг сэргээх хүсэлт хэт олон байна, түр хүлээгээд дахин оролдоно уу",
	"email not found":                                          "Имэйл олдсонгүй",
	"user not found":                                           "Хэрэглэгч олдсонгүй",
	"users fetched successfully":                               "Хэрэглэгчдийг амжилттай татлаа",
	"user created successfully":                                "Хэрэглэгч амжилттай үүслээ",
	"role updated successfully":                                "Эрх амжилттай шинэчлэгдлээ",
	"user deleted successfully":                                "Хэрэглэгч амжилттай устгагдлаа",
	"cannot delete yourself":                                   "Та өөрийгөө устгах боломжгүй",
	"invalid role":                                             "Буруу эрх",
	"roles fetched successfully":                               "Эрхүүдийг амжилттай татлаа",
	"permissions fetched successfully":                         "Зөвшөөрлүүдийг амжилттай татлаа",
	"role created successfully":                                "Эрх амжилттай үүслээ",
	"role deleted successfully":                                "Эрх амжилттай устгагдлаа",
	"role permissions updated successfully":                    "Эрхийн зөвшөөрөл шинэчлэгдлээ",
	"role not found":                                           "Эрх олдсонгүй",
	"role not found or is a system role":                       "Эрх олдсонгүй эсвэл системийн эрх",
	"role key already exists":                                  "Энэ түлхүүртэй эрх аль хэдийн байна",
	"role is assigned to users":                                "Энэ эрх хэрэглэгчдэд оноогдсон тул устгах боломжгүй",
	"role key is required":                                     "Эрхийн түлхүүр шаардлагатай",
	"role name is required":                                    "Эрхийн нэр шаардлагатай",
	"current password is incorrect":                            "Одоогийн нууц үг буруу байна",
	"invalid email or password":                                "Имэйл эсвэл нууц үг буруу байна",
	"invalid refresh token":                                    "Refresh токен хүчингүй байна",
	"refresh token has been revoked":                           "Refresh токен хүчингүй болгогдсон байна",
	"reset token is invalid or expired":                        "Сэргээх токен хүчингүй эсвэл хугацаа нь дууссан байна",
	"user no longer exists":                                    "Хэрэглэгч байхгүй болсон байна",

	// --- амжилтын мессежүүд ---
	"registration user success":                              "Бүртгэл амжилттай үүслээ",
	"if the email is registered, a reset link has been sent": "Хэрэв энэ имэйл бүртгэлтэй бол нууц үг сэргээх холбоос илгээгдлээ",
	"login success":                  "Амжилттай нэвтэрлээ",
	"logout success":                 "Амжилттай гарлаа",
	"otp verification success":       "Код амжилттай баталгаажлаа",
	"password changed":               "Нууц үг солигдлоо",
	"password reset":                 "Нууц үг сэргээгдлээ",
	"token refreshed":                "Токен шинэчлэгдлээ",
	"user data fetched successfully": "Хэрэглэгчийн мэдээлэл амжилттай татагдлаа",

	// --- AI ---
	"ai service is not configured":       "AI үйлчилгээ тохируулагдаагүй байна",
	"ai daily request limit exceeded":    "Өнөөдрийн AI хүсэлтийн хязгаарт хүрлээ, маргааш дахин оролдоно уу",
	"conversation not found":             "Харилцан яриа олдсонгүй",
	"message is required":                "Мессеж шаардлагатай",
	"chat stream started":                "Чат урсгал эхэллээ",
	"conversations fetched successfully": "Харилцан яриануудыг амжилттай татлаа",
	"messages fetched successfully":      "Мессежүүдийг амжилттай татлаа",
	"knowledge not found":                "Мэдлэгийн бичлэг олдсонгүй",
	"knowledge fetched successfully":     "Мэдлэгийг амжилттай татлаа",
	"knowledge created successfully":     "Мэдлэг амжилттай нэмэгдлээ",
	"knowledge updated successfully":     "Мэдлэг амжилттай шинэчлэгдлээ",
	"knowledge deleted successfully":     "Мэдлэг амжилттай устгагдлаа",

	// --- Voice (дуу хоолойн орчуулга) ---
	"voice service is not configured":    "Дуу хоолойн үйлчилгээ тохируулагдаагүй байна",
	"voice daily request limit exceeded": "Өнөөдрийн орчуулгын хязгаарт хүрлээ, маргааш дахин оролдоно уу",
	"source language must be mn or en":   "Эх хэл нь mn эсвэл en байх ёстой",
	"audio is required":                  "Аудио шаардлагатай",
	"audio is too large":                 "Аудио хэт том байна",
	"audio is not valid base64":          "Аудио буруу base64 хэлбэртэй байна",
	"voice translation failed":           "Дуу хоолойн орчуулга амжилтгүй боллоо",
	"voice synthesis failed":             "Дуу хоолой үүсгэх амжилтгүй боллоо",
	"voice translated successfully":      "Орчуулга амжилттай боллоо",
	"voice history fetched successfully": "Орчуулгын түүхийг амжилттай татлаа",
	"language must be mn or en":          "Хэл нь mn эсвэл en байх ёстой",
	"text is required":                   "Бичвэр шаардлагатай",
	"voice transcription failed":         "Дуу бичвэрлэх амжилтгүй боллоо",
	"voice transcribed successfully":     "Дуу амжилттай бичвэрлэгдлээ",
	"voice synthesized successfully":     "Дуу хоолой амжилттай үүслээ",

	// --- Федераци (ROADMAP Үе 1) ---
	"peer registered successfully":       "Гишүүн node амжилттай бүртгэгдлээ",
	"peer updated successfully":          "Гишүүн node амжилттай шинэчлэгдлээ",
	"peer deleted successfully":          "Гишүүн node амжилттай устгагдлаа",
	"peer not found":                     "Гишүүн node олдсонгүй",
	"peer key already exists":            "Энэ node-ийн key аль хэдийн бүртгэлтэй байна",
	"peer key is required":               "node-ийн key шаардлагатай",
	"base_url and jwks_url are required": "base_url ба jwks_url шаардлагатай",
	"peer is not active":                 "Гишүүн node идэвхгүй байна",
	"federation is not configured":       "Федераци тохируулагдаагүй байна",
	"unknown peer":                       "Танигдахгүй node",
	"cannot resolve peer key":            "node-ийн түлхүүрийг тогтоож чадсангүй",
	"signature verification failed":      "Гарын үсгийн баталгаажуулалт амжилтгүй",
	"invalid envelope":                   "Мессежийн дугтуй буруу байна",
	"invalid peer id":                    "node-ийн id буруу байна",

	// --- BPM (Business Process Management) ---
	"process not found":                         "Процесс олдсонгүй",
	"instance not found":                        "Гүйлт олдсонгүй",
	"task not found":                            "Даалгавар олдсонгүй",
	"form not found":                            "Маягт олдсонгүй",
	"form name is required":                     "Маягтын нэр шаардлагатай",
	"form id is required":                       "Маягтын ID шаардлагатай",
	"organization not found":                    "Байгууллага олдсонгүй",
	"parent organization not found":             "Эцэг байгууллага олдсонгүй",
	"organization name is required":             "Байгууллагын нэр шаардлагатай",
	"parent organization is required":           "Эцэг байгууллага шаардлагатай",
	"invalid organization kind":                 "Байгууллагын төрөл буруу",
	"cannot delete the root organization":       "Үндэс байгууллагыг устгах боломжгүй",
	"task already completed":                    "Даалгавар аль хэдийн дууссан байна",
	"invalid process definition":                "Процессын тодорхойлолт буруу байна",
	"invalid submission data":                   "Бөглөсөн өгөгдөл буруу байна",
	"process has no bpmn diagram":               "Процесст BPMN диаграмм алга байна",
	"process has too many nodes":                "Процесст хэт олон node байна",
	"invalid bpmn xml":                          "BPMN XML буруу байна",
	"process must have exactly one start event": "Процесст яг нэг эхлэл (start) event байх ёстой",
	"process must have at least one end event":  "Процесст дор хаяж нэг төгсгөл (end) event байх ёстой",
	"service task failed":                       "Гадаад үйлчилгээ (service task) дуудлага амжилтгүй боллоо",
	"service task is not configured":            "Service task тохируулагдаагүй байна (URL дутуу)",
	"no gateway branch matched":                 "Шийдвэрийн (gateway) нөхцөлд тохирох салаа олдсонгүй",
	"process generated successfully":            "Процесс амжилттай үүсгэгдлээ",
	"ai generation failed":                      "AI процесс үүсгэх амжилтгүй боллоо",
	"ai returned invalid spec":                  "AI буруу бүтэцтэй хариу буцаалаа",
	"ai returned an empty process":              "AI хоосон процесс буцаалаа",
	"generated process is invalid":              "Үүсгэсэн процесс хүчингүй байна",
	"description is required":                   "Тайлбар шаардлагатай",
	"process created successfully":              "Процесс амжилттай үүслээ",
	"process updated successfully":              "Процесс амжилттай шинэчлэгдлээ",
	"process fetched successfully":              "Процессыг амжилттай татлаа",
	"processes fetched successfully":            "Процессуудыг амжилттай татлаа",
	"process deleted successfully":              "Процесс амжилттай устгагдлаа",
	"instance started successfully":             "Гүйлт амжилттай эхэллээ",
	"task fetched successfully":                 "Даалгаврыг амжилттай татлаа",
	"instances fetched successfully":            "Гүйлтүүдийг амжилттай татлаа",
	"events fetched successfully":               "Бүртгэлийг амжилттай татлаа",
	"task submitted successfully":               "Даалгавар амжилттай илгээгдлээ",
}
