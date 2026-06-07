package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWT(t *testing.T) {
	// 设置测试配置
	cfg := Config{
		Secret:           "test-secret-key",
		AccessExpireSec:  3600,
		RefreshExpireSec: 7 * 86400,
	}

	m, _ := NewManager(cfg)

	t.Run("GenerateTokenPair", func(t *testing.T) {
		userID := uint(123)

		tokenPair, err := m.GenerateTokenPair(userID)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenPair.AccessToken)
		assert.NotEmpty(t, tokenPair.RefreshToken)
		assert.NotEqual(t, tokenPair.AccessToken, tokenPair.RefreshToken)
	})

	t.Run("ParseAccessToken", func(t *testing.T) {
		userID := uint(456)

		tokenPair, err := m.GenerateTokenPair(userID)
		require.NoError(t, err)

		// 解析 access token
		claims, err := m.ParseToken(tokenPair.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, AccessTokenType, claims.Type)
	})

	t.Run("ParseRefreshToken", func(t *testing.T) {
		userID := uint(789)

		tokenPair, err := m.GenerateTokenPair(userID)
		require.NoError(t, err)

		// 解析 refresh token
		claims, err := m.ParseToken(tokenPair.RefreshToken)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, RefreshTokenType, claims.Type)
	})

	t.Run("RefreshToken", func(t *testing.T) {
		userID := uint(999)

		// 生成初始 token 对
		tokenPair, err := m.GenerateTokenPair(userID)
		require.NoError(t, err)

		// 使用 refresh token 生成新的 token 对
		newTokenPair, err := m.RefreshToken(tokenPair.RefreshToken)
		require.NoError(t, err)
		assert.NotEmpty(t, newTokenPair.AccessToken)
		assert.NotEmpty(t, newTokenPair.RefreshToken)

		// 由于 JWT 的 iat (issued at) 是基于当前时间的，如果生成得太快可能相同
		// 我们主要验证功能是否正常工作
		// 解析新 token 确保它们有效
		newAccessClaims, err := m.ParseToken(newTokenPair.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, userID, newAccessClaims.UserID)
		assert.Equal(t, AccessTokenType, newAccessClaims.Type)

		newRefreshClaims, err := m.ParseToken(newTokenPair.RefreshToken)
		require.NoError(t, err)
		assert.Equal(t, userID, newRefreshClaims.UserID)
		assert.Equal(t, RefreshTokenType, newRefreshClaims.Type)

		// 确保新旧 token 的用户 ID 一致
		oldAccessClaims, err := m.ParseToken(tokenPair.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, userID, oldAccessClaims.UserID)
	})

	t.Run("RefreshTokenWithAccessToken", func(t *testing.T) {
		userID := uint(111)

		tokenPair, err := m.GenerateTokenPair(userID)
		require.NoError(t, err)

		// 尝试使用 access token 刷新（应该失败）
		_, err = m.RefreshToken(tokenPair.AccessToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token type: expected refresh token")
	})

	t.Run("TokenExpiration", func(t *testing.T) {
		m.cfg.AccessExpireSec = 1
		m.cfg.RefreshExpireSec = 1

		userID := uint(222)

		tokenPair, err := m.GenerateTokenPair(userID)
		require.NoError(t, err)

		// 等待 token 过期
		time.Sleep(2 * time.Second)

		// 尝试解析过期的 token
		_, err = m.ParseToken(tokenPair.AccessToken)
		assert.Error(t, err)

		_, err = m.ParseToken(tokenPair.RefreshToken)
		assert.Error(t, err)
	})
}
