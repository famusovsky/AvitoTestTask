package usersegmentation

import (
	"bytes"
	"encoding/json"
	"net/http"

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
		err := c.Status(http.StatusBadRequest).JSON(models.Err{Text: `request's body must implement the template {"id":0,"append":["test1","test2"],"remove":["test3","test4"]}`})
		return mod, false, err
	}

	return mod, true, nil
}

// getLogTimestamps - получение временных рамок для логов из контекста.
//
// Принимает: контекст.
//
// Возвращает: временные рамки, флаг успешности, ошибку.
func getLogTimestamps(c *fiber.Ctx) (models.LogTimestamps, bool, error) {
	timestamps := models.LogTimestamps{}

	dec := json.NewDecoder(bytes.NewReader(c.Body()))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&timestamps); err != nil {
		err := c.Status(http.StatusBadRequest).JSON(models.Err{Text: `request's body must implement the template {"from":"2021-01-01T00:00:00Z","to":"2021-01-01T00:00:00Z"}`})
		return timestamps, false, err
	}

	return timestamps, true, nil
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
