package logging

// Placeholder is the fixed string that stands in for a secret in a log line.
const Placeholder = "[REDACTED]"

// Redact returns Placeholder for any non-empty value and "" for an empty one.
// Use it when a log line needs to record that a credential was present --
// an Authorization header, a token, a password field -- without disclosing it.
// The value itself must never reach a log call unredacted.
func Redact(v string) string {
	if v == "" {
		return ""
	}
	return Placeholder
}
