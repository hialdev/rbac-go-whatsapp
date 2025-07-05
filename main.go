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
        AllowOrigins:     "http://localhost:4321, http://localhost:5173, http://localhost:8081, http://localhost:19000, http://localhost:19001, http://localhost:19002",
        AllowMethods:     "GET,POST,DELETE",
        AllowHeaders:     "Origin,Content-Type,Authorization",
        AllowCredentials: true,
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
