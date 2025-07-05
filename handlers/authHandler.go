package handlers

import (
	"al/connection"
	"al/models"
	"al/utils"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

// generateTokenWithPermissions - Updated to include permissions in token
func generateTokenWithPermissions(userID string, typeToken string, expiry time.Duration, permissions []string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":     userID,
		"type":        typeToken,
		"permissions": permissions,
		"exp":         time.Now().Add(expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("APP_SECRET")))
}

func generateToken(userID string, typeToken string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"type":    typeToken,
		"exp":     time.Now().Add(expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("APP_SECRET")))
}

// getUserPermissions - Updated for single role per user
func (h *AuthHandler) getUserPermissions(userID string) ([]string, error) {
	var user models.User
	var permissions []string

	// Get user with role and permissions
	err := h.DB.Preload("Role.Permissions").First(&user, "id = ?", userID).Error
	if err != nil {
		return permissions, err
	}
	fmt.Println(user)

	// If user has no role, return empty permissions
	if user.RoleID == nil {
		return permissions, nil
	}

	// Collect all permissions from user's role
	for _, permission := range user.Role.Permissions {
		if permission.Name != "" {
			permissions = append(permissions, permission.Name)
		}
	}

	return permissions, nil
}

// REGISTER
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var input struct {
		Name     string `json:"name" validate:"required"`
		Username string `json:"username" validate:"required"`
		Phone    string `json:"phone" validate:"required"`
		Password string `json:"password" validate:"required,min=6"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.RespApi(c, "bad", "Invalid input", err.Error())
	}

	if err := utils.Validate.Struct(input); err != nil {
		return utils.RespApi(c, "bad", "Validasi gagal", err.Error())
	}

	if !utils.CheckPasswordCriteria(input.Password) {
		return utils.RespApi(c, "bad", "Password tidak sesuai kriteria", nil)
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	hashedStr := string(hashed)

	newUser := models.User{
		Name:     &input.Name,
		Username: &input.Username,
		Password: &hashedStr,
	}

	user := models.User{
		Phone: input.Phone,
	}

	if err := h.DB.First(&user, "phone = ?", input.Phone).Error; err != nil {
		return utils.RespApi(c, "ise", "Tidak menemukan User", err.Error())
	}
	if !user.VerifiedAt {
		return utils.RespApi(c, "bad", "Anda tidak dapat mendaftarkan akun tanpa validasi OTP", nil)
	}
	if err := h.DB.Model(&user).Updates(newUser).Error; err != nil {
		return utils.RespApi(c, "ise", "Gagal memverifikasi User", err.Error())
	}

	return utils.RespApi(c, "ok", "Register berhasil", user)
}

// LOGIN - Updated to include permissions in token
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var input struct {
		Login    string `json:"login" validate:"required"` // username / phone
		Password string `json:"password" validate:"required"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.RespApi(c, "bad", "Input tidak valid", err.Error())
	}

	var user models.User
	if err := h.DB.Preload("Role").Where("username = ? OR phone = ?", input.Login, input.Login).
		First(&user).Error; err != nil {
		return utils.RespApi(c, "bad", "User tidak ditemukan", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(input.Password)); err != nil {
		return utils.RespApi(c, "bad", "Password salah", nil)
	}

	// Get user permissions
	permissions, err := h.getUserPermissions(user.ID.String())
	if err != nil {
		return utils.RespApi(c, "ise", "Gagal mengambil permissions", err.Error())
	}

	// Generate tokens with permissions
	accessToken, _ := generateTokenWithPermissions(user.ID.String(), "access", time.Hour, permissions)
	refreshToken, _ := generateToken(user.ID.String(), "refresh", time.Hour*24*7)

	// Simpan refresh token di Redis/cache
	_ = connection.SetToken("refresh:"+user.ID.String(), refreshToken, time.Hour*24*7)

	// Set refresh token sebagai HttpOnly cookie
	c.Cookie(&fiber.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		MaxAge:   int((time.Hour * 24 * 7).Seconds()), // 7 hari
		HTTPOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production", // HTTPS only di production
		SameSite: "Strict",
		Path:     "/",
	})

	// Remove password from response
	user.Password = nil

	// Hanya kirim access token di response body
	return utils.RespApi(c, "ok", "Login berhasil", fiber.Map{
		"access_token": accessToken,
		"user":         user,
		"permissions":  permissions,
	})
}

func (h *AuthHandler) CheckAccessToken(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return utils.RespApi(c, "perm", "Token tidak ditemukan atau tidak valid", nil)
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
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

	return utils.RespApi(c, "ok", "Token Valid", fiber.Map{
		"user_id":     claims["user_id"],
		"permissions": claims["permissions"],
	})
}

// REFRESH TOKEN - Updated to regenerate permissions
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	// Ambil refresh token dari cookie
	refreshToken := c.Cookies("refreshToken")
	if refreshToken == "" {
		fmt.Println("DEBUG: No refresh token in cookie")
		return utils.RespApi(c, "perm", "Refresh token tidak ditemukan", nil)
	}

	fmt.Printf("DEBUG: Refresh token found: %s\n", refreshToken[:50]+"...") // Log partial token

	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv("APP_SECRET")), nil
	})
	if err != nil || !token.Valid {
		fmt.Printf("DEBUG: Invalid token: %v\n", err)
		c.ClearCookie("refreshToken")
		return utils.RespApi(c, "perm", "Token tidak valid", nil)
	}

	claims := token.Claims.(jwt.MapClaims)
	if claims["type"] != "refresh" {
		fmt.Println("DEBUG: Token is not refresh type")
		c.ClearCookie("refreshToken")
		return utils.RespApi(c, "perm", "Token bukan refresh token", nil)
	}

	userID := claims["user_id"].(string)
	fmt.Printf("DEBUG: Checking token for user: %s\n", userID)

	// Validasi dengan token yang tersimpan di Redis
	savedToken, err := connection.GetToken("refresh:" + userID)
	if err != nil {
		fmt.Printf("DEBUG: Token not found in Redis: %v\n", err)
		c.ClearCookie("refreshToken")
		return utils.RespApi(c, "perm", "Token tidak ditemukan di server", nil)
	}
	
	if savedToken != refreshToken {
		fmt.Println("DEBUG: Token mismatch with saved token")
		c.ClearCookie("refreshToken")
		return utils.RespApi(c, "perm", "Refresh token tidak cocok", nil)
	}

	fmt.Println("DEBUG: Token validation successful, generating new tokens")

	// Get fresh permissions for the user
	permissions, err := h.getUserPermissions(userID)
	if err != nil {
		return utils.RespApi(c, "ise", "Gagal mengambil permissions", err.Error())
	}

	// Generate access token baru dengan permissions terbaru
	newAccessToken, _ := generateTokenWithPermissions(userID, "access", time.Hour, permissions)

	// Generate refresh token baru dan rotate
	newRefreshToken, _ := generateToken(userID, "refresh", time.Hour*24*7)
	_ = connection.SetToken("refresh:"+userID, newRefreshToken, time.Hour*24*7)

	// Update cookie dengan refresh token baru
	c.Cookie(&fiber.Cookie{
		Name:     "refreshToken",
		Value:    newRefreshToken,
		MaxAge:   int((time.Hour * 24 * 7).Seconds()),
		HTTPOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
		SameSite: "Strict",
		Path:     "/",
	})

	return utils.RespApi(c, "ok", "Token diperbarui", fiber.Map{
		"access_token": newAccessToken,
		"user_id":      userID,
		"permissions":  permissions,
	})
}

// LOGOUT
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var userID string
	
	// Ambil user ID dari access token jika ada
	user := c.Locals("user")
	if user != nil {
		token := user.(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)
		userID = claims["user_id"].(string)
	} else {
		// Jika tidak ada access token, coba ambil dari refresh token di cookie
		refreshToken := c.Cookies("refreshToken")
		if refreshToken != "" {
			token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (any, error) {
				return []byte(os.Getenv("APP_SECRET")), nil
			})
			if err == nil && token.Valid {
				claims := token.Claims.(jwt.MapClaims)
				userID = claims["user_id"].(string)
			}
		}
	}

	// Hapus refresh token dari Redis jika ada userID
	if userID != "" {
		_ = connection.DeleteToken("refresh:" + userID)
	}

	// Clear dengan expired date yang jauh di masa lalu
	c.Cookie(&fiber.Cookie{
		Name:     "refreshToken",
		Value:    "",
		MaxAge:   -86400, // -24 jam
		Expires:  time.Now().Add(-24 * time.Hour), // Tambahan explicit expires
		HTTPOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
		SameSite: "Strict",
		Path:     "/",
	})
	
	// Clear juga dengan path alternatif
	c.Cookie(&fiber.Cookie{
		Name:     "refreshToken",
		Value:    "",
		MaxAge:   -86400,
		Expires:  time.Now().Add(-24 * time.Hour),
		HTTPOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
		SameSite: "Strict",
		Path:     "/api",
	})

	// Clear default method
	c.ClearCookie("refreshToken")

	return utils.RespApi(c, "ok", "Logout berhasil", nil)
}

// Check Registered User
func (h *AuthHandler) CheckRegistered(c *fiber.Ctx) error {
	var input struct {
		Phone string `validate:"required,min=6,max=14"`
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
	
	var user models.User
	if err := h.DB.First(&user, "phone = ?", input.Phone).Error; err != nil {
		return utils.RespApi(c, "empty", "User tidak ditemukan", input.Phone)
	}
	return utils.RespApi(c, "ok", "User Terdaftar di Database", user)
}