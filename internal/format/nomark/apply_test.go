package nomark

import "testing"

var mockCfg = formatConfig{
	image: imageConfig{
		domains:    []string{"example.org", "example.net"},
		extensions: []string{"png", "jpeg"},
	},
}

func TestApply(t *testing.T) {
	tests := []struct {
		name  string
		title string
		text  string
		wantTitle  string
		wantText  string
		wantPlain  string
		wantHTML  string
	}{
		{"simple", "WikiPage", "This is a test.", "WikiPage", "\nThis is a test.", "\nThis is a test.", "<p>\nThis is a test.\n</p>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTitle, gotText, gotPlain, gotHTML := Apply(mockCfg, tt.title, tt.text)
			if gotTitle != tt.wantTitle || gotText != tt.wantText || gotPlain != tt.wantPlain || gotHTML != tt.wantHTML {
				t.Errorf("Apply(\"%s\", \"%s\") = %s, %s, %s, %s; want %s, %s, %s, %s", tt.title, tt.text, gotTitle, gotText, gotPlain, gotHTML, tt.wantTitle, tt.wantText, tt.wantPlain, tt.wantHTML)
			}
		})
	}
}
