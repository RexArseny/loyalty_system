package middlewares

import (
	"crypto"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	Authorization = "Authorization"
	UserID        = "UserID"
	maxAge        = 900
)

var ErrNoJWT = errors.New("no jwt in cookie")

type Middleware struct {
	publicKey  crypto.PublicKey
	privateKey crypto.PrivateKey
	logger     *zap.Logger
}

func NewMiddleware(publicKeyPath string, privateKeyPath string, logger *zap.Logger) (*Middleware, error) {
	publicKeyFile, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("can not open public.pem file: %w", err)
	}
	publicKey, err := jwt.ParseEdPublicKeyFromPEM(publicKeyFile)
	if err != nil {
		return nil, fmt.Errorf("can not parse public key: %w", err)
	}

	privateKeyFile, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("can not open private.pem file: %w", err)
	}
	privateKey, err := jwt.ParseEdPrivateKeyFromPEM(privateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("can not parse private key: %w", err)
	}

	return &Middleware{
		publicKey:  publicKey,
		privateKey: privateKey,
		logger:     logger,
	}, nil
}

func (m *Middleware) Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		raw := ctx.Request.URL.RawQuery

		ctx.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		m.logger.Info("Request",
			zap.Int("code", ctx.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("method", ctx.Request.Method),
			zap.String("path", path),
			zap.Int("size", ctx.Writer.Size()))
	}
}

type JWT struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
}

func (m *Middleware) SetJWT() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		userIDValue, ok := ctx.Get(UserID)
		if !ok {
			return
		}
		userID, ok := userIDValue.(*uuid.UUID)
		if !ok || userID == nil {
			return
		}

		claims := &JWT{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "loyalty_system",
				Subject:   userID.String(),
				Audience:  jwt.ClaimStrings{},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * maxAge)),
				NotBefore: jwt.NewNumericDate(time.Now()),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ID:        uuid.New().String(),
			},
			UserID: *userID,
		}

		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

		tokenString, err := token.SignedString(m.privateKey)
		if err != nil {
			m.logger.Error("Can not sign token", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
			ctx.Abort()
			return
		}

		ctx.SetCookie(
			Authorization,
			tokenString,
			maxAge,
			"/",
			"",
			false,
			false,
		)

		ctx.JSON(http.StatusOK, gin.H{"status": http.StatusText(http.StatusOK)})
	}
}

func (m *Middleware) GetJWT() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString, err := ctx.Cookie(Authorization)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
			ctx.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(
			tokenString,
			&JWT{},
			func(token *jwt.Token) (interface{}, error) {
				if token.Method != jwt.SigningMethodEdDSA {
					return nil, errors.New("jwt signature mismatch")
				}
				return m.publicKey, nil
			},
		)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(*JWT)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": http.StatusText(http.StatusUnauthorized)})
			ctx.Abort()
			return
		}

		ctx.Set(Authorization, claims)

		ctx.Next()
	}
}
