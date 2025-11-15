package format

// TODO
func creole(fc formatConfig, title string, text string) (string, string, string, string) {
	return nomark(fc, title, text) // fallback
}
