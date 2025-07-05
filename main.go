package main

import (
	"al/connection"
	"al/models"
	"al/routes"
	"al/utils"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	time.LoadLocation(os.Getenv("APP_TIMEZONE"))
	
	utils.ValidationTranslationInit()

	app := fiber.New()
	app.Use(cors.New(cors.Config{
        AllowOrigins:     "*",
        AllowMethods:     "GET,POST,DELETE",
        AllowHeaders:     "Origin,Content-Type,Authorization",
        AllowCredentials: false,
    }))

	connection.InitDB()
	connection.InitRedis()
	connection.InitWAClient()

	//Migration
	connection.DB.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Otp{},
		&models.Setting{},
	)

	routes.SetupRoutes(app, connection.DB)
	app.Static("/uploads", "./uploads")
	app.Listen(":6789")
}
