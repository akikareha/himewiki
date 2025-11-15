package format

import "strings"

// indexLineEnd finds line end and start of next line
func indexLineEnd(text string, index int) (int, int) {
	lineFeed := strings.IndexByte(text[index:], '\n')
	if lineFeed == -1 {
		return len(text), len(text)
	}
	lineFeed += index
	lineEnd := lineFeed
	if lineEnd > 0 {
		if text[lineEnd-1] == '\r' {
			lineEnd -= 1
		}
	}
	return lineEnd, lineFeed + 1
}
