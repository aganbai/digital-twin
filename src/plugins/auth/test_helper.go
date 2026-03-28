package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateTokenForTest 测试辅助函数：生成自定义过期时间偏移的 token
// offset 为负值表示已过期，例如 -1*time.Hour 表示 1 小时前过期
// 仅供集成测试使用
func GenerateTokenForTest(secret string, userID int64, username, role string, offset time.Duration) (string, time.Time, error) {
	now := time.Now()
	// expiresAt = now + offset（offset 为负值时表示已过期）
	expiresAt := now.Add(offset)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "digital-twin",
			IssuedAt:  jwt.NewNumericDate(now.Add(offset - 24*time.Hour)), // 签发时间也往前推
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}
