package format

import (
	"runtime"
	"testing"
)

func mustPanicIndexLineEnd(t *testing.T, text string, index int) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic, got none")
		}
		if _, ok := r.(runtime.Error); !ok {
			t.Fatalf("expected runtime.Error, got %T", r)
		}
	}()
	indexLineEnd(text, index)
}

func TestIndexLineEndCrash(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		index int
	}{
		{"word", "test", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mustPanicIndexLineEnd(t, tt.text, tt.index)
		})
	}
}

func TestIndexLineEnd(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		index    int
		wantEnd  int
		wantNext int
	}{
		{"null", "", 0, 0, 0},
		{"word", "test", 0, 4, 4},
		{"word end", "test", 4, 4, 4},
		{"first line", "This is a test.\r\nI love tests.\r\n", 0, 15, 17},
		{"second line", "This is a test.\r\nI love tests.\r\n", 17, 30, 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEnd, gotNext := indexLineEnd(tt.text, tt.index)
			if gotEnd != tt.wantEnd || gotNext != tt.wantNext {
				t.Errorf("indexLineEnd(\"%s\", %d) = %d, %d; want %d, %d", tt.text, tt.index, gotEnd, gotNext, tt.wantEnd, tt.wantNext)
			}
		})
	}
}
