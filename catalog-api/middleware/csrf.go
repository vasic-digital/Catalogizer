package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CSRF is a minimal double-submit-cookie CSRF guard suitable for the
// catalog-api admin surface. It does not depend on session storage:
// callers present the same random token in both a cookie and the
// X-CSRF-Token header, and the middleware HMACs both with a server
// secret to compare.
//
// For full SameSite=strict + session-bound CSRF, integrate
// github.com/gorilla/csrf into the production deployment. This
// middleware ships as a lightweight default for deployments that
// cannot adopt gorilla/csrf (Gin uses different idioms for the token
// stream).
type CSRF struct {
	secret []byte
	cookie string
	header string
	maxAge time.Duration
}

// CSRFOption tunes the middleware at construction.
type CSRFOption func(*CSRF)

// WithCSRFCookieName overrides the cookie name (default "__Host-csrf").
func WithCSRFCookieName(name string) CSRFOption { return func(c *CSRF) { c.cookie = name } }

// WithCSRFHeaderName overrides the header name (default "X-CSRF-Token").
func WithCSRFHeaderName(name string) CSRFOption { return func(c *CSRF) { c.header = name } }

// WithCSRFMaxAge sets the token validity window (default 24h).
func WithCSRFMaxAge(d time.Duration) CSRFOption { return func(c *CSRF) { c.maxAge = d } }

// NewCSRF returns a middleware bound to a HMAC secret.
func NewCSRF(secret []byte, opts ...CSRFOption) (*CSRF, error) {
	if len(secret) < 16 {
		return nil, errors.New("csrf: secret must be at least 16 bytes")
	}
	c := &CSRF{
		secret: append([]byte(nil), secret...),
		cookie: "__Host-csrf",
		header: "X-CSRF-Token",
		maxAge: 24 * time.Hour,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// Handler returns a gin.HandlerFunc that mints tokens on GET + HEAD
// requests and validates them on every mutating method.
func (c *CSRF) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		method := strings.ToUpper(ctx.Request.Method)
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			if ctx.GetHeader(c.header) == "" && ctx.Writer.Header().Get("X-CSRF-Token") == "" {
				c.mintToken(ctx)
			}
			ctx.Next()
			return
		}
		cookie, err := ctx.Cookie(c.cookie)
		if err != nil || cookie == "" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "missing csrf cookie"})
			return
		}
		header := ctx.GetHeader(c.header)
		if header == "" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "missing csrf header"})
			return
		}
		if !c.validate(cookie, header) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "csrf token mismatch"})
			return
		}
		ctx.Next()
	}
}

func (c *CSRF) mintToken(ctx *gin.Context) {
	raw := make([]byte, 32)
	_, _ = rand.Read(raw)
	token := hex.EncodeToString(raw)
	signed := c.sign(token)
	ctx.SetCookie(
		c.cookie,
		signed,
		int(c.maxAge.Seconds()),
		"/", "",
		true, // secure — only over TLS
		true, // httpOnly
	)
	ctx.Writer.Header().Set(c.header, token)
}

func (c *CSRF) sign(token string) string {
	mac := hmac.New(sha256.New, c.secret)
	mac.Write([]byte(token))
	return token + ":" + hex.EncodeToString(mac.Sum(nil))
}

func (c *CSRF) validate(cookie, header string) bool {
	parts := strings.SplitN(cookie, ":", 2)
	if len(parts) != 2 {
		return false
	}
	token, sig := parts[0], parts[1]
	if subtleCompare(token, header) != 1 {
		return false
	}
	wantSig := c.sign(token)
	// wantSig = token + ":" + sig2 — strip the token+colon prefix
	expected := wantSig[len(token)+1:]
	return subtleCompare(sig, expected) == 1
}

// subtleCompare is a constant-time comparison that returns 1 on
// equality and 0 otherwise. We keep it local so the middleware file
// does not pull in crypto/subtle just for one call.
func subtleCompare(a, b string) int {
	if len(a) != len(b) {
		return 0
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	if diff == 0 {
		return 1
	}
	return 0
}
