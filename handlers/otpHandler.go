package handlers

import (
	"al/connection"
	"al/models"
	"al/utils"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type OtpHandler struct {
	DB *gorm.DB
}

func (h *OtpHandler) SendOTP(c *fiber.Ctx) error {
	var input struct {
		Phone   string `json:"phone" validate:"required,min=4,max=14"`
		Purpose string `json:"purpose" validate:"oneof=register changes verify"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.RespApi(c, "bad", "Request Body tidak valid", err.Error())
	}

	if err := utils.Validate.Struct(input); err != nil {
		if verrs, ok := err.(validator.ValidationErrors); ok {
			return utils.RespApi(c, "bad", "Validasi gagal", verrs.Translate(utils.Translator))
		}
		return utils.RespApi(c, "bad", "Validasi gagal", err.Error())
	}

	code, err := GenerateUniqueOTP(h.DB)
	if err != nil {
		return utils.RespApi(c, "ise", "Tidak dapat membuat Kode OTP", err.Error())
	}

	otp := models.Otp{
		Phone:     input.Phone,
		Code:      code,
		ExpiredAt: time.Now().Add(10 * time.Minute),
		Purpose:   input.Purpose,
	}

	if err := h.DB.Create(&otp).Error; err != nil {
		return utils.RespApi(c, "ise", "Tidak dapat membuat record OTP", err.Error())
	}

	req := SendMessageRequest{
		To:      otp.Phone,
		Message: fmt.Sprintf("Ping! Pong! Kode OTP Datang! Masukan kode *%s* untuk melanjutkan.", otp.Code),
	}

	isValidNumber := false
	isValid, err := connection.CheckNumber(req.To)
	if err != nil {
		log.Printf("Failed to check number %s: %v", req.To, err)
	} else {
		isValidNumber = isValid
		if !isValid {
			return c.Status(400).JSON(SendMessageResponse{
				Success: false,
				Message: "Phone number is not registered on WhatsApp",
				Data: &SendMessageData{
					To:            req.To,
					Message:       req.Message,
					IsValidNumber: &isValidNumber,
				},
			})
		}
	}

	if otp.Purpose == "register" {
		user := models.User{
			Phone:      otp.Phone,
		}
		if err := h.DB.FirstOrCreate(&user, user).Error; err != nil {
			return utils.RespApi(c, "ise", "Gagal menyimpan Phone ke Data User", err.Error())
		}
	}

	// ✅ Tangani error dari sendMessageService
	if err := sendMessageService(c, req, isValidNumber); err != nil {
		return utils.RespApi(c, "ise", "Kesalahan dalam mengirim pesan whatsapp", err.Error())
	}

	return utils.RespApi(c, "ok", "OTP berhasil dikirim", otp)
}

func generateRandomOTP(length int) string {
	const otpCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, length)
	for i := range code {
		code[i] = otpCharset[rand.Intn(len(otpCharset))]
	}
	return string(code)
}

func GenerateUniqueOTP(db *gorm.DB) (string, error) {
	var code string
	for {
		code = generateRandomOTP(6)

		var count int64
		if err := db.Model(&models.Otp{}).Where("code = ?", code).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			break
		}
	}
	return code, nil
}

func (h *OtpHandler) ValidateOTP(c *fiber.Ctx) error {
	var input struct {
		Code string `json:"code" validate:"required,len=6"`
		Purpose string `json:"purpose" validate:"oneof=register changes verify"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.RespApi(c, "bad", "Request Body tidak valid", err.Error())
	}

	if err := utils.Validate.Struct(input); err != nil {
		if verrs, ok := err.(validator.ValidationErrors); ok {
			return utils.RespApi(c, "bad", "Validasi gagal", verrs.Translate(utils.Translator))
		}
		return utils.RespApi(c, "bad", "Validasi gagal", err.Error())
	}

	var otp models.Otp
	if err := h.DB.
		Where("LOWER(code) = ?", strings.ToLower(input.Code)).
		Where("purpose = ?", input.Purpose).
		First(&otp).Error; err != nil {
		return utils.RespApi(c, "empty", "Kode OTP tidak ditemukan / tidak cocok", err.Error())
	}

	loc, _ := time.LoadLocation(os.Getenv("APP_TIMEZONE"))
	now := time.Now().In(loc)

	if otp.ExpiredAt.Before(now) {
		return utils.RespApi(c, "perm", "Kode OTP sudah kedaluwarsa", nil)
	}

	if otp.Purpose == "register" {
		user := models.User{
			Phone: otp.Phone,
		}
		if err := h.DB.First(&user, "phone = ?", otp.Phone).Error; err != nil {
			return utils.RespApi(c, "ise", "Tidak menemukan User", err.Error())
		}
		if err := h.DB.Model(&user).Updates(models.User{VerifiedAt: true}).Error; err != nil {
			return utils.RespApi(c, "ise", "Gagal memverifikasi User", err.Error())
		}
	}

	if err := h.DB.Delete(&models.Otp{}, "phone = ?", otp.Phone).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal membersihkan OTP setelah Validasi", err.Error())
	}

	return utils.RespApi(c, "ok", "Berhasil validasi OTP", nil)
}