package middleware

import "github.com/gofiber/fiber/v3"

// Response headers set on every reply. Both were found by the nightly DAST scan
// against the running service, not chosen from a list.
const (
	// contentTypeOptions stops a browser from MIME-sniffing a response into
	// something other than what Content-Type says. Every route here answers
	// JSON; a browser that decides one is HTML turns a reflected value into
	// stored XSS, and the fix costs one header.
	contentTypeOptions = "nosniff"

	// resourcePolicy is deliberately permissive: golauth is a public API meant
	// to be called cross-origin, and which origins may do so is already decided
	// by CORS_ALLOWED_ORIGINS. Declaring `cross-origin` states that intent
	// rather than leaving the header absent for a scanner to flag. `same-origin`
	// would be a lie about what this service is for.
	resourcePolicy = "cross-origin"
)

// SecurityHeaders sets the response headers that apply to every route.
//
// It is intentionally small and explicit rather than fiber's helmet middleware:
// most of what helmet sets (CSP, X-Frame-Options, and the rest) governs HTML
// documents, and this service never returns one. Configuring nine headers away
// to keep two is harder to review than writing the two.
//
// Register it early, so a response produced by an auth rejection or a recovered
// panic carries the headers too.
func SecurityHeaders() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		ctx.Set(fiber.HeaderXContentTypeOptions, contentTypeOptions)
		ctx.Set("Cross-Origin-Resource-Policy", resourcePolicy)
		return ctx.Next()
	}
}
