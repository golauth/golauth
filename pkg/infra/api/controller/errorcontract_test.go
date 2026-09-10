package controller

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/infra/api/httperr"
)

// newErrApp returns a fiber app wired with the production error handler, so a
// controller that returns a plain domain error is rendered as the real API
// error contract rather than fiber's default plain-text 500.
func newErrApp() *fiber.App {
	return fiber.New(fiber.Config{ErrorHandler: httperr.Handler})
}

type contractError struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"requestId"`
	} `json:"error"`
	Fields []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	} `json:"fields"`
}

// readContract decodes the error envelope and returns it alongside the raw body,
// so a test can both assert the code and check that no internal text leaked.
func readContract(resp *http.Response) (contractError, string) {
	raw, _ := io.ReadAll(resp.Body)
	var c contractError
	_ = json.Unmarshal(raw, &c)
	return c, string(raw)
}
