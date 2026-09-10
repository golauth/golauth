package entity

import "testing"

func TestUserIsActive(t *testing.T) {
	cases := []struct {
		name    string
		enabled bool
		want    bool
	}{
		{"enabled account authenticates", true, true},
		{"disabled account does not", false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			u := &User{Enabled: c.enabled}
			if got := u.IsActive(); got != c.want {
				t.Fatalf("IsActive() = %v, want %v", got, c.want)
			}
		})
	}
}
