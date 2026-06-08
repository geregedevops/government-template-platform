// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package fedsign нь федерацийн node хооронд солилцох мессежийг ES256 (EC P-256)
// гарын үсгээр баталгаажуулна. Энэ нь JWT-ийн (хэрэглэгчийн токен) түлхүүрээс
// ТУСДАА node-ийн "e-seal" түлхүүр — node бүр өөрийн EC түлхүүртэй, нийтийн
// түлхүүрээ JWKS-ээр түгээж, бусад node гарын үсгийг шалгана. (Ирээдүйд eID
// Mongolia CA + OCSP/CRL-ээр сольж болно.)
//
// Мессеж нь ES256-аар гарын үсэглэсэн JWT хэлбэртэй "envelope": claims нь
// jti (давталт хаах), iss (илгээгч node), aud (хүлээн авагч node), iat, мөн
// typ + body (мессежийн агуулга).
package fedsign

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
	"strings"
	"time"

	golangJWT "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Envelope нь федерацийн мессежийн гарын үсэгтэй дугтуй (signed JWT claims).
type Envelope struct {
	Typ  string          `json:"typ"`
	Body json.RawMessage `json:"body"`
	golangJWT.RegisteredClaims
}

// Signer нь node-ийн гарын үсэг зурагч (хувийн түлхүүр эзэмшинэ).
type Signer struct {
	priv   *ecdsa.PrivateKey
	kid    string
	nodeID string
}

// NewSigner нь PEM (PKCS#8/SEC1, P-256) EC хувийн түлхүүрээс signer үүсгэнэ.
func NewSigner(ecPEM, kid, nodeID string) (*Signer, error) {
	priv, err := parseECPrivate(ecPEM)
	if err != nil {
		return nil, err
	}
	if kid == "" {
		kid = "fed"
	}
	return &Signer{priv: priv, kid: kid, nodeID: nodeID}, nil
}

func (s *Signer) NodeID() string { return s.nodeID }
func (s *Signer) KID() string    { return s.kid }

// Sign нь typ + body-той envelope үүсгэж ES256-аар гарын үсэглэнэ. jti нь
// давталт хаах (idempotency) түлхүүр болно.
func (s *Signer) Sign(typ, audience string, body json.RawMessage, now time.Time) (token, jti string, err error) {
	jti = uuid.NewString()
	env := Envelope{
		Typ:  typ,
		Body: body,
		RegisteredClaims: golangJWT.RegisteredClaims{
			ID:        jti,
			Issuer:    s.nodeID,
			Audience:  golangJWT.ClaimStrings{audience},
			IssuedAt:  golangJWT.NewNumericDate(now),
			ExpiresAt: golangJWT.NewNumericDate(now.Add(10 * time.Minute)),
		},
	}
	t := golangJWT.NewWithClaims(golangJWT.SigningMethodES256, env)
	t.Header["kid"] = s.kid
	signed, err := t.SignedString(s.priv)
	if err != nil {
		return "", "", fmt.Errorf("fedsign: sign: %w", err)
	}
	return signed, jti, nil
}

// JWKS нь node-ийн нийтийн түлхүүрийг JWK Set JSON болгож буцаана.
func (s *Signer) JWKS() ([]byte, error) {
	return marshalJWKS(&s.priv.PublicKey, s.kid)
}

// PeekUnverified нь гарын үсэг шалгахаас ӨМНӨ header-ийн kid болон claims-ийн
// iss (илгээгч)-ийг гаргаж авна — registry-ээс зөв түлхүүрийг олоход хэрэгтэй.
func PeekUnverified(token string) (kid, iss string, err error) {
	p := golangJWT.NewParser()
	var env Envelope
	tok, _, perr := p.ParseUnverified(token, &env)
	if perr != nil {
		return "", "", fmt.Errorf("fedsign: peek: %w", perr)
	}
	if k, ok := tok.Header["kid"].(string); ok {
		kid = k
	}
	return kid, env.Issuer, nil
}

// Verify нь envelope-ийг өгөгдсөн нийтийн түлхүүрээр (ES256) шалгаж, aud нь
// selfNodeID-тэй таарч буйг шалгана.
func Verify(token string, pub *ecdsa.PublicKey, selfNodeID string) (Envelope, error) {
	var env Envelope
	_, err := golangJWT.ParseWithClaims(token, &env, func(t *golangJWT.Token) (interface{}, error) {
		if t.Method.Alg() != golangJWT.SigningMethodES256.Alg() {
			return nil, fmt.Errorf("fedsign: unexpected alg %v", t.Header["alg"])
		}
		return pub, nil
	})
	if err != nil {
		return Envelope{}, fmt.Errorf("fedsign: verify: %w", err)
	}
	okAud := false
	for _, a := range env.Audience {
		if a == selfNodeID {
			okAud = true
		}
	}
	if !okAud {
		return Envelope{}, errors.New("fedsign: audience mismatch")
	}
	return env, nil
}

// PublicKeyFromJWKS нь JWK Set JSON-оос өгөгдсөн kid-тэй EC нийтийн түлхүүрийг
// (kid хоосон бол эхнийхийг) задлан гаргана.
func PublicKeyFromJWKS(jwksJSON []byte, kid string) (*ecdsa.PublicKey, error) {
	var set struct {
		Keys []map[string]string `json:"keys"`
	}
	if err := json.Unmarshal(jwksJSON, &set); err != nil {
		return nil, fmt.Errorf("fedsign: jwks parse: %w", err)
	}
	for _, k := range set.Keys {
		if kid != "" && k["kid"] != kid {
			continue
		}
		if k["kty"] != "EC" || k["crv"] != "P-256" {
			continue
		}
		x, err := b64ToInt(k["x"])
		if err != nil {
			return nil, err
		}
		y, err := b64ToInt(k["y"])
		if err != nil {
			return nil, err
		}
		return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
	}
	return nil, errors.New("fedsign: no matching EC key in JWKS")
}

// --- helpers ---

func parseECPrivate(ecPEM string) (*ecdsa.PrivateKey, error) {
	// Нэг мөрт env var-д тааруулахын тулд base64(PEM)-ийг ч хүлээж авна:
	// PEM толгой байхгүй бол base64 гэж үзэж задална.
	if !strings.Contains(ecPEM, "BEGIN") {
		if dec, derr := base64.StdEncoding.DecodeString(strings.TrimSpace(ecPEM)); derr == nil {
			ecPEM = string(dec)
		}
	}
	block, _ := pem.Decode([]byte(ecPEM))
	if block == nil {
		return nil, errors.New("fedsign: invalid EC private key PEM")
	}
	if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		ec, ok := k.(*ecdsa.PrivateKey)
		if !ok {
			return nil, errors.New("fedsign: PKCS#8 key is not ECDSA")
		}
		if ec.Curve != elliptic.P256() {
			return nil, errors.New("fedsign: EC key must be P-256")
		}
		return ec, nil
	}
	ec, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("fedsign: parse EC key: %w", err)
	}
	if ec.Curve != elliptic.P256() {
		return nil, errors.New("fedsign: EC key must be P-256")
	}
	return ec, nil
}

func marshalJWKS(pub *ecdsa.PublicKey, kid string) ([]byte, error) {
	b64 := func(b *big.Int) string {
		buf := make([]byte, 32)
		bb := b.Bytes()
		copy(buf[32-len(bb):], bb)
		return base64.RawURLEncoding.EncodeToString(buf)
	}
	jwk := map[string]any{
		"kty": "EC", "crv": "P-256", "use": "sig", "alg": "ES256",
		"kid": kid, "x": b64(pub.X), "y": b64(pub.Y),
	}
	return json.Marshal(map[string]any{"keys": []any{jwk}})
}

func b64ToInt(s string) (*big.Int, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("fedsign: b64 decode: %w", err)
	}
	return new(big.Int).SetBytes(raw), nil
}
