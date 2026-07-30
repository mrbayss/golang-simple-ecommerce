package jwt

import (
	"github.com/gofiber/fiber/v3"
	"github.com/mrbayss/golang-simple-ecommerce/internal/model"
	"github.com/sirupsen/logrus"
)

type MiddlewareConfig struct {
	Log  *logrus.Logger
	Jwt  *Key
	Skip map[string]bool // path -> skip auth
}

type JWTMiddleware struct {
	Log  *logrus.Logger
	Jwt  *Key
	Skip map[string]bool
}

func NewJWTMiddleware(config *MiddlewareConfig) *JWTMiddleware {
	if config.Skip == nil {
		config.Skip = map[string]bool{}
	}
	return &JWTMiddleware{
		Log:  config.Log,
		Jwt:  config.Jwt,
		Skip: config.Skip,
	}
}

func (m *JWTMiddleware) Handle() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		path := ctx.Path()
		if m.Skip[path] {
			return ctx.Next()
		}

		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			m.Log.Warn("missing authorization header")
			return ctx.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse("Token tidak ditemukan", nil))
		}

		if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			m.Log.Warn("invalid authorization format")
			return ctx.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse("Format token tidak valid", nil))
		}

		tokenString := authHeader[7:]
		claims, err := m.Jwt.VerifyToken(tokenString)
		if err != nil {
			m.Log.Warnf("invalid token: %v", err)
			return ctx.Status(fiber.StatusUnauthorized).JSON(model.ErrorResponse("Token tidak valid atau expired", nil))
		}

		ctx.Locals("user_id", claims.ID)
		ctx.Locals("user_email", claims.Email)
		ctx.Locals("user_role", claims.Role)
		ctx.Locals("session_id", claims.ID)

		m.Log.Debugf("auth success: user=%s role=%s", claims.Email, claims.Role)
		return ctx.Next()
	}
}

func GetUserID(ctx fiber.Ctx) (string, bool) {
	v, ok := ctx.Locals("user_id").(string)
	return v, ok
}

func GetUserEmail(ctx fiber.Ctx) (string, bool) {
	v, ok := ctx.Locals("user_email").(string)
	return v, ok
}

func GetUserRole(ctx fiber.Ctx) (string, bool) {
	v, ok := ctx.Locals("user_role").(string)
	return v, ok
}

func GetSessionID(ctx fiber.Ctx) (string, bool) {
	v, ok := ctx.Locals("session_id").(string)
	return v, ok
}
