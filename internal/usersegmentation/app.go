// usersegmentation - пакет, реализующий сегментацию пользователей.
package usersegmentation

import (
	"log"
	"net/http"

	"github.com/famusovsky/AvitoTestTask/internal/usersegmentation/models"

	"github.com/gofiber/fiber/v2"
	fiberLog "github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/swagger"
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
	application := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			fiberLog.Errorf("Error: %v", err)
			return c.Status(http.StatusInternalServerError).JSON(models.Err{Text: err.Error()})
		},
	})

	result := &App{
		webApp:      application,
		dbProcessor: dbProcessor,
		errorLog:    logger, // XXX not used now
	}

	result.webApp.Post("/segments", result.PostSegment)
	result.webApp.Delete("/segments", result.DeleteSegment)
	result.webApp.Patch("/users", result.ModifyUser)
	result.webApp.Get("/users", result.GetUserRelations)

	return result
}

// Run - запуск приложения.
//
// Принимает: адрес.
func (app *App) Run(addr string) {
	app.webApp.Get("/swagger/*", swagger.New()) // default
	// app.webApp.Use(middleware.Recover)

	app.errorLog.Fatalln(app.webApp.Listen(addr))
}
