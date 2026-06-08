// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package jwt_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"testing"

	"geregetemplateai/pkg/jwt"
)

func ecPEM(t *testing.T) string {
	t.Helper()
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("gen ec key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(k)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

func TestES256RoundTripAndJWKS(t *testing.T) {
	base := jwt.NewJWTServiceWithRefresh("0123456789abcdef0123456789abcdef", "gerege", 1, 7)
	svc, err := jwt.EnableES256(base, ecPEM(t), "k1")
	if err != nil {
		t.Fatalf("EnableES256: %v", err)
	}

	// Sign + parse roundtrip (ES256).
	tok, err := svc.GenerateToken("u1", false, 2, "a@b.mn", "org1")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	claims, err := svc.ParseToken(tok)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.UserID != "u1" || claims.OrgID != "org1" {
		t.Fatalf("claims mismatch: %+v", claims)
	}

	// JWKS contains one EC P-256 key.
	raw, ok := svc.JWKSet()
	if !ok {
		t.Fatal("JWKSet should be available in ES256 mode")
	}
	var set struct {
		Keys []map[string]string `json:"keys"`
	}
	if err := json.Unmarshal(raw, &set); err != nil || len(set.Keys) != 1 {
		t.Fatalf("bad JWKS: %v %s", err, raw)
	}
	k := set.Keys[0]
	if k["kty"] != "EC" || k["crv"] != "P-256" || k["alg"] != "ES256" || k["kid"] != "k1" || k["x"] == "" || k["y"] == "" {
		t.Fatalf("bad JWK: %+v", k)
	}

	// alg-confusion: an HS256 token must be REJECTED by the ES256-pinned parser.
	hs := jwt.NewJWTServiceWithRefresh("0123456789abcdef0123456789abcdef", "gerege", 1, 7)
	hsTok, _ := hs.GenerateToken("u1", false, 2, "a@b.mn", "org1")
	if _, err := svc.ParseToken(hsTok); err == nil {
		t.Fatal("ES256 parser must reject an HS256 token (alg confusion)")
	}

	// HS256 mode has no JWKS.
	if _, ok := base.JWKSet(); ok {
		t.Fatal("HS256 mode must not expose a JWKS")
	}
}
