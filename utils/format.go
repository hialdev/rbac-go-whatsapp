package utils

import (
	"unicode"

	"github.com/gofiber/fiber/v2"
)

func RespApi(c *fiber.Ctx, respType string, message string, data interface{}) error {
	type respCondition struct {
		status   bool
		respcode int
		code     int
		message  string
	}
	condition := map[string]respCondition{
		"ok": {
			respcode: fiber.StatusOK,
			status:   true,
			code:     fiber.StatusOK,
			message:  "Berhasil! ",
		},
		"bad": {
			respcode: fiber.StatusBadRequest,
			status:   false,
			code:     fiber.StatusBadRequest,
			message:  "Kesalahan Permintaan! ",
		},
		"ise": {
			respcode: fiber.StatusInternalServerError,
			status:   false,
			code:     fiber.StatusInternalServerError,
			message:  "Ooops.. Terdapat Kesalahan! ",
		},
		"add": {
			status:  true,
			code:    fiber.StatusCreated,
			message: "Berhasil dibuat! ",
		},
		"perm": {
			respcode: fiber.StatusUnauthorized,
			status:   false,
			code:     fiber.StatusUnauthorized,
			message:  "Perizinan Error! ",
		},
		"empty": {
			respcode: fiber.StatusOK,
			status:   false,
			code:     fiber.StatusNotFound,
			message:  "Tidak Ada Data Ditemukan! ",
		},
	}

	resp, ok := condition[respType]
	if !ok {
		resp = respCondition{status: false, code: fiber.StatusInternalServerError}
	}

	return c.Status(resp.respcode).JSON(fiber.Map{
		"success": resp.status,
		"message": resp.message + message,
		"code":    resp.code,
		"data":    data,
	})
}

func CheckPasswordCriteria(password string) (bool) {
	if len(password) < 8 || len(password) > 16 {
		return false
	}

	var hasUpper, hasLower, hasNumber, hasSymbol bool

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasNumber = true
		case ch == '#' || ch == '@' || ch == '!' || ch == '&':
			hasSymbol = true
		}
	}

	if hasUpper && hasLower && hasNumber && hasSymbol {
		return true
	}
	return false
}
