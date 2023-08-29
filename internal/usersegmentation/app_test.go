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
			testErr         = fmt.Errorf("test error %d", rand.Int())
			wrongReqErrText = []byte(`{"error":"request must implement the template {\"slug\":\"some text\"}"}`)
		)

		t.Run("normal case", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(`{"slug":"test"}`, "POST", "/segments", "application/json")
			id := rand.Int()
			processor.resOnAddSegment = id

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, []byte(fmt.Sprintf(`{"id":%d}`, id)), http.StatusOK, "application/json", t)

			req.Method = "DELETE"

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, []byte("OK"), http.StatusOK, "application/json", t)
		})

		t.Run("error while handling db", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(`{"slug":"test"}`, "POST", "/segments", "application/json")
			processor.errOnAddSegment = testErr

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, []byte(fmt.Sprintf(`{"error":"%s"}`, testErr)), http.StatusInternalServerError, "application/json", t)

			req.Method = "DELETE"
			processor.errOnDeleteSegment = testErr

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, []byte(fmt.Sprintf(`{"error":"%s"}`, testErr)), http.StatusInternalServerError, "application/json", t)
		})

		t.Run("bad request - wrong content type", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(`{"slug":"test"}`, "POST", "/segments", "xml")

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, contentTypeErr, http.StatusBadRequest, "application/json", t)

			req.Method = "DELETE"

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, contentTypeErr, http.StatusBadRequest, "application/json", t)
		})

		t.Run("bad request - non marshable body", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(`{"smth":"is wrong"}`, "POST", "/segments", "application/json")

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, wrongReqErrText, http.StatusBadRequest, "application/json", t)

			req.Method = "DELETE"

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, wrongReqErrText, http.StatusBadRequest, "application/json", t)
		})

		// TODO
		// t.Run("bad request - empty slug", func(t *testing.T) {
		// 	defer processor.CleanUp()
		// 	body := strings.NewReader(`{"slug":""}`)
		// 	req := httptest.NewRequest("POST", "/segments", body)
		// 	req.Header.Set("Content-Type", "application/json")

		// 	resp, err := app.webApp.Test(req)

		// 	checkResponse(resp, err, wrongReqErrText, http.StatusBadRequest, "application/json", t)

		// 	req.Method = "DELETE"

		// 	resp, err = app.webApp.Test(req)
		// 	checkResponse(resp, err, wrongReqErrText, http.StatusBadRequest, "application/json", t)
		// })

		t.Run("bad request - empty input", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(``, "POST", "/segments", "application/json")

			resp, err := app.webApp.Test(req)

			checkResponse(resp, err, wrongReqErrText, http.StatusBadRequest, "application/json", t)

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, wrongReqErrText, http.StatusBadRequest, "application/json", t)
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
			userModWrongReqErrText = []byte(`{"error":"request must implement the template {\"id\":0,\"append\":[\"test1\",\"test2\"],\"remove\":[\"test3\",\"test4\"]}"}`)
			getUserWrongReqErrText = []byte(`{"error":"request must implement the template {\"id\":0}"}`)
		)

		t.Run("normal case", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(`{"id":10,"append":["test1","test2"],"remove":["test3","test4"]}`,
				"PATCH", "/users", "application/json")

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, []byte("OK"), http.StatusOK, "application/json", t)

			req = createRequest(`{"id":0}`, "GET", "/users", "application/json")
			processor.resOnGetUserRelations = []string{"test1", "test2"}

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, []byte(`[{"slug":"test1"},{"slug":"test2"}]`), http.StatusOK, "application/json", t)
		})

		t.Run("error while handling db", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(`{"id":10,"append":["test1","test2"],"remove":["test3","test4"]}`,
				"PATCH", "/users", "application/json")
			processor.errOnModifyUser = testErr

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, []byte(fmt.Sprintf(`{"error":"%s"}`, testErr)), http.StatusInternalServerError, "application/json", t)

			req = createRequest(`{"id":0}`, "GET", "/users", "application/json")
			processor.errOnGetUserRelations = testErr

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, []byte(fmt.Sprintf(`{"error":"%s"}`, testErr)), http.StatusInternalServerError, "application/json", t)
		})

		t.Run("bad request - wrong content type", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(`{"id":10,"append":["test1","test2"],"remove":["test3","test4"]}`,
				"PATCH", "/users", "xml")

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, contentTypeErr, http.StatusBadRequest, "application/json", t)

			req.Method = "GET"
			processor.resOnGetUserRelations = []string{"test1", "test2"}

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, contentTypeErr, http.StatusBadRequest, "application/json", t)
		})

		t.Run("bad request - non marshable body", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(`{"smth":"is wrong"}`, "PATCH", "/users", "application/json")

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, userModWrongReqErrText, http.StatusBadRequest, "application/json", t)

			req.Method = "GET"

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, getUserWrongReqErrText, http.StatusBadRequest, "application/json", t)
		})

		t.Run("bad request - empty input", func(t *testing.T) {
			defer processor.CleanUp()
			req := createRequest(``, "PATCH", "/users", "application/json")

			resp, err := app.webApp.Test(req)
			checkResponse(resp, err, userModWrongReqErrText, http.StatusBadRequest, "application/json", t)

			req.Method = "GET"

			resp, err = app.webApp.Test(req)
			checkResponse(resp, err, getUserWrongReqErrText, http.StatusBadRequest, "application/json", t)
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
