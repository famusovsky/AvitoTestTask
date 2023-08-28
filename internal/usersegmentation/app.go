package usersegmentation

import (
	"database/sql"

	"github.com/famusovsky/AvitoTestTask/internal/usersegmentation/models"
	"github.com/famusovsky/AvitoTestTask/internal/usersegmentation/postgres"
	"github.com/gofiber/fiber/v2"
)

type App struct {
	webApp      *fiber.App
	dbProcessor models.UserSegmentationDbProcessor
}

func CreateApp(db *sql.DB, getProcessor func(*sql.DB) (models.UserSegmentationDbProcessor, error)) (App, error) {
	// TODO
	application := fiber.New()
	dbProcessor, err := postgres.GetModel(db)
	if err != nil {
		return App{}, err
	}

	return App{
		webApp:      application,
		dbProcessor: dbProcessor,
	}, nil
}

func (app App) Run() {
	// TODO
	return
}
