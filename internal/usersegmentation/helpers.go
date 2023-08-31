package usersegmentation

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/famusovsky/AvitoTestTask/internal/usersegmentation/models"
	"github.com/gofiber/fiber/v2"
)

// getSlug - получение имени сегмента из контекста.
//
// Принимает: контекст.
//
// Возвращает: имя сегмента, флаг успешности, ошибку.
func getSlug(c *fiber.Ctx) (string, bool, error) {
	segment := models.Segment{}

	dec := json.NewDecoder(bytes.NewReader(c.Body()))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&segment); err != nil {
		err := c.Status(http.StatusBadRequest).JSON(models.Err{Text: `request's body must implement the template {"slug":"some text"}`})
		return "", false, err
	}

	return segment.Slug, true, nil
}

// getUserMod - получение требуемых изменений пользователя из контекста.
//
// Принимает: контекст.
//
// Возвращает: требуемые изменения, флаг успешности, ошибку.
func getUserMod(c *fiber.Ctx) (models.UserModification, bool, error) {
	mod := models.UserModification{}

	dec := json.NewDecoder(bytes.NewReader(c.Body()))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&mod); err != nil {
		err := c.Status(http.StatusBadRequest).JSON(models.Err{Text: `request's body must implement the template {"id":1,"append":[{"slug":"test1","expires":"2023-08-31T15:15:39.104033+03:00"},{"slug":"test2","expires":"0001-01-01T00:00:00Z"}],"remove":[{"slug":"test3"},{"slug":"test4"}]}`})
		return mod, false, err
	}
	for i := 0; i < len(mod.Append); i++ {
		defTime := time.Time{}
		if mod.Append[i].Expires == defTime {
			mod.Append[i].Expires = time.Now().AddDate(100, 0, 0)
		}
	}

	return mod, true, nil
}

// checkType - проверка типа запроса на json.
//
// Принимает: контекст.
//
// Возвращает: флаг успешности, ошибку.
func checkType(c *fiber.Ctx) (bool, error) {
	if !c.Is("json") {
		err := c.Status(http.StatusBadRequest).JSON(models.Err{Text: `request's Content-Type must be application/json`})
		return false, err
	}

	return true, nil
}
