package auth

import (
	"log/slog"
	"os"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func GetSecretKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "fallbacksecret"
	}
	return []byte(secret)
}

// Login credentials for Swagger documentation
type Login struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// DummyLogin is a dummy function to generate Swagger docs for the gin-jwt LoginHandler
// @Summary Login
// @Description Authenticate and get JWT token (admin/admin or user/user)
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body Login true "Login credentials"
// @Success 200 {object} map[string]string "token response"
// @Failure 401 {object} map[string]string "unauthorized"
// @Router /api/auth/login [post]
func DummyLogin() {}

// SetupJWTMiddleware creates and configures the gin-jwt middleware
func SetupJWTMiddleware() (*jwt.GinJWTMiddleware, error) {
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "qisur-service",
		Key:         GetSecretKey(),
		Timeout:     time.Hour * 24,
		MaxRefresh:  time.Hour * 24,
		IdentityKey: "id",
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(map[string]interface{}); ok {
				return jwt.MapClaims{
					"id":   v["id"],
					"role": v["role"],
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return map[string]interface{}{
				"id":   claims["id"],
				"role": claims["role"],
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals LoginRequest
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			role := "client"
			if loginVals.Username == "admin" {
				role = "admin"
			}

			return map[string]interface{}{
				"id":   "user-uuid-1234",
				"role": role,
			}, nil
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			return true // general validation, roles checked in RequireRole
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"status":  "error",
				"message": message,
			})
		},
		TokenLookup:   "header: Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})

	if err != nil {
		slog.Error("JWT Error", "error", err)
		return nil, err
	}

	if errInit := authMiddleware.MiddlewareInit(); errInit != nil {
		return nil, errInit
	}

	return authMiddleware, nil
}

// RequireRole checks the claims injected by gin-jwt
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		role, ok := claims["role"]
		if !ok || role != requiredRole {
			c.JSON(403, gin.H{"status": "error", "message": "Forbidden: insufficient permissions"})
			c.Abort()
			return
		}
		c.Next()
	}
}
