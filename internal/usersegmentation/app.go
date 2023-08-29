// usersegmentation - пакет, реализующий сегментацию пользователей.
package usersegmentation

import (
	"log"

	"github.com/famusovsky/AvitoTestTask/internal/usersegmentation/models"

	"github.com/gofiber/fiber/v2"
)

// App - структура, описывающая приложение.
type App struct {
	webApp      *fiber.App                         // webApp - веб-приложение на основе фреймворка Fiber.
	dbProcessor models.UserSegmentationDbProcessor // dbProcessor - обработчик БД.
	errorLog    *log.Logger                        // errorLog - логгер ошибок.
}

// CreateApp - создание приложения.
//
// Принимает: логгер, обработчик БД.
//
// Возвращает: приложение.
func CreateApp(logger *log.Logger, dbProcessor models.UserSegmentationDbProcessor) *App {
	// TODO implement
	application := fiber.New()
	// application.Use(middleware.Recover)

	result := &App{
		webApp:      application,
		dbProcessor: dbProcessor,
		errorLog:    logger,
	}

	result.webApp.Post("/segments", result.PostSegment)
	result.webApp.Delete("/segments", result.DeleteSegment)
	result.webApp.Patch("/users", result.ModifyUser)
	result.webApp.Get("/users", result.GetUserRelations)

	return result
}

// Run - запуск приложения.
func (app *App) Run() {
	// TODO implement
	return
}
