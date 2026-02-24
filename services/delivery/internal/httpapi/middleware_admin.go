package httpapi

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AdminClaims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func (h *SourcesHandler) adminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.adminPublicKey == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "admin auth not configured"})
			c.Abort()
			return
		}

		tokenString, ok := parseBearerToken(c.GetHeader("Authorization"))
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization"})
			c.Abort()
			return
		}

		claims := &AdminClaims{}
		token, err := jwt.ParseWithClaims(
			tokenString,
			claims,
			func(t *jwt.Token) (any, error) {
				if t.Method == nil || t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
					return nil, jwt.ErrTokenSignatureInvalid
				}
				return h.adminPublicKey, nil
			},
			jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}),
			jwt.WithIssuer("auth-service"),
			jwt.WithAudience("delivery-service"),
		)
		if err != nil || token == nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		if !slices.Contains(claims.Roles, "admin") {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			c.Abort()
			return
		}

		c.Set("admin_claims", claims)
		c.Next()
	}
}
