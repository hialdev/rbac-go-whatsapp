package handlers

import (
	"al/models"
	"al/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type DangerHandler struct {
	DB *gorm.DB
}

func (h *DangerHandler) CleanUpDatabase(c *fiber.Ctx) error {
	if err := h.DB.Exec("DELETE FROM role_permissions").Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal menghapus data role_permissions", err.Error())
	}

	// Hapus semua data model utama, AllowGlobalUpdate supaya bisa delete tanpa where
	if err := h.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Role{}).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal menghapus data roles", err.Error())
	}

	if err := h.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Permission{}).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal menghapus data permissions", err.Error())
	}

	if err := h.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.User{}).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal menghapus data users", err.Error())
	}

	return utils.RespApi(c, "ok", "Berhasil membersihkan seluruh data di database", nil)
}
