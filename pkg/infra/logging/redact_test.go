package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestRedact(t *testing.T) {
	if got := Redact(""); got != "" {
		t.Fatalf("Redact(\"\") = %q, want empty", got)
	}
	for _, secret := range []string{"Bearer abc.def.ghi", "hunter2", "x"} {
		if got := Redact(secret); got != Placeholder {
			t.Fatalf("Redact(%q) = %q, want %q", secret, got, Placeholder)
		}
	}
}

func TestContextHandlerAddsRequestID(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := slog.New(WithContext(slog.NewJSONHandler(buf, nil)))

	ctx := ContextWithRequestID(context.Background(), "req-42")
	logger.InfoContext(ctx, "hello")
	logger.InfoContext(context.Background(), "no id here")
	logger.InfoContext(ctx, "explicit wins", "request_id", "caller-set")

	lines := decodeLines(t, buf)
	if got := lines[0]["request_id"]; got != "req-42" {
		t.Fatalf("line 0 request_id = %v, want req-42", got)
	}
	if _, ok := lines[1]["request_id"]; ok {
		t.Fatalf("line 1 should carry no request_id, got %v", lines[1]["request_id"])
	}
	if got := lines[2]["request_id"]; got != "caller-set" {
		t.Fatalf("line 2 request_id = %v, want caller-set (no duplicate, caller wins)", got)
	}
}

func decodeLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var out []map[string]any
	dec := json.NewDecoder(buf)
	for dec.More() {
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			t.Fatalf("decode: %v", err)
		}
		out = append(out, m)
	}
	return out
}

func TestLevelDefaultsToInfo(t *testing.T) {
	for in, want := range map[string]string{
		"":        "INFO",
		"bogus":   "INFO",
		"debug":   "DEBUG",
		" WARN ":  "WARN",
		"warning": "WARN",
		"error":   "ERROR",
	} {
		if got := level(in).String(); got != want {
			t.Fatalf("level(%q) = %q, want %q", in, got, want)
		}
	}
}
