package usersegmentation

import (
	"github.com/famusovsky/AvitoTestTask/internal/usersegmentation/models/postgres"
	"github.com/gofiber/fiber/v2"
)

type App struct {
	webApp *fiber.App
	db     *postgres.UserSegmentation
}

func Create() App {
	// TODO
	application := fiber.New()

	return App{
		webApp: application,
	}
}

func (app App) Run() error {
	// TODO
	return nil
}
