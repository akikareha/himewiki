package nomark

import "testing"

var mockCfg = formatConfig{
	image: imageConfig{
		domains:    []string{"example.org", "example.net"},
		extensions: []string{"png", "jpeg"},
	},
}

func TestApply(t *testing.T) {
	title, _, _, _ := Apply(mockCfg, "WikiPage", "Hello World")

	if title != "WikiPage" {
		t.Errorf("title = %s; want %s", title, "WikiPage")
	}
}
