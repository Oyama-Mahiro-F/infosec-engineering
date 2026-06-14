// Package middleware HTTP中间件
// 提供JWT认证、RBAC鉴权、CORS等中间件
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/fansicheng/paddle-trace/config"
)

// ============================================================================
// JWT Claims
// ============================================================================

// JWTClaims JWT Token载荷
type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// ============================================================================
// AuthMiddleware JWT认证中间件
// ============================================================================

// AuthMiddleware JWT认证中间件
// 从Authorization Header提取并验证JWT Token
// 验证通过后将用户信息注入Gin Context
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 提取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":  401,
				"error": "authorization header required",
			})
			c.Abort()
			return
		}

		// 解析Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":  401,
				"error": "invalid authorization format, expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenStr := parts[1]

		// 验证JWT
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":  401,
				"error": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// 注入用户信息到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// ============================================================================
// RBACMiddleware 基于角色的访问控制中间件
// ============================================================================

// RequireRole RBAC鉴权中间件工厂函数
// 接收一个或多个允许的角色，检查当前用户是否具备所需角色
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code":  403,
				"error": "authentication required before authorization",
			})
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":  500,
				"error": "invalid role type in context",
			})
			c.Abort()
			return
		}

		// 检查角色是否在允许列表中
		allowed := false
		for _, r := range roles {
			if r == roleStr {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"error":   "insufficient permissions",
				"detail":  "required one of: " + strings.Join(roles, ", "),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ============================================================================
// CORSMiddleware 跨域中间件
// ============================================================================

// CORSMiddleware 处理跨域请求
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// ============================================================================
// RateLimitMiddleware 简易速率限制中间件
// ============================================================================

// RateLimitMiddleware 基于IP的简易速率限制
// 防止恶意高频查询耗尽API资源（DDoS防护）
func RateLimitMiddleware() gin.HandlerFunc {
	// 生产环境应集成Redis实现分布式速率限制
	// 原型阶段使用内存计数器
	rateLimit := make(map[string]int)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// 简化实现：每IP每分钟最多60次请求
		count := rateLimit[clientIP]
		if count > 60 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":  429,
				"error": "rate limit exceeded, please try again later",
			})
			c.Abort()
			return
		}

		rateLimit[clientIP] = count + 1
		c.Next()
	}
}
