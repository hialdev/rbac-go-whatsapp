package handlers

import (
	"al/models"
	"al/utils"

	"encoding/json"
    "strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SettingInitialInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty" validate:"omitempty"`
	SetKey      string  `json:"set_key" validate:"required"`
	SetGroupKey string  `json:"set_group_key" validate:"required"`
	SetType     string  `json:"set_type" gorm:"type:varchar(100);default:'string'" validate:"oneof=text number checkbox radio select selects file image files images richtext markdown"`
	IsUrgent    bool    `json:"is_urgent" gorm:"type:bool;default:false"`
}

type SettingHandler struct {
	DB *gorm.DB
}

func NewSettingHandler(db *gorm.DB) *SettingHandler {
	return &SettingHandler{DB: db}
}

func (h *SettingHandler) GetSettings(c *fiber.Ctx) error {
	var settings []models.Setting
	if err := h.DB.Find(&settings).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal mendapatkan data Settings", err.Error())
	}

	// grouped := make(map[string][]models.Setting)
	// for _, s := range settings {
	// 	grouped[s.SetGroupKey] = append(grouped[s.SetGroupKey], s)
	// }

	return utils.RespApi(c, "ok", "Berhasil mendapatkan data Settings", settings)
}

func (h *SettingHandler) GetSetting(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.RespApi(c, "bad", "Id yang diberikan tidak valid", nil)
	}

	var setting models.Setting
	if err := h.DB.First(&setting, "id = ?", id).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal mendapatkan data Setting", err.Error())
	}

	return utils.RespApi(c, "ok", "Berhasil mendapatkan data Setting", setting)
}

func (h *SettingHandler) AddSetting(c *fiber.Ctx) error {
	var input SettingInitialInput

	if err := c.BodyParser(&input); err != nil {
		return utils.RespApi(c, "bad", "Request Body tidak valid", err.Error())
	}

	if err := utils.Validate.Struct(input); err != nil {
		if verrs, ok := err.(validator.ValidationErrors); ok {
			return utils.RespApi(c, "bad", "Validasi gagal", verrs.Translate(utils.Translator))
		}
		return utils.RespApi(c, "bad", "Validasi gagal", err.Error())
	}

	setting := models.Setting{
		Name:        input.Name,
		Description: input.Description,
		SetKey:      input.SetKey,
		SetGroupKey: input.SetGroupKey,
		SetType:     input.SetType,
		IsUrgent:    input.IsUrgent,
	}

	if err := h.DB.Create(&setting).Error; err != nil {
		return utils.RespApi(c, "ise", "Tidak dapat membuat Setting", err.Error())
	}

	return utils.RespApi(c, "ok", "Berhasil membuat data Setting", setting)
}

func (h *SettingHandler) ValueSetting(c *fiber.Ctx) error {
	// Parse ID dari URL parameter
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.RespApi(c, "bad", "ID yang diberikan tidak valid", nil)
	}

	// Cari setting berdasarkan ID
	var setting models.Setting
	if err := h.DB.First(&setting, "id = ?", id).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal Mendapatkan Setting", err.Error())
	}

	var newValue string
	var filePaths []string

	// Handle berdasarkan SetType
	switch setting.SetType {
	case "text", "richtext", "markdown":
		var input struct {
			SetValue string `json:"set_value"`
		}
		if err := c.BodyParser(&input); err != nil {
			return utils.RespApi(c, "bad", "Request Body tidak valid", err.Error())
		}
		newValue = input.SetValue

	case "number":
		var input struct {
			SetValue string `json:"set_value"`
		}
		if err := c.BodyParser(&input); err != nil {
			return utils.RespApi(c, "bad", "Request Body tidak valid", err.Error())
		}
		// Validasi apakah value adalah number
		if input.SetValue != "" {
			if _, err := strconv.ParseFloat(input.SetValue, 64); err != nil {
				return utils.RespApi(c, "bad", "Nilai harus berupa angka", nil)
			}
		}
		newValue = input.SetValue

	case "checkbox":
		var input struct {
			SetValue bool `json:"set_value"`
		}
		if err := c.BodyParser(&input); err != nil {
			return utils.RespApi(c, "bad", "Request Body tidak valid", err.Error())
		}
		newValue = strconv.FormatBool(input.SetValue)

	case "radio", "select":
		var input struct {
			SetValue string `json:"set_value"`
		}
		if err := c.BodyParser(&input); err != nil {
			return utils.RespApi(c, "bad", "Request Body tidak valid", err.Error())
		}
		newValue = input.SetValue

	case "selects":
		var input struct {
			SetValue []string `json:"set_value"`
		}
		if err := c.BodyParser(&input); err != nil {
			return utils.RespApi(c, "bad", "Request Body tidak valid", err.Error())
		}
		// Convert array ke JSON string
		valueBytes, err := json.Marshal(input.SetValue)
		if err != nil {
			return utils.RespApi(c, "ise", "Gagal mengkonversi nilai", err.Error())
		}
		newValue = string(valueBytes)

	case "file", "image":
		// Handle single file upload - FIXED: using "set_value" as field name
		filePath, err := utils.UploadFile(c, "set_value", "settings")
		if err != nil {
			return utils.RespApi(c, "bad", "Gagal upload file", err.Error())
		}
		
		// Hapus file lama jika ada
		if setting.SetValue != nil && *setting.SetValue != "" {
			utils.DeleteFile(*setting.SetValue)
		}
		
		newValue = filePath

	case "files", "images":
		// Handle multiple files upload - FIXED: using "set_value" as field name
		uploadedPaths, err := utils.UploadFileFlex(c, "set_value", "settings")
		if err != nil {
			return utils.RespApi(c, "bad", "Gagal upload files", err.Error())
		}
		
		// Hapus file lama jika ada
		if setting.SetValue != nil && *setting.SetValue != "" {
			var oldPaths []string
			if err := json.Unmarshal([]byte(*setting.SetValue), &oldPaths); err == nil {
				for _, oldPath := range oldPaths {
					utils.DeleteFile(oldPath)
				}
			}
		}
		
		filePaths = uploadedPaths
		// Convert array ke JSON string
		pathBytes, err := json.Marshal(filePaths)
		if err != nil {
			return utils.RespApi(c, "ise", "Gagal mengkonversi path files", err.Error())
		}
		newValue = string(pathBytes)

	default:
		return utils.RespApi(c, "bad", "Tipe setting tidak didukung", nil)
	}

	// Update nilai setting di database
	if err := h.DB.Model(&setting).Update("set_value", newValue).Error; err != nil {
		// Jika gagal update dan ada file yang diupload, hapus file tersebut
		if setting.SetType == "file" || setting.SetType == "image" {
			if newValue != "" {
				utils.DeleteFile(newValue)
			}
		} else if setting.SetType == "files" || setting.SetType == "images" {
			for _, path := range filePaths {
				utils.DeleteFile(path)
			}
		}
		return utils.RespApi(c, "ise", "Gagal memberikan nilai ke Setting", err.Error())
	}

	// Refresh data setting untuk response
	if err := h.DB.First(&setting, "id = ?", id).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal mendapatkan data setting terbaru", err.Error())
	}

	return utils.RespApi(c, "ok", "Berhasil memberikan nilai data Setting", setting)
}

func (h *SettingHandler) UpdateSetting(c *fiber.Ctx) error{
	// Parse ID dari URL parameter
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.RespApi(c, "bad", "ID yang diberikan tidak valid", nil)
	}

	var input SettingInitialInput
	if err := c.BodyParser(&input); err != nil {
		return utils.RespApi(c, "bad", "Request Body tidak valid", err.Error())
	}

	// Cari setting berdasarkan ID
	var setting models.Setting
	if err := h.DB.First(&setting, "id = ?", id).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal Mendapatkan Setting", err.Error())
	}

	// Update fields
	setting.Name = input.Name
	setting.Description = input.Description
	setting.SetKey = input.SetKey
	setting.SetGroupKey = input.SetGroupKey
	setting.SetType = input.SetType
	setting.IsUrgent = input.IsUrgent

	if err := h.DB.Save(&setting).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal memperbarui data Setting", err.Error())
	}

	return utils.RespApi(c, "ok", "Berhasil memperbarui data Setting", setting)
}

func (h *SettingHandler) DeleteSetting(c *fiber.Ctx) error{
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return utils.RespApi(c, "bad", "ID yang diberikan tidak valid", nil)
	}

	var setting models.Setting
	if err := h.DB.First(&setting, "id = ?", id).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal Mendapatkan Setting", err.Error())
	}

	if !setting.IsUrgent {
		if err := h.DB.Delete(&setting).Error; err != nil{
			return utils.RespApi(c, "ise", "Gagal Menghapus Setting", err.Error())
		}
	}

	return utils.RespApi(c, "ok", "Berhasil Menghapus Setting", nil)
}