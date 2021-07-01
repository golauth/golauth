package util

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func TestExtractToken_ok(t *testing.T) {
	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZmlyc3ROYW1lIjoiQWRtaW4iLCJsYXN0TmFtZSI6IkFkbWluIiwiYXV0aG9yaXRpZXMiOlsiQURNSU4iLCJVU0VSIl0sImV4cCI6MTYyNTExMDI4MH0.aXZnvA7IGvVbXcv3xYWv2ApCzb4mSfCElDS2-8I0Eoey2yZjTXun7ToKZEp3ANUSNsAp0Cc2T-NwsvXw-28ZzJG6OW1BmZ8in6DGk5c82zWEuokt_oqF496jZC4doeomop39dO-ETgpD1j63M6jzwz0joecbvCg_rixYdtN52Ix6ekIFMae6mvElD68wLTIlJLp6ld58on_jyHV3o5K13SUhP8SHkFJzUfgVaJxLGFRAa8qeOPJakTDsIqigbOUQVw3RdNGVpCGwCj86G9NWhcz0SdMsOMLsnLAhqUSOf6sqyagt3-mvquD_ehv4KDdx8g1wLzsz62bwJUzl85PdJQ"
	reader := strings.NewReader("")
	req, err := http.NewRequest(http.MethodPost, "/check_token", reader)
	if err != nil {
		t.Fatal("error when creating request: %w", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	extracted, err := ExtractToken(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, extracted)
	assert.Equal(t, token, extracted)
}

func TestExtractToken_not_ok(t *testing.T) {
	reader := strings.NewReader("")
	req, err := http.NewRequest(http.MethodPost, "/check_token", reader)
	if err != nil {
		t.Fatal("error when creating request: %w", err)
	}
	extracted, err := ExtractToken(req)
	assert.Error(t, err)
	assert.Empty(t, extracted)
	assert.ErrorAs(t, err, &ErrBearerTokenExtract)
}
