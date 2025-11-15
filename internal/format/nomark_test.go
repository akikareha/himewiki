package format

import "testing"

var mockCfg = formatConfig{
	image: imageConfig{
		domains: []string{"example.org", "example.net"},
		extensions: []string{"png", "jpeg"},
	},
}

func TestNomark(t *testing.T) {
	title, _, _, _ := nomark(mockCfg, "WikiPage", "Hello World")

	if title != "WikiPage" {
		t.Errorf("title = %s; want %s", title, "WikiPage")
	}
}
