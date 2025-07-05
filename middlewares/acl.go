package middlewares

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"al/utils"
)

func DoACL(requiredPerms ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// =====  DEBUG SECTION =====
		fmt.Println("\n=== DEBUG ACL MIDDLEWARE ===")
		fmt.Printf("Route: %s %s\n", c.Method(), c.Path())
		fmt.Printf("Required permissions: %v\n", requiredPerms)
		
		// Cek semua locals
		fmt.Println("All locals keys:")
		if c.Locals("user_id") != nil {
			fmt.Printf("  - user_id: %v\n", c.Locals("user_id"))
		}
		if c.Locals("user") != nil {
			fmt.Printf("  - user: present\n")
		}
		
		// Cek permissions secara detail
		permsRaw := c.Locals("permissions")
		fmt.Printf("Permissions raw: %+v (type: %T)\n", permsRaw, permsRaw)
		
		if permsRaw == nil {
			fmt.Println("❌ Permissions is nil!")
			return utils.RespApi(c, "perm", "Tidak ada permission di token", nil)
		}
		// ===========================

		perms, ok := c.Locals("permissions").([]any)
		if !ok {
			fmt.Printf("❌ Type assertion failed! Expected []any, got %T\n", c.Locals("permissions"))
			
			// Coba type assertion lain yang mungkin
			if permsSlice, ok2 := permsRaw.([]interface{}); ok2 {
				fmt.Printf("✅ Found as []interface{}: %+v\n", permsSlice)
				perms = permsSlice
			} else if permsString, ok3 := permsRaw.([]string); ok3 {
				fmt.Printf("✅ Found as []string: %+v\n", permsString)
				// Convert []string to []any
				perms = make([]any, len(permsString))
				for i, v := range permsString {
					perms[i] = v
				}
			} else {
				return utils.RespApi(c, "perm", "Tidak ada permission di token", nil)
			}
		}

		fmt.Printf("Permissions parsed: %+v\n", perms)

		permMap := make(map[string]bool)
		for _, p := range perms {
			if str, ok := p.(string); ok {
				permMap[strings.ToLower(str)] = true
				fmt.Printf("✅ Added permission: %s\n", str)
			} else {
				fmt.Printf("⚠️ Non-string permission: %+v (type: %T)\n", p, p)
			}
		}

		fmt.Printf("Final permission map: %+v\n", permMap)

		// Check required permissions
		for _, rp := range requiredPerms {
			rpLower := strings.ToLower(rp)
			if !permMap[rpLower] {
				fmt.Printf("❌ Missing permission: %s\n", rp)
				return utils.RespApi(c, "perm", "Permission tidak mencukupi: "+rp, nil)
			}
			fmt.Printf("✅ Permission OK: %s\n", rp)
		}

		fmt.Println("=== ACL CHECK PASSED ===\n")
		return c.Next()
	}
}