package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	apiPkg "readbud/internal/api"
	"readbud/internal/pkg/crypto"
)

// Context keys for user information.
const (
	CtxKeyUserID   = "user_id"
	CtxKeyUsername = "username"
	CtxKeyRole     = "role"
	CtxKeyClaims   = "claims"
)

// JWTAuth creates a Gin middleware that validates JWT tokens.
func JWTAuth(jwtConfig crypto.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			apiPkg.Unauthorized(c, "缺少认证令牌")
			c.Abort()
			return
		}

		// Extract "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			apiPkg.Unauthorized(c, "认证令牌格式错误")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims, err := crypto.ParseToken(jwtConfig, tokenStr)
		if err != nil {
			apiPkg.Error(c, http.StatusUnauthorized, 401, "认证令牌无效或已过期")
			c.Abort()
			return
		}

		// Inject user info into context
		c.Set(CtxKeyUserID, claims.UserID)
		c.Set(CtxKeyUsername, claims.Username)
		c.Set(CtxKeyRole, claims.Role)
		c.Set(CtxKeyClaims, claims)

		c.Next()
	}
}

// GetUserID extracts the user ID from the Gin context.
func GetUserID(c *gin.Context) (int64, bool) {
	val, exists := c.Get(CtxKeyUserID)
	if !exists {
		return 0, false
	}
	id, ok := val.(int64)
	return id, ok
}

// GetClaims extracts the JWT claims from the Gin context.
func GetClaims(c *gin.Context) (*crypto.Claims, bool) {
	val, exists := c.Get(CtxKeyClaims)
	if !exists {
		return nil, false
	}
	claims, ok := val.(*crypto.Claims)
	return claims, ok
}
