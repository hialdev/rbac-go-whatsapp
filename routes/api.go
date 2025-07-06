package routes

import (
	"al/connection"
	"al/handlers"
	"al/middlewares"
	"al/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB) {
	api := app.Group("/api")

	wa := api.Group("/wa")
	wa.Get("/ws", websocket.New(handlers.WAHandler))
	wa.Post("/send", handlers.SendMessageHandler)
	wa.Post("/check", handlers.CheckNumberHandler)
	wa.Get("/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "WhatsApp API is running",
			"connected": connection.IsConnected(),
			"user_id": connection.GetUserID(),
		})
	})

	otpHandler := handlers.OtpHandler{DB:db}
	otp := api.Group("/otp")
	otp.Post("/request", otpHandler.SendOTP)
	otp.Post("/validate", otpHandler.ValidateOTP)

	auth := handlers.NewAuthHandler(db)
	api.Post("/checkuser", auth.CheckRegistered)
	api.Post("/auth/register", auth.Register)
	api.Post("/auth/login", auth.Login)
	api.Post("/auth/refresh", auth.RefreshToken)

	protected := api.Group("/auth")
	protected.Use(middlewares.JWTProtected())
	protected.Post("/checktoken", auth.CheckAccessToken)
	protected.Post("/logout", auth.Logout)

	danger := handlers.DangerHandler{DB: db}
	api.Delete("/db/cleanup", danger.CleanUpDatabase)

	permissions := handlers.NewHandlerGeneric[models.Permission](db)
	pm := api.Group("/permissions")
	pm.Use(middlewares.JWTProtected())
	pm.Get("/",middlewares.DoACL("list_permission"), permissions.GetAll)
	pm.Get("/:id",middlewares.DoACL("find_permission"), permissions.GetById)
	pm.Post("/",middlewares.DoACL("add_permission"), permissions.Create)
	pm.Post("/:id",middlewares.DoACL("update_permission"), permissions.Update)
	pm.Delete("/:id",middlewares.DoACL("delete_permission"), permissions.Delete)

	userHandler := handlers.NewUserHandler(db)
	usr := api.Group("/users")
	usr.Use(middlewares.JWTProtected())
	usr.Get("/",middlewares.DoACL("list_user"), userHandler.GetUsers)
	usr.Post("/",middlewares.DoACL("add_user"), userHandler.Create)
	usr.Get("/:id",middlewares.DoACL("find_user"), userHandler.GetUser)
	usr.Post("/:id",middlewares.DoACL("update_user"), userHandler.Update)
	usr.Post("/:id/assign",middlewares.DoACL("update_user"), userHandler.AssignRole)
	usr.Delete("/:id",middlewares.DoACL("delete_user"), userHandler.Delete)

	roles := handlers.NewRoleHandler(db)
	rl := api.Group("/roles")
	rl.Use(middlewares.JWTProtected())
	rl.Get("/",middlewares.DoACL("list_role"), roles.GetRoles)
	rl.Get("/:id",middlewares.DoACL("find_role"), roles.GetRole)
	rl.Post("/",middlewares.DoACL("add_role"), roles.CreateRole)
	rl.Post("/:id",middlewares.DoACL("update_role"), roles.UpdateRole)
	rl.Delete("/:id",middlewares.DoACL("delete_role"), roles.DeleteRole)

	settings := handlers.NewSettingHandler(db)
	setting := api.Group("/settings")
	setting.Use(middlewares.JWTProtected())
	setting.Get("/",middlewares.DoACL("list_setting"), settings.GetSettings)
	setting.Post("/",middlewares.DoACL("add_setting"), settings.AddSetting)
	setting.Get("/:id",middlewares.DoACL("find_setting"), settings.GetSetting)
	setting.Post("/:id/value",middlewares.DoACL("value_setting"), settings.ValueSetting)
	setting.Post("/:id",middlewares.DoACL("update_setting"), settings.UpdateSetting)
	setting.Delete("/:id",middlewares.DoACL("delete_setting"), settings.DeleteSetting)

	group := handlers.NewHandlerGeneric[models.TodoGroup](db)
	tg := api.Group("/group")
	tg.Use(middlewares.JWTProtected())
	tg.Get("/", group.GetAll)
	tg.Get("/:id", group.GetById)
	tg.Post("/", group.Create)
	tg.Post("/:id", group.Update)
	tg.Delete("/:id", group.Delete)

	join := handlers.NewHandlerGeneric[models.TodoGroupMember](db)
	jg := api.Group("/join")
	jg.Use(middlewares.JWTProtected())
	jg.Get("/", join.GetAll)
	jg.Get("/:id", join.GetById)
	jg.Post("/", join.Create)
	jg.Post("/:id", join.Update)
	jg.Delete("/:id", join.Delete)

	task := handlers.NewHandlerGeneric[models.Task](db)
	tsk := api.Group("/task")
	tsk.Use(middlewares.JWTProtected())
	tsk.Get("/", task.GetAll)
	tsk.Get("/:id", task.GetById)
	tsk.Post("/", task.Create)
	tsk.Post("/:id", task.Update)
	tsk.Delete("/:id", task.Delete)

	discussion := handlers.NewHandlerGeneric[models.TaskDiscussion](db)
	dsc := api.Group("/discussion")
	dsc.Use(middlewares.JWTProtected())
	dsc.Get("/", discussion.GetAll)
	dsc.Get("/:id", discussion.GetById)
	dsc.Post("/", discussion.Create)
	dsc.Post("/:id", discussion.Update)
	dsc.Delete("/:id", discussion.Delete)
}