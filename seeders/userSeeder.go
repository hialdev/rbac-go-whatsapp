package seeders

import (
	"al/models" // sesuaikan dengan path project Anda
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Seeder struct {
	DB *gorm.DB
}

func NewSeeder(db *gorm.DB) *Seeder {
	return &Seeder{DB: db}
}

// RunAll menjalankan semua seeder
func (s *Seeder) RunAll() error {
	if err := s.SeedPermissions(); err != nil {
		return fmt.Errorf("failed to seed permissions: %w", err)
	}

	if err := s.SeedRoles(); err != nil {
		return fmt.Errorf("failed to seed roles: %w", err)
	}

	if err := s.SeedUsers(); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	log.Println("All seeders completed successfully!")
	return nil
}

// SeedPermissions membuat semua permissions berdasarkan route yang ada
func (s *Seeder) SeedPermissions() error {
	permissions := []models.Permission{
		// Permission untuk Users
		{Name: "list_user", Description: stringPtr("Can list all users")},
		{Name: "add_user", Description: stringPtr("Can add new user")},
		{Name: "find_user", Description: stringPtr("Can find specific user")},
		{Name: "update_user", Description: stringPtr("Can update user")},
		{Name: "delete_user", Description: stringPtr("Can delete user")},

		// Permission untuk Roles
		{Name: "list_role", Description: stringPtr("Can list all roles")},
		{Name: "find_role", Description: stringPtr("Can find specific role")},
		{Name: "add_role", Description: stringPtr("Can add new role")},
		{Name: "update_role", Description: stringPtr("Can update role")},
		{Name: "delete_role", Description: stringPtr("Can delete role")},

		// Permission untuk Permissions
		{Name: "list_permission", Description: stringPtr("Can list all permissions")},
		{Name: "find_permission", Description: stringPtr("Can find specific permission")},
		{Name: "add_permission", Description: stringPtr("Can add new permission")},
		{Name: "update_permission", Description: stringPtr("Can update permission")},
		{Name: "delete_permission", Description: stringPtr("Can delete permission")},

		// Permission untuk Settings
		{Name: "list_setting", Description: stringPtr("Can list all settings")},
		{Name: "add_setting", Description: stringPtr("Can add new setting")},
		{Name: "find_setting", Description: stringPtr("Can find specific setting")},
		{Name: "value_setting", Description: stringPtr("Can update setting value")},
		{Name: "update_setting", Description: stringPtr("Can update setting")},
		{Name: "delete_setting", Description: stringPtr("Can delete setting")},
	}

	for _, permission := range permissions {
		// Check if permission already exists
		var existingPermission models.Permission
		if err := s.DB.Where("name = ?", permission.Name).First(&existingPermission).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create permission if not exists
				if err := s.DB.Create(&permission).Error; err != nil {
					return fmt.Errorf("failed to create permission %s: %w", permission.Name, err)
				}
				log.Printf("Created permission: %s", permission.Name)
			} else {
				return fmt.Errorf("error checking permission %s: %w", permission.Name, err)
			}
		} else {
			log.Printf("Permission %s already exists", permission.Name)
		}
	}

	return nil
}

// SeedRoles membuat roles developer dan content dengan permissions yang sesuai
func (s *Seeder) SeedRoles() error {
	// Get all permissions
	var allPermissions []models.Permission
	if err := s.DB.Find(&allPermissions).Error; err != nil {
		return fmt.Errorf("failed to get permissions: %w", err)
	}

	// Create Developer Role (has all permissions)
	developerRole := models.Role{
		Name:        "developer",
		Description: stringPtr("Developer role with full access"),
		Permissions: allPermissions, // All permissions
	}

	// Check if developer role exists
	var existingDevRole models.Role
	if err := s.DB.Where("name = ?", "developer").First(&existingDevRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := s.DB.Session(&gorm.Session{FullSaveAssociations: true}).Omit("Permissions.*").Create(&developerRole).Error; err != nil {
				return fmt.Errorf("failed to create developer role: %w", err)
			}
			log.Println("Created developer role with all permissions")
		} else {
			return fmt.Errorf("error checking developer role: %w", err)
		}
	} else {
		log.Println("Developer role already exists")
	}

	// Create Content Role (limited permissions)
	var contentPermissions []models.Permission
	excludedPermissions := []string{"update_permission", "delete_permission", "update_setting", "delete_setting"}

	for _, permission := range allPermissions {
		// Include all permissions except update_permission and delete_permission
		excluded := false
		for _, excludedPerm := range excludedPermissions {
			if permission.Name == excludedPerm {
				excluded = true
				break
			}
		}
		if !excluded {
			contentPermissions = append(contentPermissions, permission)
		}
	}

	contentRole := models.Role{
		Name:        "content",
		Description: stringPtr("Content role with limited access"),
		Permissions: contentPermissions,
	}

	// Check if content role exists
	var existingContentRole models.Role
	if err := s.DB.Where("name = ?", "content").First(&existingContentRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := s.DB.Session(&gorm.Session{FullSaveAssociations: true}).Omit("Permissions.*").Create(&contentRole).Error; err != nil {
				return fmt.Errorf("failed to create content role: %w", err)
			}
			log.Println("Created content role with limited permissions")
		} else {
			return fmt.Errorf("error checking content role: %w", err)
		}
	} else {
		log.Println("Content role already exists")
	}

	return nil
}

// SeedUsers membuat users dengan role yang sesuai
func (s *Seeder) SeedUsers() error {
	// Get roles
	var developerRole, contentRole models.Role

	if err := s.DB.Where("name = ?", "developer").First(&developerRole).Error; err != nil {
		return fmt.Errorf("developer role not found: %w", err)
	}

	if err := s.DB.Where("name = ?", "content").First(&contentRole).Error; err != nil {
		return fmt.Errorf("content role not found: %w", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	passwordStr := string(hashedPassword)

	users := []models.User{
		{
			Name:       stringPtr("Developer User"),
			Username:   stringPtr("developer"),
			Password:   &passwordStr,
			Phone:      "6281234567890",
			VerifiedAt: true,
			RoleID:     &developerRole.ID,
		},
		{
			Name:       stringPtr("Content Manager"),
			Username:   stringPtr("content"),
			Password:   &passwordStr,
			Phone:      "6281234567891",
			VerifiedAt: true,
			RoleID:     &contentRole.ID,
		},
	}

	for _, user := range users {
		// Check if user already exists
		var existingUser models.User
		if err := s.DB.Where("username = ?", *user.Username).First(&existingUser).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := s.DB.Create(&user).Error; err != nil {
					return fmt.Errorf("failed to create user %s: %w", *user.Username, err)
				}
				log.Printf("Created user: %s", *user.Username)
			} else {
				return fmt.Errorf("error checking user %s: %w", *user.Username, err)
			}
		} else {
			log.Printf("User %s already exists", *user.Username)
		}
	}

	return nil
}

// Helper function untuk membuat pointer string
func stringPtr(s string) *string {
	return &s
}

// CleanAll menghapus semua data seeder (untuk testing)
func (s *Seeder) CleanAll() error {
	// Hapus dalam urutan yang benar untuk menghindari constraint errors
	if err := s.DB.Exec("DELETE FROM role_permissions").Error; err != nil {
		return fmt.Errorf("failed to clean role_permissions: %w", err)
	}

	if err := s.DB.Delete(&models.User{}, "username IN (?)", []string{"developer", "content"}).Error; err != nil {
		return fmt.Errorf("failed to clean users: %w", err)
	}

	if err := s.DB.Delete(&models.Role{}, "name IN (?)", []string{"developer", "content"}).Error; err != nil {
		return fmt.Errorf("failed to clean roles: %w", err)
	}

	if err := s.DB.Delete(&models.Permission{}).Error; err != nil {
		return fmt.Errorf("failed to clean permissions: %w", err)
	}

	log.Println("All seeder data cleaned successfully!")
	return nil
}
