package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	Secret           string
	AccessExpireSec  int64
	RefreshExpireSec int64
}

type Manager struct {
	cfg Config
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

	return &Manager{cfg: cfg}, nil
}

type TokenType string

const (
	AccessTokenType  TokenType = "access"
	RefreshTokenType TokenType = "refresh"
)

type Claims struct {
	UserID uint      `json:"user_id"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (m *Manager) generateToken(userID uint, tokenType TokenType, expireSeconds int64) (string, error) {
	if len(m.cfg.Secret) == 0 {
		return "", errors.New("JWT secret not configured")
	}

	claims := Claims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireSeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.cfg.Secret))
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

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	if len(m.cfg.Secret) == 0 {
		return nil, errors.New("JWT secret not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (m *Manager) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := m.ParseToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if claims.Type != RefreshTokenType {
		return nil, errors.New("invalid token type: expected refresh token")
	}

	return m.GenerateTokenPair(claims.UserID)
}
