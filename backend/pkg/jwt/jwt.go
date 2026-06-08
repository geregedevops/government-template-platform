// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"

	"geregetemplateai/pkg/clock"
	"geregetemplateai/pkg/logger"
	golangJWT "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ErrInvalidToken нь токен задлан унших эсвэл баталгаажуулахад амжилтгүй болоход буцаагдана.
var ErrInvalidToken = errors.New("token is not valid")

// ErrWrongTokenKind нь дуудагч access токеныг refresh токен мэтээр (эсвэл
// эсрэгээр) задлан уншихыг оролдоход буцаагдана.
var ErrWrongTokenKind = errors.New("token kind mismatch")

// Kind-ууд нь access болон refresh токеныг ялгана. Claim дотор гарын үсэг
// зурагдсан тул эвдэрсэн refresh токеныг access токен болгон дахин ашиглах боломжгүй.
const (
	KindAccess  = "access"
	KindRefresh = "refresh"
)

type JWTService interface {
	// GenerateToken нь нэг access токен үүсгэнэ. Дуудагчид зэрэгцээ refresh
	// токен хэрэгтэй бол GenerateTokenPair-г илүүд үзнэ.
	GenerateToken(userId string, isAdmin bool, roleID int, email, orgID string) (t string, err error)
	// GenerateTokenPair нь access+refresh хосыг үүсгэнэ, хоёулаа ижил secret-ээр
	// гарын үсэг зурагдсан боловч Kind claim-ээр ялгагдана.
	GenerateTokenPair(userID string, isAdmin bool, roleID int, email, orgID string) (TokenPair, error)
	// ParseToken нь access токены гарын үсэг, хүчинтэй хугацаа болон HMAC аргыг
	// шалгана. Refresh токенуудыг ErrWrongTokenKind-ээр татгалзана.
	ParseToken(tokenString string) (claims JwtCustomClaim, err error)
	// ParseRefreshToken нь ParseToken-ийн refresh токены эквивалент юм.
	// Access токенуудыг ErrWrongTokenKind-ээр татгалзана.
	ParseRefreshToken(tokenString string) (claims JwtCustomClaim, err error)
	// JWKSet нь ES256 (asymmetric) горимд нийтийн түлхүүрийг JWK Set (RFC 7517)
	// JSON болгож буцаана — федерацийн бусад node токеныг үүгээр баталгаажуулна.
	// HS256 (symmetric) горимд ok=false (нийтлэх нийтийн түлхүүр байхгүй).
	JWKSet() (json []byte, ok bool)
}

// TokenPair нь login / refresh үед хамт олгогддог богино настай access
// токен болон урт настай refresh токеныг багцална.
type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	AccessJTI        string    `json:"-"`
	RefreshJTI       string    `json:"-"`
}

type JwtCustomClaim struct {
	UserID  string
	IsAdmin bool
	RoleID  int
	OrgID   string
	Email   string
	Kind    string
	golangJWT.RegisteredClaims
}

type jwtService struct {
	secretKey      string
	issuer         string
	expired        int
	refreshExpired int // өдөр
	// clock нь IssuedAt болон хүчинтэй хугацааны тооцоололд ашиглагдах "одоо"-гийн
	// эх сурвалж юм. Анхдагч нь RealClock; тестүүд унтахгүйгээр яг таг цагийн
	// тэмдгийг шалгахын тулд clock.Frozen эсвэл clock.Stub-г тарьдаг.
	clock clock.Clock
	// ecPriv тохируулагдсан бол ES256 (asymmetric) горим: токеныг EC хувийн
	// түлхүүрээр гарын үсэглэж, нийтийн түлхүүрээр баталгаажуулна, JWKS нийтэлнэ.
	// nil бол HS256 (symmetric, анхдагч) — secretKey ашиглана.
	ecPriv *ecdsa.PrivateKey
	kid    string
}

// EnableES256 нь HS256 сервисийн ES256 (asymmetric) хувилбарыг буцаана: PEM
// (PKCS#8 эсвэл SEC1) EC хувийн түлхүүрийг ачаалж клон үүсгэнэ. kid хоосон бол
// "gerege". Энэ нь федерацийн урьдчилсан нөхцөл (бусад node JWKS-ээр шалгана).
func EnableES256(svc JWTService, ecPEM, kid string) (JWTService, error) {
	s, ok := svc.(*jwtService)
	if !ok {
		return svc, errors.New("jwt: EnableES256 requires the default service")
	}
	block, _ := pem.Decode([]byte(ecPEM))
	if block == nil {
		return svc, errors.New("jwt: invalid EC private key PEM")
	}
	var priv *ecdsa.PrivateKey
	if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		ec, ok := k.(*ecdsa.PrivateKey)
		if !ok {
			return svc, errors.New("jwt: PKCS#8 key is not ECDSA")
		}
		priv = ec
	} else if k, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		priv = k
	} else {
		return svc, fmt.Errorf("jwt: parse EC private key: %w", err)
	}
	if priv.Curve != elliptic.P256() {
		return svc, errors.New("jwt: EC key must use the P-256 curve (ES256)")
	}
	clone := *s
	clone.ecPriv = priv
	if kid == "" {
		kid = "gerege"
	}
	clone.kid = kid
	return &clone, nil
}

// JWKSet нь ES256 горимд нийтийн түлхүүрийг JWK Set JSON болгож буцаана.
func (j *jwtService) JWKSet() ([]byte, bool) {
	if j.ecPriv == nil {
		return nil, false
	}
	pub := j.ecPriv.PublicKey
	b64 := func(b *big.Int) string {
		// P-256: координат бүр 32 байт (зүүн талаас 0-ээр нөхөнө).
		buf := make([]byte, 32)
		bb := b.Bytes()
		copy(buf[32-len(bb):], bb)
		return base64.RawURLEncoding.EncodeToString(buf)
	}
	jwk := map[string]any{
		"kty": "EC", "crv": "P-256", "use": "sig", "alg": "ES256",
		"kid": j.kid, "x": b64(pub.X), "y": b64(pub.Y),
	}
	out, err := json.Marshal(map[string]any{"keys": []any{jwk}})
	if err != nil {
		return nil, false
	}
	return out, true
}

func NewJWTService(secretKey, issuer string, expired int) JWTService {
	return &jwtService{
		issuer:         issuer,
		secretKey:      secretKey,
		expired:        expired,
		refreshExpired: 7,
		clock:          clock.RealClock{},
	}
}

// NewJWTServiceWithRefresh нь тус тусдаа тохируулж болох настай access +
// refresh токены хосыг үүсгэдэг сервис байгуулна.
func NewJWTServiceWithRefresh(secretKey, issuer string, expiredHours, refreshExpiredDays int) JWTService {
	return &jwtService{
		issuer:         issuer,
		secretKey:      secretKey,
		expired:        expiredHours,
		refreshExpired: refreshExpiredDays,
		clock:          clock.RealClock{},
	}
}

// WithClock нь өгөгдсөн clock-оор орлуулсан сервисийн хуулбарыг буцаана.
// Тестүүд токен олголтын үеийн цагийг царцаах (freeze) болон яг таг ExpiresAt
// утгуудыг шалгахын тулд үүнийг ашигладаг.
func WithClock(svc JWTService, c clock.Clock) JWTService {
	if s, ok := svc.(*jwtService); ok {
		clone := *s
		clone.clock = c
		return &clone
	}
	return svc
}

func (j *jwtService) GenerateToken(userID string, isAdmin bool, roleID int, email, orgID string) (string, error) {
	tok, _, _, err := j.signAccess(userID, isAdmin, roleID, email, orgID)
	return tok, err
}

func (j *jwtService) GenerateTokenPair(userID string, isAdmin bool, roleID int, email, orgID string) (TokenPair, error) {
	access, accessExp, accessJTI, err := j.signAccess(userID, isAdmin, roleID, email, orgID)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, refreshExp, refreshJTI, err := j.signRefresh(userID, email)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:      access,
		RefreshToken:     refresh,
		AccessExpiresAt:  accessExp,
		RefreshExpiresAt: refreshExp,
		AccessJTI:        accessJTI,
		RefreshJTI:       refreshJTI,
	}, nil
}

func (j *jwtService) signAccess(userID string, isAdmin bool, roleID int, email, orgID string) (token string, expiresAt time.Time, jti string, err error) {
	now := j.clock.Now()
	expiresAt = now.Add(time.Hour * time.Duration(j.expired))
	jti = uuid.NewString()
	claims := &JwtCustomClaim{
		UserID:  userID,
		IsAdmin: isAdmin,
		RoleID:  roleID,
		OrgID:   orgID,
		Email:   email,
		Kind:    KindAccess,
		RegisteredClaims: golangJWT.RegisteredClaims{
			ID:        jti,
			ExpiresAt: golangJWT.NewNumericDate(expiresAt),
			Issuer:    j.issuer,
			IssuedAt:  golangJWT.NewNumericDate(now),
		},
	}
	token, err = j.sign(claims)
	if err != nil {
		return "", time.Time{}, "", err
	}
	return token, expiresAt, jti, nil
}

func (j *jwtService) signRefresh(userID, email string) (token string, expiresAt time.Time, jti string, err error) {
	now := j.clock.Now()
	expiresAt = now.Add(24 * time.Hour * time.Duration(j.refreshExpired))
	jti = uuid.NewString()
	claims := &JwtCustomClaim{
		UserID: userID,
		Email:  email,
		Kind:   KindRefresh,
		RegisteredClaims: golangJWT.RegisteredClaims{
			ID:        jti,
			ExpiresAt: golangJWT.NewNumericDate(expiresAt),
			Issuer:    j.issuer,
			IssuedAt:  golangJWT.NewNumericDate(now),
		},
	}
	token, err = j.sign(claims)
	if err != nil {
		return "", time.Time{}, "", err
	}
	return token, expiresAt, jti, nil
}

func (j *jwtService) sign(claims *JwtCustomClaim) (string, error) {
	// ES256 (asymmetric) горим тохируулагдсан бол EC хувийн түлхүүрээр,
	// эс бөгөөс HS256 (symmetric, анхдагч) secretKey-ээр гарын үсэглэнэ.
	if j.ecPriv != nil {
		token := golangJWT.NewWithClaims(golangJWT.SigningMethodES256, claims)
		token.Header["kid"] = j.kid
		signed, err := token.SignedString(j.ecPriv)
		if err != nil {
			logger.Error("jwt: ES256 sign failed", logger.Fields{"package": "jwt", "step": "signed_string", "kind": claims.Kind, "error": err.Error()})
			return "", fmt.Errorf("sign jwt: %w", err)
		}
		return signed, nil
	}
	token := golangJWT.NewWithClaims(golangJWT.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		logger.Error("jwt: sign failed", logger.Fields{
			"package": "jwt",
			"step":    "signed_string",
			"kind":    claims.Kind,
			"error":   err.Error(),
		})
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}

func (j *jwtService) ParseToken(tokenString string) (JwtCustomClaim, error) {
	claims, err := j.parse(tokenString)
	if err != nil {
		return JwtCustomClaim{}, err
	}
	// Хоосон Kind-г access токен гэж хүлээн авна; зөвхөн илэрхий access биш
	// утгыг (жишээ нь KindRefresh) энд татгалзана.
	if claims.Kind != "" && claims.Kind != KindAccess {
		return JwtCustomClaim{}, ErrWrongTokenKind
	}
	return claims, nil
}

func (j *jwtService) ParseRefreshToken(tokenString string) (JwtCustomClaim, error) {
	claims, err := j.parse(tokenString)
	if err != nil {
		return JwtCustomClaim{}, err
	}
	if claims.Kind != KindRefresh {
		return JwtCustomClaim{}, ErrWrongTokenKind
	}
	return claims, nil
}

func (j *jwtService) parse(tokenString string) (JwtCustomClaim, error) {
	var claims JwtCustomClaim
	token, err := golangJWT.ParseWithClaims(tokenString, &claims, func(token *golangJWT.Token) (interface{}, error) {
		// Гарын үсгийн аргыг ЯГ тохируулсан горимоор тогтооно (alg-confusion-
		// аас сэргийлнэ): ES256 горимд зөвхөн ES256 + нийтийн түлхүүр, HS256
		// горимд зөвхөн HS256 + secret. Толгойн alg-ийг өөрчилж "downgrade"
		// хийх боломжгүй.
		if j.ecPriv != nil {
			if token.Method.Alg() != golangJWT.SigningMethodES256.Alg() {
				logger.Warn("jwt: unexpected signing method", logger.Fields{"package": "jwt", "step": "verify_signing_method", "alg": fmt.Sprintf("%v", token.Header["alg"])})
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return &j.ecPriv.PublicKey, nil
		}
		if token.Method.Alg() != golangJWT.SigningMethodHS256.Alg() {
			alg := token.Header["alg"]
			logger.Warn("jwt: unexpected signing method", logger.Fields{
				"package": "jwt",
				"step":    "verify_signing_method",
				"alg":     fmt.Sprintf("%v", alg),
			})
			return nil, fmt.Errorf("unexpected signing method: %v", alg)
		}
		return []byte(j.secretKey), nil
	})
	if err != nil {
		logger.Warn("jwt: parse failed", logger.Fields{
			"package": "jwt",
			"step":    "parse_with_claims",
			"error":   err.Error(),
		})
		return JwtCustomClaim{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if !token.Valid {
		logger.Warn("jwt: token reported invalid by parser", logger.Fields{
			"package": "jwt",
			"step":    "validity_check",
		})
		return JwtCustomClaim{}, ErrInvalidToken
	}
	return claims, nil
}
