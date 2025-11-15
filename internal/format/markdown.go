package format

// TODO
func markdown(cfg formatConfig, title string, text string) (string, string, string, string) {
	return nomark(cfg, title, text) // fallback
}
