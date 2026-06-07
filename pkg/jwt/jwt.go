package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// --- 配置结构体 ---

// Config 包含 JWT 生成和验证所需的所有配置
type Config struct {
	Secret           string
	AccessExpireSec  int64 // Access Token 有效期（秒）
	RefreshExpireSec int64 // Refresh Token 有效期（秒）
}

// --- 核心管理器 ---

// Manager 负责处理 JWT 的创建、解析和刷新
type Manager struct {
	cfg Config
}

func NewManager(cfg Config) (*Manager, error) {
	if cfg.Secret == "" {
		return nil, errors.New("JWT secret cannot be empty")
	}
	if cfg.AccessExpireSec <= 0 {
		cfg.AccessExpireSec = 7200 // 默认 2 小时
	}
	if cfg.RefreshExpireSec <= 0 {
		cfg.RefreshExpireSec = 604800 // 默认 7 天
	}

	return &Manager{cfg: cfg}, nil
}

// --- Token 类型和 Claims ---

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

// generateToken 内部统一生成逻辑
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

// GenerateToken 生成单个 Access Token
func (m *Manager) GenerateToken(userID uint) (string, error) {
	return m.generateToken(userID, AccessTokenType, m.cfg.AccessExpireSec)
}

// GenerateTokenPair 生成 access_token 和 refresh_token
func (m *Manager) GenerateTokenPair(userID uint) (*TokenPair, error) {
	// 生成 access token
	accessToken, err := m.generateToken(userID, AccessTokenType, m.cfg.AccessExpireSec)
	if err != nil {
		return nil, err
	}

	// 生成 refresh token
	refreshToken, err := m.generateToken(userID, RefreshTokenType, m.cfg.RefreshExpireSec)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ParseToken 解析并验证 Token
func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	if len(m.cfg.Secret) == 0 {
		return nil, errors.New("JWT secret not configured")
	}

	// 解析 token
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

// RefreshToken 使用 Refresh Token 换取新的 Token 对
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
