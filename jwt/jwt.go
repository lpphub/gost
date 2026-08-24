package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	Secret           string
	AccessExpireSec  int64
	RefreshExpireSec int64
}

type Manager struct {
	cfg    Config
	secret []byte
}

func NewManager(cfg Config) (*Manager, error) {
	if cfg.Secret == "" {
		return nil, errors.New("JWT secret cannot be empty")
	}
	if cfg.AccessExpireSec <= 0 {
		cfg.AccessExpireSec = 7200
	}
	if cfg.RefreshExpireSec <= 0 {
		cfg.RefreshExpireSec = 604800
	}

	return &Manager{cfg: cfg, secret: []byte(cfg.Secret)}, nil
}

type TokenType string

const (
	AccessTokenType  TokenType = "access"
	RefreshTokenType TokenType = "refresh"
)

var ErrInvalidTokenType = errors.New("invalid token type")

type Claims struct {
	UserID uint      `json:"user_id"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (m *Manager) generateToken(userID uint, tokenType TokenType, expireSec int64) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Second * time.Duration(expireSec))),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *Manager) GenerateToken(userID uint) (string, error) {
	return m.generateToken(userID, AccessTokenType, m.cfg.AccessExpireSec)
}

func (m *Manager) GenerateTokenPair(userID uint) (*TokenPair, error) {
	accessToken, err := m.generateToken(userID, AccessTokenType, m.cfg.AccessExpireSec)
	if err != nil {
		return nil, err
	}

	refreshToken, err := m.generateToken(userID, RefreshTokenType, m.cfg.RefreshExpireSec)
	if err != nil {
		return nil, err
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (m *Manager) ParseToken(tokenString string, expected ...TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(*jwt.Token) (interface{}, error) {
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if len(expected) > 0 && claims.Type != expected[0] {
		return nil, fmt.Errorf("%w: expected %s token", ErrInvalidTokenType, expected[0])
	}

	return claims, nil
}

func (m *Manager) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := m.ParseToken(refreshToken, RefreshTokenType)
	if err != nil {
		return nil, err
	}
	return m.GenerateTokenPair(claims.UserID)
}
