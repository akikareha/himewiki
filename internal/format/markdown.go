package format

// TODO
func markdown(fc formatConfig, title string, text string) (string, string, string, string) {
	return nomark(fc, title, text) // fallback
}
