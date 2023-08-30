package usersegmentation

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber"
)

// processorMock - mock для обработчика БД.
type processorMock struct {
	resOnAddSegment       int
	errOnAddSegment       error
	errOnDeleteSegment    error
	errOnModifyUser       error
	resOnGetUserRelations []string
	errOnGetUserRelations error
}

func (p processorMock) AddSegment(slug string) (int, error) {
	return p.resOnAddSegment, p.errOnAddSegment
}
func (p processorMock) DeleteSegment(slug string) error {
	return p.errOnDeleteSegment
}
func (p processorMock) ModifyUser(id int, append []string, remove []string) error {
	return p.errOnModifyUser
}
func (p processorMock) GetUserRelations(id int) ([]string, error) {
	return p.resOnGetUserRelations, p.errOnGetUserRelations
}
func (p *processorMock) CleanUp() {
	p.resOnAddSegment = 0
	p.errOnAddSegment = nil
	p.errOnDeleteSegment = nil
	p.errOnModifyUser = nil
	p.resOnGetUserRelations = []string{}
	p.errOnGetUserRelations = nil
}

var (
	contentTypeErr = []byte(`{"error":"request's Content-Type must be application/json"}`)
)

// Test_Segments - тестирование обработки запросов по адресу /segments.
func Test_Segments(t *testing.T) {
	processor := &processorMock{}
	app := CreateApp(log.Default(), processor)

	for i := 0; i < 10; i++ {
		var (
			reqBody         = `{"slug":"test"}`
			testErr         = fmt.Errorf("test error %d", rand.Int())
			wrongReqErrText = []byte(`{"error":"request's body must implement the template {\"slug\":\"some text\"}"}`)
		)

		t.Run("normal case", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(reqBody, fiber.MethodPost, "/segments", fiber.MIMEApplicationJSON)
			id := rand.Int()
			processor.resOnAddSegment = id

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, []byte(fmt.Sprintf(`{"id":%d}`, id)), http.StatusOK, fiber.MIMEApplicationJSON, t)

			req.Method = fiber.MethodDelete

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, []byte(`"OK"`), http.StatusOK, fiber.MIMEApplicationJSON, t)
		})

		t.Run("error while handling db", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(reqBody, fiber.MethodPost, "/segments", fiber.MIMEApplicationJSON)
			processor.errOnAddSegment = testErr

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, []byte(fmt.Sprintf(`{"error":"%s"}`, testErr)), http.StatusInternalServerError, fiber.MIMEApplicationJSON, t)

			req.Method = fiber.MethodDelete
			processor.errOnDeleteSegment = testErr

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, []byte(fmt.Sprintf(`{"error":"%s"}`, testErr)), http.StatusInternalServerError, fiber.MIMEApplicationJSON, t)
		})

		t.Run("bad request - wrong content type", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(reqBody, fiber.MethodPost, "/segments", "xml")

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, contentTypeErr, http.StatusBadRequest, fiber.MIMEApplicationJSON, t)

			req.Method = fiber.MethodDelete

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, contentTypeErr, http.StatusBadRequest, fiber.MIMEApplicationJSON, t)
		})

		t.Run("bad request - non marshable body", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(`{"smth":"is wrong"}`, fiber.MethodPost, "/segments", fiber.MIMEApplicationJSON)

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, wrongReqErrText, http.StatusBadRequest, fiber.MIMEApplicationJSON, t)

			req.Method = fiber.MethodDelete

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, wrongReqErrText, http.StatusBadRequest, fiber.MIMEApplicationJSON, t)
		})

		t.Run("bad request - empty input", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(``, fiber.MethodPost, "/segments", fiber.MIMEApplicationJSON)

			resp, err := app.webApp.Test(req)

			checkResponse(resp, err, wrongReqErrText, http.StatusBadRequest, fiber.MIMEApplicationJSON, t)

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, wrongReqErrText, http.StatusBadRequest, fiber.MIMEApplicationJSON, t)
		})
	}
}

// Test_Users - тестирование обработки запросов по адресу /users.
func Test_Users(t *testing.T) {
	processor := &processorMock{}
	app := CreateApp(log.Default(), processor)

	for i := 0; i < 10; i++ {
		var (
			testErr                = fmt.Errorf("test error %d", rand.Int())
			userModReqBody         = `{"id":10,"append":["test1","test2"],"remove":["test3","test4"]}`
			userModRespBody        = []byte(`[{"slug":"test1"},{"slug":"test2"}]`)
			userModWrongReqErrText = []byte(`{"error":"request's body must implement the template {\"id\":0,\"append\":[\"test1\",\"test2\"],\"remove\":[\"test3\",\"test4\"]}"}`)
			getUserWrongReqErrText = []byte(`{"error":"path parameter \"id\" must be an integer"}`)
		)

		t.Run("normal case", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(userModReqBody,
				fiber.MethodPatch, "/users", fiber.MIMEApplicationJSON)

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, []byte(`"OK"`), http.StatusOK, fiber.MIMEApplicationJSON, t)

			req = createRequest(``, fiber.MethodGet, "/users/0", fiber.MIMEApplicationJSON)
			processor.resOnGetUserRelations = []string{"test1", "test2"}

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, userModRespBody, http.StatusOK, fiber.MIMEApplicationJSON, t)
		})

		t.Run("error while handling db", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(userModReqBody,
				fiber.MethodPatch, "/users", fiber.MIMEApplicationJSON)
			processor.errOnModifyUser = testErr

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, []byte(fmt.Sprintf(`{"error":"%s"}`, testErr)), http.StatusInternalServerError, fiber.MIMEApplicationJSON, t)

			req = createRequest(``, fiber.MethodGet, "/users/0", fiber.MIMEApplicationJSON)
			processor.errOnGetUserRelations = testErr

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, []byte(fmt.Sprintf(`{"error":"%s"}`, testErr)), http.StatusInternalServerError, fiber.MIMEApplicationJSON, t)
		})

		t.Run("bad request - wrong content type", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(userModReqBody,
				fiber.MethodPatch, "/users", "xml")

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, contentTypeErr, http.StatusBadRequest, fiber.MIMEApplicationJSON, t)

			req = createRequest(``, fiber.MethodGet, "/users/0", "xml")
			processor.resOnGetUserRelations = []string{"test1", "test2"}

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, userModRespBody, http.StatusOK, fiber.MIMEApplicationJSON, t)
		})

		t.Run("bad request - non marshable body", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(`{"smth":"is wrong"}`, fiber.MethodPatch, "/users", fiber.MIMEApplicationJSON)

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, userModWrongReqErrText, http.StatusBadRequest, fiber.MIMEApplicationJSON, t)

			req = createRequest(``, fiber.MethodGet, "/users/wrong", fiber.MIMEApplicationJSON)

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, getUserWrongReqErrText, http.StatusBadRequest, fiber.MIMEApplicationJSON, t)
		})

		t.Run("bad request - empty input", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(``, fiber.MethodPatch, "/users", fiber.MIMEApplicationJSON)

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, userModWrongReqErrText, http.StatusBadRequest, fiber.MIMEApplicationJSON, t)

			expectedErr := []byte(fmt.Sprintf(`{"error":"%s"}`, fiber.ErrMethodNotAllowed.Message))
			req = createRequest(``, fiber.MethodGet, "/users", fiber.MIMEApplicationJSON)

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, expectedErr, http.StatusInternalServerError, fiber.MIMEApplicationJSON, t)
		})
	}
}

func checkResponse(got *http.Response, gotErr error, expectedBody []byte, expectedStatusCode int, expectedContentType string, t *testing.T) {
	if gotErr != nil {
		t.Errorf("unexpected: %s\n", gotErr)
	}

	if got.StatusCode != expectedStatusCode {
		t.Errorf("got status code: %d\nexpected: %d\n", got.StatusCode, expectedStatusCode)
	}

	gotContentType := got.Header.Get("Content-Type")
	if gotContentType != expectedContentType {
		t.Errorf("got content type: %s\nexpected: %s\n", gotContentType, expectedContentType)
	}

	gotBody := make([]byte, 0)
	if got.Body != nil {
		gotBody, _ = io.ReadAll(got.Body)
	}
	if !bytes.Equal(gotBody, expectedBody) {
		t.Errorf("got body: %s\nexpected: %s\n", gotBody, expectedBody)
	}
}

func createRequest(body string, method string, path string, contentType string) *http.Request {
	b := strings.NewReader(body)
	req := httptest.NewRequest(method, path, b)
	req.Header.Set("Content-Type", contentType)

	return req
}
