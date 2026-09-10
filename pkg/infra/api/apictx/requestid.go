package apictx

import "github.com/gofiber/fiber/v3"

// RequestIDHeader is the inbound and outbound header carrying the correlation id.
const RequestIDHeader = "X-Request-Id"

type requestIDKey struct{}

// SetRequestID publishes the correlation id for the current request. Only the
// RequestID middleware should call it.
func SetRequestID(ctx fiber.Ctx, id string) {
	ctx.Locals(requestIDKey{}, id)
}

// RequestID returns the correlation id assigned to the current request, or ""
// when the request did not pass through the RequestID middleware.
func RequestID(ctx fiber.Ctx) string {
	if v, ok := ctx.Locals(requestIDKey{}).(string); ok {
		return v
	}
	return ""
}
