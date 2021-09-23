package generator

import (
	"testing"
)

func TestPayload(t *testing.T) {
	result := payload("POST", 200, "/foo/bar")
	want := "[200] POST /foo/bar"
	if result != want {
		t.Errorf("payload(2) = \"%s\"; want %s", result, want)
	}
}

func TestGenerate(t *testing.T) {
	result := Generate(2)
	if len(result) != 2 {
		t.Errorf("Generate(2) = %d; want 2", len(result))
	}
}
