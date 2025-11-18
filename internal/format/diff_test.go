package format

import "testing"

func TestDiff(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		want  string
	}{
		{"zero", "--- old\r\n+++ new\r\n", ""},
		{"plus", "--- old\r\n+++ new\r\n+test\r\n", "<span class=\"plus\">+</span><span class=\"plus-line\">test</span><br />\n"},
		{"minus", "--- old\r\n+++ new\r\n-test\r\n", "<span class=\"minus\">-</span><span class=\"minus-line\">test</span><br />\n"},
		{"hunk", "--- old\r\n+++ new\r\n@@ -0,0 +0,0 @@\r\n", "<span class=\"hunk\">@@ -0,0 +0,0 @@</span><br />\n"},
		{"text", "--- old\r\n+++ new\r\n test\r\n", "&nbsp;test<br />\n"},
		{"plus indent", "--- old\r\n+++ new\r\n+  test\r\n", "<span class=\"plus\">+</span><span class=\"plus-line\">&nbsp;&nbsp;test</span><br />\n"},
		{"minus indent", "--- old\r\n+++ new\r\n-  test\r\n", "<span class=\"minus\">-</span><span class=\"minus-line\">&nbsp;&nbsp;test</span><br />\n"},
		{"text indent", "--- old\r\n+++ new\r\n   test\r\n", "&nbsp;&nbsp;&nbsp;test<br />\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Diff(tt.text)
			if got != tt.want {
				t.Errorf("Diff(\"%s\") = %s; want %s", tt.text, got, tt.want)
			}
		})
	}
}
