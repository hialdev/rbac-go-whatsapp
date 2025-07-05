package middlewares

import (
	"os"
	"strings"
	"al/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func JWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return utils.RespApi(c, "perm", "Token tidak ditemukan atau tidak valid", nil)
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			// Validasi signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Signing method tidak valid")
			}
			return []byte(os.Getenv("APP_SECRET")), nil
		})

		if err != nil || !token.Valid {
			return utils.RespApi(c, "perm", "Token tidak valid", nil)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return utils.RespApi(c, "perm", "Claim token tidak valid", nil)
		}

		if claims["type"] != "access" {
			return utils.RespApi(c, "perm", "Token bukan access token", nil)
		}

		// Simpan token dan user info di context
		c.Locals("user", token)
		c.Locals("user_id", claims["user_id"])

		// Simpan permissions di context jika ada
		if perms, exists := claims["permissions"]; exists {
			c.Locals("permissions", perms)
		} else {
			// Jika tidak ada permissions di token, set array kosong
			c.Locals("permissions", []interface{}{})
		}

		return c.Next()
	}
}