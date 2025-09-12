package format

func Nomark(text string) string {
	// TODO
	return escapeHTML("(Nomark)\n" + text)
}
