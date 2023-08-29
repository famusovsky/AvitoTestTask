package usersegmentation

import (
	"net/http"

	"github.com/famusovsky/AvitoTestTask/internal/usersegmentation/models"
	"github.com/gofiber/fiber/v2"
)

// PostSegment - добавляет сегмент в БД.
//
// Принимает: контекст.
//
// Возвращает: ошибку.
func (app *App) PostSegment(c *fiber.Ctx) error {
	if ok, err := checkType(c); !ok {
		return err
	}
	slug, ok, err := getSlug(c)
	if !ok {
		return err
	}

	id, err := app.dbProcessor.AddSegment(slug)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.Err{Text: err.Error()})
	}

	return c.JSON(models.ID{Value: id})
}

// DeleteSegment - удаляет сегмент из БД.
//
// Принимает: контекст.
//
// Возвращает: ошибку.
func (app *App) DeleteSegment(c *fiber.Ctx) error {
	if ok, err := checkType(c); !ok {
		return err
	}
	slug, ok, err := getSlug(c)
	if !ok {
		return err
	}

	err = app.dbProcessor.DeleteSegment(slug)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.Err{Text: err.Error()})
	}

	return c.Type("json").SendStatus(http.StatusOK)
}

// ModifyUser - изменяет сегменты пользователя.
//
// Принимает: контекст.
//
// Возвращает: ошибку.
func (app *App) ModifyUser(c *fiber.Ctx) error {
	if ok, err := checkType(c); !ok {
		return err
	}
	mod, ok, err := getUserMod(c)
	if !ok {
		return err
	}

	err = app.dbProcessor.ModifyUser(mod.Value, mod.Append, mod.Remove)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.Err{Text: err.Error()})
	}

	return c.Type("json").SendStatus(http.StatusOK)
}

// GetUserRelations - возвращает сегменты, в которых состоит пользователь.
//
// Принимает: контекст.
//
// Возвращает: ошибку.
func (app *App) GetUserRelations(c *fiber.Ctx) error {
	if ok, err := checkType(c); !ok {
		return err
	}
	id, ok, err := getID(c)
	if !ok {
		return err
	}

	slugs, err := app.dbProcessor.GetUserRelations(id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.Err{Text: err.Error()})
	}
	segments := make([]models.Segment, len(slugs))
	for i, slug := range slugs {
		segments[i].Slug = slug
	}

	return c.Type("json").JSON(segments)
}
