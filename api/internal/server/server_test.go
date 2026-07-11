package server_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"testing"
	"time"

	jose "github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/test/gtest"

	"platform/gokit/authjwt"
	"platform/products/resource/api/internal/server"
)

const (
	testIssuer = "http://localhost:8081"
	testKID    = "resource-test-kid"
	testSub    = "11111111-1111-1111-1111-111111111111"
	testSub2   = "22222222-2222-2222-2222-222222222222"
)

func prefix(s *ghttp.Server) string {
	return fmt.Sprintf("http://127.0.0.1:%d", s.GetListenedPort())
}

func mustVerifier(t *gtest.T, priv *rsa.PrivateKey) *authjwt.Verifier {
	set := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{
		Key: priv.Public(), KeyID: testKID, Algorithm: "RS256", Use: "sig",
	}}}
	v, err := authjwt.NewVerifier(authjwt.VerifierConfig{
		Keys:   authjwt.NewStaticKeySource(set),
		Issuer: testIssuer,
	})
	t.AssertNil(err)
	return v
}

func signToken(t *gtest.T, priv *rsa.PrivateKey, sub string, exp time.Time) string {
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: priv},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", testKID),
	)
	t.AssertNil(err)
	now := time.Now().UTC()
	raw, err := jwt.Signed(signer).Claims(jwt.Claims{
		Issuer:   testIssuer,
		Subject:  sub,
		IssuedAt: jwt.NewNumericDate(now.Add(-time.Minute)),
		Expiry:   jwt.NewNumericDate(exp),
	}).CompactSerialize()
	t.AssertNil(err)
	return raw
}

func TestHealthzPublic(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		t.AssertNil(err)
		s := g.Server(t.Name())
		s.SetAddr("127.0.0.1:0")
		server.Configure(s, server.Deps{Verifier: mustVerifier(t, priv)})
		s.SetDumpRouterMap(false)
		s.Start()
		defer s.Shutdown()

		c := g.Client()
		c.SetPrefix(prefix(s))
		resp, err := c.Get(context.Background(), "/healthz")
		t.AssertNil(err)
		defer resp.Close()
		t.Assert(resp.StatusCode, 200)
		t.Assert(gjson.New(resp.ReadAllString()).Get("data.status").String(), "up")
	})
}
