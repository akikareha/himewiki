package format

import (
	"html/template"
	"strings"
)

// Diff formats diff text to HTML.
func Diff(text string) string {
	index := 0

	// skip headers
	lineNumber := 0
	for index < len(text) {
		_, index = indexLineEnd(text, index)

		lineNumber += 1
		if lineNumber >= 3 {
			break
		}
	}

	var html strings.Builder
	for index < len(text) {
		lineEnd, nextLine := indexLineEnd(text, index)
		line := text[index:lineEnd]
		index = nextLine

		if len(line) < 1 {
			html.WriteString("<br />\n")
			continue
		}

		c := line[0]
		htmlLine := template.HTMLEscapeString(string(line))
		if c == '+' {
			html.WriteString("<span class=\"plus\">+</span>")
			html.WriteString("<span class=\"plus-line\">")
			html.WriteString(htmlLine[1:])
			html.WriteString("</span><br />\n")
		} else if c == '-' {
			html.WriteString("<span class=\"minus\">-</span>")
			html.WriteString("<span class=\"minus-line\">")
			html.WriteString(htmlLine[1:])
			html.WriteString("</span><br />\n")
		} else if c == '@' {
			html.WriteString("<span class=\"hunk\">@")
			html.WriteString(htmlLine[1:])
			html.WriteString("</span><br />\n")
		} else if c == ' ' {
			html.WriteString("&nbsp;")
			html.WriteString(htmlLine[1:])
			html.WriteString("<br />\n")
		} else {
			html.WriteString(htmlLine)
			html.WriteString("<br />\n")
		}
	}

	return html.String()
}
