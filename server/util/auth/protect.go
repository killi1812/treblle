package auth

import (
	"net/http"
	"slices"
	"treblle/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Protect protects routes allowing access only to given roles (model.UserRole)
// if roles are empty they it only checks for the validity of tokens
func Protect(roles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Missing token")
			return
		}

		token, claims, err := ParseToken(authHeader)
		if err != nil {
			zap.S().Infof("Auth failed with err = %+v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Invalid token format")
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Invalid token")
			return
		}

		if len(roles) != 0 && !slices.Contains(roles, claims.Role) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}
