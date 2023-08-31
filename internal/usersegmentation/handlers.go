package usersegmentation

import (
	"net/http"
	"strconv"

	"github.com/famusovsky/AvitoTestTask/internal/usersegmentation/models"
	"github.com/gofiber/fiber/v2"
)

// PostSegment - добавляет сегмент в БД.
//
// Принимает: контекст.
//
// Возвращает: ошибку.

// @Summary      Adds segment to DB.
// @Description  Add segment with the specified slug to DB and get it's ID.
// @Tags         Segments
// @Accept       json
// @Produce      json
// @Param        slug body models.Segment true "Segment slug"
// @Success      200 {object} models.ID
// @Failure      400 {object} models.Err
// @Failure      500 {object} models.Err
// @Router       /segments [post]
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

// @Summary      Deletes segment from DB.
// @Description  Delete segment with the specified slug from DB.
// @Tags         Segments
// @Accept       json
// @Produce      json
// @Param        slug body models.Segment true "Segment slug"
// @Success      200 {string} string "OK"
// @Failure      400 {object} models.Err
// @Failure      500 {object} models.Err
// @Router       /segments [delete]
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

	return c.JSON("OK")
}

// ModifyUser - изменяет сегменты пользователя.
//
// Принимает: контекст.
//
// Возвращает: ошибку.

// @Summary      Modifies user's relations with segments.
// @Description  Append and remove user with the specified ID to/from segments.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        params body models.UserModification true "User modification parameters"
// @Success      200 {string} string "OK"
// @Failure      400 {object} models.Err
// @Failure      500 {object} models.Err
// @Router       /users [patch]
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

	return c.JSON("OK")
}

// GetUserRelations - возвращает сегменты, в которых состоит пользователь.
//
// Принимает: контекст.
//
// Возвращает: ошибку.

// @Summary      Returns segments in which the user is located.
// @Description  Get a list of segments in which the user with the specified ID is located.
// @Tags         Users
// @Produce      json
// @Param        id path int true "User ID"
// @Success      200 {object} []models.Segment
// @Failure      400 {object} models.Err
// @Failure      500 {object} models.Err
// @Router       /users/{id} [get]
func (app *App) GetUserRelations(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.Err{Text: `path parameter "id" must be an integer`})
	}

	slugs, err := app.dbProcessor.GetUserRelations(id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.Err{Text: err.Error()})
	}
	segments := make([]models.Segment, len(slugs))
	for i, slug := range slugs {
		segments[i].Slug = slug
	}

	return c.JSON(segments)
}
