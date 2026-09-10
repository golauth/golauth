package user

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/golauth/golauth/pkg/domain/entity"
)

// Password policy. The 12-character floor is a deliberate minimum; the 72-byte
// ceiling is not a preference: bcrypt silently ignores every byte past 72, so a
// longer secret only gives a false sense of strength.
const (
	MinPasswordLen  = 12
	MaxPasswordByte = 72
)

const (
	minUsernameLen = 3
	maxUsernameLen = 50
	maxNameLen     = 255 // matches golauth_user.first_name / last_name varchar(255)
	maxDocumentLen = 100 // matches golauth_user.document varchar(100)
)

// usernamePattern is the full set of characters a username may contain. It is
// applied after lower-casing, so no uppercase class is needed.
var usernamePattern = regexp.MustCompile(`^[a-z0-9._-]+$`)

// FieldError names one rejected field and says why in prose meant for the API
// caller, never for a log grep.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError collects every field-level failure from a single request so
// the caller fixes the whole form at once rather than one round trip per rule.
type ValidationError struct {
	Fields []FieldError `json:"fields"`
}

func (e *ValidationError) Error() string {
	parts := make([]string, len(e.Fields))
	for i, f := range e.Fields {
		parts[i] = fmt.Sprintf("%s: %s", f.Field, f.Message)
	}
	return "invalid signup input: " + strings.Join(parts, "; ")
}

func (e *ValidationError) add(field, message string) {
	e.Fields = append(e.Fields, FieldError{Field: field, Message: message})
}

// validateAndNormalize checks input against the signup policy and rewrites the
// fields it accepts into canonical form -- username and email lower-cased, every
// string trimmed -- so the stored row is the normalised one. It returns a
// *ValidationError when any rule fails and leaves input partially normalised in
// that case; callers must not persist a rejected input.
func validateAndNormalize(input *entity.User, denylist PasswordDenylist) error {
	ve := &ValidationError{}

	input.Username = strings.ToLower(strings.TrimSpace(input.Username))
	switch {
	case input.Username == "":
		ve.add("username", "is required")
	case len(input.Username) < minUsernameLen || len(input.Username) > maxUsernameLen:
		ve.add("username", fmt.Sprintf("must be between %d and %d characters", minUsernameLen, maxUsernameLen))
	case !usernamePattern.MatchString(input.Username):
		ve.add("username", "may contain only letters, digits, '.', '_' and '-'")
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Email == "" {
		ve.add("email", "is required")
	} else if addr, err := mail.ParseAddress(input.Email); err != nil {
		ve.add("email", "is not a valid e-mail address")
	} else {
		// ParseAddress accepts "Name <addr>" forms; keep only the addr-spec.
		input.Email = strings.ToLower(addr.Address)
	}

	input.FirstName = strings.TrimSpace(input.FirstName)
	if input.FirstName == "" {
		ve.add("firstName", "is required")
	} else if len(input.FirstName) > maxNameLen {
		ve.add("firstName", fmt.Sprintf("must be at most %d characters", maxNameLen))
	}

	input.LastName = strings.TrimSpace(input.LastName)
	if input.LastName == "" {
		ve.add("lastName", "is required")
	} else if len(input.LastName) > maxNameLen {
		ve.add("lastName", fmt.Sprintf("must be at most %d characters", maxNameLen))
	}

	// document is required: the golauth_user.document column is NOT NULL and has
	// no default. Relaxing that needs a migration (see Plan 10).
	input.Document = strings.TrimSpace(input.Document)
	if input.Document == "" {
		ve.add("document", "is required")
	} else if len(input.Document) > maxDocumentLen {
		ve.add("document", fmt.Sprintf("must be at most %d characters", maxDocumentLen))
	}

	// The password is never trimmed or otherwise rewritten -- leading and
	// trailing spaces are legitimate secret material.
	switch {
	case input.Password == "":
		ve.add("password", "is required")
	case len([]rune(input.Password)) < MinPasswordLen:
		ve.add("password", fmt.Sprintf("must be at least %d characters", MinPasswordLen))
	case len(input.Password) > MaxPasswordByte:
		ve.add("password", fmt.Sprintf("must be at most %d bytes", MaxPasswordByte))
	case denylist != nil && denylist.Contains(input.Password):
		ve.add("password", "is among the most commonly used passwords")
	}

	if len(ve.Fields) > 0 {
		return ve
	}
	return nil
}
