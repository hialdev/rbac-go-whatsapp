package handlers

import (
	"al/models"
	"al/utils"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserInitialInput struct {
	Name       string     `json:"name" validate:"required,min=2,max=20"`
	Username   string     `json:"username" validate:"required,min=4,max=12"`
	Password   string     `json:"-"`
	Phone      string     `json:"phone" validate:"required,min=4,max=14"`
	Image      *string     `json:"image"`
}

type UserUpdateInput struct {
	Name       string     `json:"name" validate:"required,min=2,max=20"`
	Username   string     `json:"username" validate:"required,min=4,max=12"`
	Phone      string     `json:"phone" validate:"required,min=4,max=14"`
	Image      *string     `json:"image"`
}

type UserHandler struct {
	DB *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{DB: db}
}

func (r *UserHandler) GetUsers(c *fiber.Ctx) error {
	var users []models.User
	if err := r.DB.Preload("Role").Find(&users).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal mendapatkan data users", err.Error())
	}

	return utils.RespApi(c, "ok", "Berhasil mendapatkan data users", users)
}

func (r *UserHandler) GetUser(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.RespApi(c, "bad", "UUID tidak valid", idStr)
	}

	var user models.User
	if err := r.DB.Preload("Role").First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.RespApi(c, "empty", "Data user tidak ditemukan", err.Error())
		}
		return utils.RespApi(c, "ise", "Kesalahan sistem dalam memproses ", err.Error())
	}
	return utils.RespApi(c, "ok", "Berhasil mendapatkan data user", user)
}

func (h *UserHandler) Create(c *fiber.Ctx) error {
	var input UserInitialInput

	// Gunakan MultipartForm untuk upload image
	if form, err := c.MultipartForm(); err == nil && form.File != nil {
		input.Name = c.FormValue("name")
		input.Username = c.FormValue("username")
		input.Phone = c.FormValue("phone")
		input.Password = c.FormValue("password")
	} else {
		if err := c.BodyParser(&input); err != nil {
			return utils.RespApi(c, "bad", "Invalid input", err.Error())
		}
	}

	if err := utils.Validate.Struct(input); err != nil {
		return utils.RespApi(c, "bad", "Validasi gagal", err.Error())
	}

	if input.Password == "" {
		return utils.RespApi(c, "bad", "Password wajib diisi", nil)
	}

	if !utils.CheckPasswordCriteria(input.Password) {
		return utils.RespApi(c, "bad", "Password tidak sesuai kriteria", nil)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.RespApi(c, "ise", "Gagal hash password", err.Error())
	}
	hashedStr := string(hashed)

	newUser := models.User{
		Name:     &input.Name,
		Username: &input.Username,
		Password: &hashedStr,
		Phone:    input.Phone,
	}

	// Upload image jika ada
	if file, err := c.FormFile("image"); err == nil && file != nil {
		filePath, err := utils.UploadFile(c, "image", "users")
		if err != nil {
			return utils.RespApi(c, "bad", "Gagal upload image User", err.Error())
		}
		newUser.Image = &filePath
	}

	if err := h.DB.Create(&newUser).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal Membuat User", err.Error())
	}

	// Ambil user terbaru untuk response (tanpa password)
	var user models.User
	if err := h.DB.First(&user, newUser.ID).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal mengambil data user", err.Error())
	}
	user.Password = nil

	return utils.RespApi(c, "ok", "Register berhasil", user)
}

func (h *UserHandler) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.RespApi(c, "bad", "ID yang diberikan tidak valid", nil)
	}

	var user models.User
	if err := h.DB.First(&user, "id = ?", id).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal Mendapatkan user", err.Error())
	}

	var input UserUpdateInput

	// Gunakan MultipartForm untuk upload image
	if form, err := c.MultipartForm(); err == nil && form.File != nil {
		input.Name = c.FormValue("name")
		input.Username = c.FormValue("username")
		input.Phone = c.FormValue("phone")
	} else {
		if err := c.BodyParser(&input); err != nil {
			return utils.RespApi(c, "bad", "Invalid input", err.Error())
		}
	}

	if err := utils.Validate.Struct(input); err != nil {
		return utils.RespApi(c, "bad", "Validasi gagal", err.Error())
	}

	updUser := models.User{
		Name:     &input.Name,
		Username: &input.Username,
		Phone:    input.Phone,
	}

	// Upload image jika ada
	oldImage := user.Image
	if file, err := c.FormFile("image"); err == nil && file != nil {
		var oldImagePath string
		if oldImage != nil {
			oldImagePath = *oldImage
		} else {
			oldImagePath = ""
		}
		filePath, err := utils.UpdateFile(c, oldImagePath, "image", "users")
		if err != nil {
			return utils.RespApi(c, "bad", "Gagal memperbarui image User", err.Error())
		}
		updUser.Image = &filePath
	}

	if err := h.DB.Model(&user).Updates(&updUser).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal Memperbarui User", err.Error())
	}

	// response tanpa password
	user.Password = nil

	var userName string
	if user.Name != nil {
		userName = *user.Name
	} else {
		userName = ""
	}

	return utils.RespApi(c, "ok", "User "+userName+" berhasil diperbarui", user)
}

func (h *UserHandler) Delete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return utils.RespApi(c, "bad", "ID yang diberikan tidak valid", nil)
	}

	var user models.User
	if err := h.DB.First(&user, "id = ?", id).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal Mendapatkan user", err.Error())
	}

	if user.Image != nil && *user.Image != "" {
		if err := utils.DeleteFile(*user.Image); err != nil{
			return utils.RespApi(c, "ise", "Gagal Menghapus Image", err.Error())
		}
	}

	if err := h.DB.Delete(&user).Error; err != nil{
		return utils.RespApi(c, "ise", "Gagal Menghapus user", err.Error())
	}
	
	var userName string
	if user.Name != nil {
		userName = *user.Name
	} else {
		userName = ""
	}

	return utils.RespApi(c, "ok", "User "+userName+" berhasil dihapus", nil)
}

func (h *UserHandler) AssignRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	var input struct {
		RoleId string `json:"role_id" validate:"required,uuid"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.RespApi(c, "bad", "Invalid input", err.Error())
	}

	userId, err := uuid.Parse(idStr)
	if err != nil {
		return utils.RespApi(c, "bad", "User ID yang diberikan tidak valid", nil)
	}

	roleId, err := uuid.Parse(input.RoleId)
	if err != nil {
		return utils.RespApi(c, "bad", "Role ID yang diberikan tidak valid", nil)
	}

	var user models.User
	if err := h.DB.First(&user, "id = ?", userId).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal Mendapatkan user", err.Error())
	}

	var role models.Role
	if err := h.DB.First(&role, "id = ?", roleId).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal Mendapatkan Role", err.Error())
	}

	if err := h.DB.Model(&user).Update("role_id", roleId).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal Assign Role ke User", err.Error())
	}

	user.Password = nil
	
	var userName string
	if user.Name != nil {
		userName = *user.Name
	} else {
		userName = ""
	}
	return utils.RespApi(c, "ok", "User "+userName+" sekarang memiliiki Role "+role.Name, user)
}