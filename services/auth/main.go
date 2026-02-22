package main

import (
	"crypto/rsa"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	Sub   string   `json:"sub"`
	Roles []string `json:"roles"`
	Scope string   `json:"scope,omitempty"`
	jwt.RegisteredClaims
}

var privateKey *rsa.PrivateKey

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return jwt.ParseRSAPrivateKeyFromPEM(b)
}

func issueAdminToken(adminUserID string) (string, error) {
	now := time.Now()

	claims := CustomClaims{
		Sub:   adminUserID,
		Roles: []string{"admin"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "auth-service",
			Audience:  []string{"delivery-service"},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-30 * time.Second)),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return t.SignedString(privateKey)
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	var err error
	privateKey, err = loadPrivateKey("private.pem")
	if err != nil {
		panic("failed to load private key: " + err.Error())
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Demo: admin login -> admin JWT for webhook-delivery-service
	r.POST("/admin/login", func(c *gin.Context) {
		var req loginReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
			return
		}

		// Demo credentials check (replace with DB lookup + password hash verify)
		if req.Username != "admin" || req.Password != "admin123" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		token, err := issueAdminToken("admin-1")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
			return
		}

		c.JSON(200, gin.H{
			"access_token": token,
			"token_type":   "Bearer",
			"expires_in":   900,
		})
	})

	_ = r.Run(":8081") //TODO: make port configurable
}
