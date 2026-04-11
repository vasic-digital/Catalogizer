package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"catalogizer/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware handles JWT authentication
type JWTMiddleware struct {
	secretKey []byte
}

// Claims represents JWT claims
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// NewJWTMiddleware creates a new JWT middleware
func NewJWTMiddleware(secretKey string) *JWTMiddleware {
	return &JWTMiddleware{
		secretKey: []byte(secretKey),
	}
}

// RequireAuth returns a middleware that requires valid JWT authentication.
// Accepts the token either as a standard "Authorization: Bearer <token>"
// header OR as an "access_token" / "token" query parameter. The query
// parameter path exists so native media players (libVLC, ffmpeg, VLC
// iOS, etc.) that cannot easily inject custom headers into their HTTP
// data source can still reach the authenticated /api/v1/stream/:id
// endpoint by appending "?access_token=<jwt>" to the URL.
func (m *JWTMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				utils.SendErrorResponse(c, http.StatusUnauthorized, "Invalid authorization header format", nil)
				c.Abort()
				return
			}
			tokenString = parts[1]
		} else {
			// Fallback to query parameter for native clients.
			tokenString = c.Query("access_token")
			if tokenString == "" {
				tokenString = c.Query("token")
			}
			if tokenString == "" {
				utils.SendErrorResponse(c, http.StatusUnauthorized, "Authorization header required", nil)
				c.Abort()
				return
			}
		}

		// Parse and validate token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return m.secretKey, nil
		})

		if err != nil {
			utils.SendErrorResponse(c, http.StatusUnauthorized, "Invalid token", err)
			c.Abort()
			return
		}

		if !token.Valid {
			utils.SendErrorResponse(c, http.StatusUnauthorized, "Invalid token", nil)
			c.Abort()
			return
		}

		// Set user info in context — convert user_id to int for handler compatibility
		c.Set("username", claims.Username)
		if uid, err := strconv.Atoi(claims.Subject); err == nil {
			c.Set("user_id", uid)
		} else {
			c.Set("user_id", claims.Subject)
		}

		c.Next()
	}
}

// GenerateToken generates a new JWT token
func (m *JWTMiddleware) GenerateToken(username, userID string, expirationHours int) (string, error) {
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "catalog-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// ValidateToken validates a JWT token and returns claims
func (m *JWTMiddleware) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return m.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenNotValidYet
	}

	return claims, nil
}
