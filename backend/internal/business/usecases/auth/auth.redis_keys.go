// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package auth

import "fmt"

// Auth domain-ийн ашигладаг Redis key-ийн угтварууд (prefix). Бичгийн алдаа
// бичигчийг түүний уншигчаас чимээгүйхэн салгахаас сэргийлэх, мөн энэ багцаас
// гадуурх адаптерууд (ялангуяа auth middleware) format string-ийг дахин
// хэрэгжүүлэхийн оронд яг ижил нэрсийг дахин ашиглахын тулд төвлөрүүлсэн.
const (
	prefixRefresh        = "refresh:"
	prefixUserOTP        = "user_otp:"
	prefixOTPAttempts    = "otp_attempts:"
	prefixLoginAttempts  = "login_attempts:"
	prefixForgotAttempts = "forgot_attempts:"
	prefixPasswordReset  = "pwd_reset:"
	prefixUserResetIndex = "pwd_reset_user:"
	prefixPasswordCutoff = "pwd_cutoff:"
)

// RefreshKey нь refresh токены jti бичлэгүүдийг хүрээлдэг; байхгүй ⇒ хүчингүй болсон.
func RefreshKey(jti string) string {
	return fmt.Sprintf("%s%s", prefixRefresh, jti)
}

// UserOTPKey нь идэвхгүй бүртгэлийн амьд 6 оронтой OTP-г хадгална.
func UserOTPKey(email string) string {
	return fmt.Sprintf("%s%s", prefixUserOTP, email)
}

// OTPAttemptsKey нь email тус бүрийн амжилтгүй VerifyOTP оролдлогуудыг тоолно.
func OTPAttemptsKey(email string) string {
	return fmt.Sprintf("%s%s", prefixOTPAttempts, email)
}

// LoginAttemptsKey нь brute-force түгжих цонхонд зориулж амжилтгүй Login
// оролдлогуудыг тоолно. IP өгөгдсөн бол (email, IP)-ээр түлхүүрлэнэ — нэг
// халдагч хохирогчийн email-ийг ГЛОБАЛЬ түгжих DoS-аас сэргийлнэ (хохирогч
// өөрийн IP-ээс саадгүй нэвтэрнэ). IP хоосон бол email-ээр (хуучин зан төлөв).
func LoginAttemptsKey(email, ip string) string {
	if ip == "" {
		return fmt.Sprintf("%s%s", prefixLoginAttempts, email)
	}
	return fmt.Sprintf("%s%s:%s", prefixLoginAttempts, email, ip)
}

// ForgotAttemptsKey нь email тус бүрд /password/forgot-ийг rate-limit хийнэ.
func ForgotAttemptsKey(email string) string {
	return fmt.Sprintf("%s%s", prefixForgotAttempts, email)
}

// PasswordResetKey нь нэг удаагийн reset токены user ID-г хадгална; токен өөрөө
// дагавар (suffix) болдог.
func PasswordResetKey(token string) string {
	return fmt.Sprintf("%s%s", prefixPasswordReset, token)
}

// UserResetIndexKey нь user ID-аас одоо хүчинтэй байгаа токен руу чиглэсэн,
// хэрэглэгч тус бүрийн урвуу индекс юм. Шинэ reset холбоос олгох нь өмнөх
// амьд токеныг хүчингүй болгох зорилгоор ашиглагдана.
func UserResetIndexKey(userID string) string {
	return fmt.Sprintf("%s%s", prefixUserResetIndex, userID)
}

// TokenCutoffKey нь энэ хэрэглэгчид олгогдсон аливаа access токеныг хүчингүй
// гэж тооцох тасалбар цэгийг unix-секундээр хадгална. Auth middleware үүнийг
// баталгаажсан хүсэлт бүр дээр уншдаг; ChangePassword болон ResetPassword нь
// үүнийг бичдэг.
func TokenCutoffKey(userID string) string {
	return fmt.Sprintf("%s%s", prefixPasswordCutoff, userID)
}
