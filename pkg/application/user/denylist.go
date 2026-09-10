package user

import (
	"bufio"
	_ "embed"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// PasswordDenylist rejects a candidate password that appears verbatim in a set
// of known-bad passwords. A nil PasswordDenylist means the check is disabled.
type PasswordDenylist interface {
	Contains(password string) bool
}

//go:embed common_passwords.txt
var embeddedCommonPasswords string

type setDenylist struct {
	entries map[string]struct{}
}

func (d *setDenylist) Contains(password string) bool {
	_, bad := d.entries[strings.ToLower(strings.TrimSpace(password))]
	return bad
}

func newSetDenylist(raw string) *setDenylist {
	entries := make(map[string]struct{})
	sc := bufio.NewScanner(strings.NewReader(raw))
	for sc.Scan() {
		line := strings.ToLower(strings.TrimSpace(sc.Text()))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entries[line] = struct{}{}
	}
	return &setDenylist{entries: entries}
}

// LoadPasswordDenylist reads PASSWORD_DENYLIST and returns the denylist to
// enforce, or nil when the feature is off. Accepted values:
//
//	unset / "" / "off" / "false" / "0"  -> disabled (nil)
//	"on" / "true" / "1" / "embedded"    -> the small embedded common-password list
//	any other value                     -> path to a newline-delimited file
//
// A configured file that cannot be read is logged and treated as disabled
// rather than blocking startup.
func LoadPasswordDenylist() PasswordDenylist {
	v := strings.TrimSpace(os.Getenv("PASSWORD_DENYLIST"))
	switch strings.ToLower(v) {
	case "", "off", "false", "0", "no":
		return nil
	case "on", "true", "1", "yes", "embedded":
		return newSetDenylist(embeddedCommonPasswords)
	}

	// #nosec G304 G703 -- v is the operator-supplied PASSWORD_DENYLIST path
	data, err := os.ReadFile(v)
	if err != nil {
		logrus.Warnf("PASSWORD_DENYLIST=%q could not be read (%v); password denylist disabled", v, err)
		return nil
	}
	return newSetDenylist(string(data))
}
