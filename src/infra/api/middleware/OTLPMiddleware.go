package middleware

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"net/http"
)

type OTLPMiddleware struct {
}

func NewOTLPMiddleware() *OTLPMiddleware {
	return &OTLPMiddleware{}
}

func (m OTLPMiddleware) Apply() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		tracer := otel.Tracer(fmt.Sprintf("%s %s", ctx.Method(), ctx.Path()))
		ctxTracer, span := tracer.Start(ctx.UserContext(), "request")
		span.SetAttributes()
		defer func() {
			span.End()
		}()

		ctx.SetUserContext(ctxTracer)

		err := ctx.Next()

		if err != nil {
			span.RecordError(err)
		}

		if ctx.Response().StatusCode() == http.StatusOK {
			span.SetStatus(codes.Ok, codes.Ok.String())
		} else {
			span.SetStatus(codes.Error, "ERROR")
		}

		return nil
	}
}
