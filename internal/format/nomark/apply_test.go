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
		name      string
		title     string
		text      string
		wantTitle string
		wantText  string
		wantPlain string
		wantHTML  string
	}{
		{
			"blank",
			"WikiPage",
			"",
			"WikiPage",
			"",
			"",
			"",
		},
		{
			"blank line",
			"WikiPage",
			"\n",
			"WikiPage",
			"",
			"",
			"",
		},
		{
			"simple",
			"WikiPage",
			"This is a test.",
			"WikiPage",
			"This is a test.\n",
			"This is a test.\n",
			"<p>\nThis is a test.\n</p>\n",
		},
		{
			"simple line",
			"WikiPage",
			"This is a test.\n",
			"WikiPage",
			"This is a test.\n",
			"This is a test.\n",
			"<p>\nThis is a test.\n</p>\n",
		},
		{
			"multi line",
			"WikiPage",
			"This is a test.\nI love tests.\n",
			"WikiPage",
			"This is a test.\nI love tests.\n",
			"This is a test.\nI love tests.\n",
			"<p>\nThis is a test.<br />\nI love tests.\n</p>\n",
		},
		{
			"paragraphs",
			"WikiPage",
			"This is a test.\nI love tests.\n\nHello, World!\n",
			"WikiPage",
			"This is a test.\nI love tests.\n\nHello, World!\n",
			"This is a test.\nI love tests.\n\nHello, World!\n",
			"<p>\nThis is a test.<br />\nI love tests.\n</p>\n<p>\nHello, World!\n</p>\n",
		},
		{
			"heading 1",
			"WikiPage",
			"!!!!! Test !!!!!\n",
			"Test",
			"!!!!! Test !!!!!\n",
			"Test\n",
			"",
		},
		{
			"heading 2",
			"WikiPage",
			"!!!! Test !!!!\n",
			"WikiPage",
			"!!!! Test !!!!\n",
			"Test\n",
			"<h2><span class=\"markup\">!!!!</span> Test <span class=\"markup\">!!!!</span></h2>\n",
		},
		{
			"heading 3",
			"WikiPage",
			"!!! Test !!!\n",
			"WikiPage",
			"!!! Test !!!\n",
			"Test\n",
			"<h3><span class=\"markup\">!!!</span> Test <span class=\"markup\">!!!</span></h3>\n",
		},
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
