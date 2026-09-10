package user

import (
	"errors"
	"strings"
	"testing"

	"github.com/golauth/golauth/pkg/domain/entity"
	"github.com/stretchr/testify/require"
)

func validUser() *entity.User {
	return &entity.User{
		Username:  "john.doe",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Document:  "12345678900",
		Password:  "correcthorsebatterystaple",
	}
}

// fieldMessage returns the message reported for field, or "" if the field was
// not rejected.
func fieldMessage(err error, field string) string {
	var ve *ValidationError
	if !errors.As(err, &ve) {
		return ""
	}
	for _, f := range ve.Fields {
		if f.Field == field {
			return f.Message
		}
	}
	return ""
}

func TestValidateAndNormalize_AcceptsAValidUser(t *testing.T) {
	u := validUser()
	require.NoError(t, validateAndNormalize(u, nil))
}

func TestValidateAndNormalize_RejectsPerRule(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*entity.User)
		field   string
		message string
	}{
		{"username empty", func(u *entity.User) { u.Username = "  " }, "username", "is required"},
		{"username too short", func(u *entity.User) { u.Username = "ab" }, "username", "must be between 3 and 50 characters"},
		{"username too long", func(u *entity.User) { u.Username = strings.Repeat("a", 51) }, "username", "must be between 3 and 50 characters"},
		{"username bad chars", func(u *entity.User) { u.Username = "john doe!" }, "username", "may contain only letters, digits, '.', '_' and '-'"},
		{"email empty", func(u *entity.User) { u.Email = "" }, "email", "is required"},
		{"email malformed", func(u *entity.User) { u.Email = "not-an-email" }, "email", "is not a valid e-mail address"},
		{"firstName empty", func(u *entity.User) { u.FirstName = "   " }, "firstName", "is required"},
		{"firstName too long", func(u *entity.User) { u.FirstName = strings.Repeat("a", 256) }, "firstName", "must be at most 255 characters"},
		{"lastName empty", func(u *entity.User) { u.LastName = "" }, "lastName", "is required"},
		{"lastName too long", func(u *entity.User) { u.LastName = strings.Repeat("a", 256) }, "lastName", "must be at most 255 characters"},
		{"document empty", func(u *entity.User) { u.Document = "" }, "document", "is required"},
		{"document too long", func(u *entity.User) { u.Document = strings.Repeat("9", 101) }, "document", "must be at most 100 characters"},
		{"password empty", func(u *entity.User) { u.Password = "" }, "password", "is required"},
		{"password too short", func(u *entity.User) { u.Password = "short123" }, "password", "must be at least 12 characters"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			u := validUser()
			c.mutate(u)
			err := validateAndNormalize(u, nil)
			require.Error(t, err)
			require.Equal(t, c.message, fieldMessage(err, c.field), "wrong message for field %q", c.field)
		})
	}
}

// A 73-byte password must be rejected, not silently truncated to 72 by bcrypt.
func TestValidateAndNormalize_RejectsPasswordOver72Bytes(t *testing.T) {
	u := validUser()
	u.Password = strings.Repeat("a", 73)
	err := validateAndNormalize(u, nil)
	require.Error(t, err)
	require.Equal(t, "must be at most 72 bytes", fieldMessage(err, "password"))
}

func TestValidateAndNormalize_Accepts72BytePassword(t *testing.T) {
	u := validUser()
	u.Password = strings.Repeat("a", 72)
	require.NoError(t, validateAndNormalize(u, nil))
}

func TestValidateAndNormalize_NormalisesUsernameEmailAndTrims(t *testing.T) {
	u := &entity.User{
		Username:  "  John.DOE  ",
		FirstName: "  John  ",
		LastName:  "  Doe  ",
		Email:     "  John@Example.COM  ",
		Document:  "  123  ",
		Password:  "correcthorsebattery",
	}
	require.NoError(t, validateAndNormalize(u, nil))
	require.Equal(t, "john.doe", u.Username)
	require.Equal(t, "john@example.com", u.Email)
	require.Equal(t, "John", u.FirstName)
	require.Equal(t, "Doe", u.LastName)
	require.Equal(t, "123", u.Document)
	require.Equal(t, "correcthorsebattery", u.Password, "password must not be trimmed")
}

func TestValidateAndNormalize_ExtractsAddrSpecFromEmail(t *testing.T) {
	u := validUser()
	u.Email = "John Doe <John.Doe@Example.com>"
	require.NoError(t, validateAndNormalize(u, nil))
	require.Equal(t, "john.doe@example.com", u.Email)
}

func TestValidateAndNormalize_CollectsEveryFailure(t *testing.T) {
	u := &entity.User{}
	err := validateAndNormalize(u, nil)
	var ve *ValidationError
	require.ErrorAs(t, err, &ve)
	got := make(map[string]bool)
	for _, f := range ve.Fields {
		got[f.Field] = true
	}
	for _, field := range []string{"username", "email", "firstName", "lastName", "document", "password"} {
		require.True(t, got[field], "expected a failure for %q", field)
	}
}

func TestValidateAndNormalize_DenylistRejectsCommonPassword(t *testing.T) {
	dl := newSetDenylist("hunter2loremipsum\n")
	u := validUser()
	u.Password = "hunter2loremipsum"
	err := validateAndNormalize(u, dl)
	require.Error(t, err)
	require.Equal(t, "is among the most commonly used passwords", fieldMessage(err, "password"))

	u2 := validUser()
	require.NoError(t, validateAndNormalize(u2, dl), "a password not on the list is accepted")
}

func TestLoadPasswordDenylist(t *testing.T) {
	t.Setenv("PASSWORD_DENYLIST", "")
	require.Nil(t, LoadPasswordDenylist())

	t.Setenv("PASSWORD_DENYLIST", "on")
	dl := LoadPasswordDenylist()
	require.NotNil(t, dl)
	require.True(t, dl.Contains("password"), "embedded list should contain \"password\"")
	require.False(t, dl.Contains("correcthorsebatterystaple"))

	t.Setenv("PASSWORD_DENYLIST", "/no/such/file/at/all")
	require.Nil(t, LoadPasswordDenylist(), "an unreadable file disables the check rather than blocking startup")
}
