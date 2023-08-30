// usersegmentation - пакет, реализующий сегментацию пользователей.
package usersegmentation

import (
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/famusovsky/AvitoTestTask/internal/usersegmentation/models"

	"github.com/gofiber/fiber/v2"
	fiberLog "github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/swagger"
)

// App - структура, описывающая приложение.
type App struct {
	webApp      *fiber.App                         // webApp - веб-приложение на основе фреймворка Fiber.
	dbProcessor models.UserSegmentationDbProcessor // dbProcessor - обработчик БД.
	logger      *log.Logger                        // errorLog - логгер ошибок.
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
		logger:      logger, // XXX not used now
	}

	result.webApp.Post("/segments", result.PostSegment)
	result.webApp.Delete("/segments", result.DeleteSegment)
	result.webApp.Patch("/users", result.ModifyUser)
	result.webApp.Get("/users/:id", result.GetUserRelations)

	return result
}

// Run - запуск приложения.
//
// Принимает: адрес.
func (app *App) Run(addr string) {
	app.webApp.Get("/swagger/*", swagger.New()) // default

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := app.webApp.Shutdown(); err != nil {
			app.logger.Printf("Error while shutting down the server: %v", err)
		}

		close(idleConnsClosed)
	}()

	app.logger.Fatalln(app.webApp.Listen(addr))
}
