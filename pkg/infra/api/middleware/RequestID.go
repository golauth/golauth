package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golauth/golauth/pkg/infra/api/apictx"
	"github.com/google/uuid"
)

// maxInboundRequestID caps how much of a client-supplied X-Request-Id we trust;
// anything longer (or empty) is replaced with a generated id.
const maxInboundRequestID = 128

// RequestID assigns every request a correlation id: an inbound X-Request-Id is
// reused when it is short and non-empty, otherwise one is generated. The id is
// published for the error handler and echoed on the response so a user
// reporting a 500 can hand over the exact key to the server-side line.
//
// Register it first, ahead of recover and the security middleware, so even a
// panic or an auth failure carries an id.
func RequestID() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		id := strings.TrimSpace(ctx.Get(apictx.RequestIDHeader))
		if id == "" || len(id) > maxInboundRequestID {
			id = uuid.NewString()
		}
		apictx.SetRequestID(ctx, id)
		ctx.Set(apictx.RequestIDHeader, id)
		return ctx.Next()
	}
}
