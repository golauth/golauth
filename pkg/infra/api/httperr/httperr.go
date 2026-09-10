// Package httperr is the single place the API turns an error into an HTTP
// response. Handlers and middleware return a plain error -- a sentinel from
// pkg/domain/apperr, a *fiber.Error with a fixed message, a *user.ValidationError,
// or anything unmapped -- and Handler chooses the status code and the
// client-visible body. No wrapped driver text, query fragment or id ever
// reaches the client.
package httperr

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/application/user"
	"github.com/golauth/golauth/pkg/domain/apperr"
	"github.com/golauth/golauth/pkg/infra/api/apictx"
	"github.com/sirupsen/logrus"
)

// Detail is the error object. code is the stable, machine-readable key; message
// is human prose that carries no internal detail.
type Detail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId,omitempty"`
}

// Body is the documented error envelope: {"error": {...}} plus an optional
// "fields" array for field-level validation failures.
type Body struct {
	Error  Detail            `json:"error"`
	Fields []user.FieldError `json:"fields,omitempty"`
}

// Handler is the fiber.Config.ErrorHandler for the API.
func Handler(ctx fiber.Ctx, err error) error {
	reqID := apictx.RequestID(ctx)

	// Field-level validation failure keeps its per-field detail (Plan 04).
	var ve *user.ValidationError
	if errors.As(err, &ve) {
		return write(ctx, fiber.StatusBadRequest, Body{
			Error:  Detail{Code: "invalid_input", Message: "one or more fields are invalid", RequestID: reqID},
			Fields: ve.Fields,
		})
	}

	status, code := classify(err)

	if status >= fiber.StatusInternalServerError {
		// The only place the full error is allowed to exist: the server log,
		// keyed by the same id the client is handed.
		logrus.WithFields(logrus.Fields{
			"event":      "request_error",
			"request_id": reqID,
			"method":     ctx.Method(),
			"path":       ctx.Path(),
			"status":     status,
		}).Errorf("unhandled error: %v", err)
		return write(ctx, status, Body{Error: Detail{Code: code, Message: "internal server error", RequestID: reqID}})
	}

	return write(ctx, status, Body{Error: Detail{Code: code, Message: clientMessage(err, status), RequestID: reqID}})
}

func classify(err error) (int, string) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		return fiber.StatusNotFound, "not_found"
	case errors.Is(err, apperr.ErrAlreadyExists):
		return fiber.StatusConflict, "already_exists"
	case errors.Is(err, apperr.ErrInvalidInput):
		return fiber.StatusBadRequest, "invalid_input"
	case errors.Is(err, apperr.ErrUnauthorized):
		return fiber.StatusUnauthorized, "unauthorized"
	case errors.Is(err, apperr.ErrForbidden):
		return fiber.StatusForbidden, "forbidden"
	}

	var fe *fiber.Error
	if errors.As(err, &fe) {
		return fe.Code, codeForStatus(fe.Code)
	}
	return fiber.StatusInternalServerError, "internal_error"
}

// clientMessage is the prose for a non-5xx failure. A *fiber.Error carries a
// message set at the call site, which is always a fixed literal we control; a
// bare domain sentinel gets a generic phrase so no wrapped context leaks.
func clientMessage(err error, status int) string {
	var fe *fiber.Error
	if errors.As(err, &fe) && fe.Message != "" {
		return fe.Message
	}
	return genericMessage(status)
}

func codeForStatus(status int) string {
	switch status {
	case fiber.StatusBadRequest:
		return "invalid_input"
	case fiber.StatusUnauthorized:
		return "unauthorized"
	case fiber.StatusForbidden:
		return "forbidden"
	case fiber.StatusNotFound:
		return "not_found"
	case fiber.StatusMethodNotAllowed:
		return "method_not_allowed"
	case fiber.StatusConflict:
		return "already_exists"
	case fiber.StatusUnprocessableEntity:
		return "unprocessable_entity"
	case fiber.StatusTooManyRequests:
		return "rate_limited"
	default:
		return "error"
	}
}

func genericMessage(status int) string {
	switch status {
	case fiber.StatusBadRequest:
		return "invalid request"
	case fiber.StatusUnauthorized:
		return "unauthorized"
	case fiber.StatusForbidden:
		return "forbidden"
	case fiber.StatusNotFound:
		return "resource not found"
	case fiber.StatusMethodNotAllowed:
		return "method not allowed"
	case fiber.StatusConflict:
		return "resource already exists"
	case fiber.StatusTooManyRequests:
		return "too many requests"
	default:
		return "error"
	}
}

func write(ctx fiber.Ctx, status int, body Body) error {
	return ctx.Status(status).JSON(body)
}
