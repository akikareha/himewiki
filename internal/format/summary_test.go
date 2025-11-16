package format

import "testing"

func mustPanicTrimForSummary(t *testing.T, want string, text string, limit int) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic, got none")
		}
		if r != want {
			t.Fatalf("panic = %v; want %v", r, want)
		}
	}()
	TrimForSummary(text, limit)
}

func TestTrimForSummaryCrash(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		limit int
		want  string
	}{
		{"minus one", "This is a test.", -1, "program error"},
		{"minus one mb", "これはテストです。", -1, "program error"},
		{"minus ten", "This is a test.", -10, "program error"},
		{"minus ten mb", "これはテストです。", -10, "program error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mustPanicTrimForSummary(t, tt.want, tt.text, tt.limit)
		})
	}
}

func TestTrimForSummary(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		limit int
		want  string
	}{
		{"zero trim", "This is a test.", 0, ""},
		{"zero trim mb", "これはテストです。", 0, ""},
		{"one trim", "This is a test.", 1, "."},
		{"one trim mb", "これはテストです。", 1, "."},
		{"one", "A", 1, "A"},
		{"one mb", "あ", 1, "あ"},
		{"two trim", "This is a test.", 2, ".."},
		{"two trim mb", "これはテストです。", 2, ".."},
		{"two", "AB", 2, "AB"},
		{"two mb", "あい", 2, "あい"},
		{"three trim", "This is a test.", 3, "T.."},
		{"three trim mb", "これはテストです。", 3, "こ.."},
		{"ten trim", "This is a test.", 10, "This is .."},
		{"five trim mb", "これはテストです。", 5, "これは.."},
		{"minus one trim", "This is a test.", 14, "This is a te.."},
		{"minus one trim mb", "これはテストです。", 8, "これはテスト.."},
		{"just", "This is a test.", 15, "This is a test."},
		{"just mb", "これはテストです。", 9, "これはテストです。"},
		{"sufficient", "This is a test.", 20, "This is a test."},
		{"sufficient mb", "これはテストです。", 15, "これはテストです。"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrimForSummary(tt.text, tt.limit)
			if got != tt.want {
				t.Errorf("TrimForSummary(\"%s\", %d) = %s; want %s", tt.text, tt.limit, got, tt.want)
			}
		})
	}
}
