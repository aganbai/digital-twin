package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 声明
type Claims struct {
	UserID    int64  `json:"user_id"`
	PersonaID int64  `json:"persona_id"` // V2.0 迭代2：分身ID
	Username  string `json:"username"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager JWT 管理器
type JWTManager struct {
	secret []byte
	issuer string
	expiry time.Duration
}

// NewJWTManager 创建 JWT 管理器
func NewJWTManager(secret, issuer string, expiry time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		issuer: issuer,
		expiry: expiry,
	}
}

// GenerateToken 生成 JWT token
// personaID 为可变参数，保持向后兼容
func (m *JWTManager) GenerateToken(userID int64, username, role string, personaID ...int64) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.expiry)

	pid := int64(0)
	if len(personaID) > 0 {
		pid = personaID[0]
	}

	claims := &Claims{
		UserID:    userID,
		PersonaID: pid,
		Username:  username,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("生成 token 失败: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ValidateToken 验证并解析 JWT token
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名算法: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token 验证失败: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("无效的 token")
	}

	return claims, nil
}

// ParseTokenIgnoreExpiry 解析 token，忽略过期错误
// 返回 claims 和是否过期
func (m *JWTManager) ParseTokenIgnoreExpiry(tokenString string) (*Claims, bool, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名算法: %v", token.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithoutClaimsValidation())
	if err != nil {
		return nil, false, fmt.Errorf("token 解析失败: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, false, fmt.Errorf("无效的 token claims")
	}

	// 检查是否过期
	isExpired := false
	if claims.ExpiresAt != nil {
		isExpired = claims.ExpiresAt.Time.Before(time.Now())
	}

	return claims, isExpired, nil
}

// RefreshGracePeriod 刷新宽限期（7天）
const RefreshGracePeriod = 7 * 24 * time.Hour
