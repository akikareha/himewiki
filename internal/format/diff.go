package format

import (
	"html/template"
	"strings"
)

func escapeIndentHTML(line string) string {
	var buf strings.Builder
	i := 0
	for i < len(line) && line[i] == ' ' {
		buf.WriteString("&nbsp;")
		i++
	}
	buf.WriteString(template.HTMLEscapeString(line[i:]))
	return buf.String()
}

// Diff formats diff text to HTML.
func Diff(text string) string {
	index := 0

	// skip headers
	lineNumber := 0
	for lineNumber < 2 && index < len(text) {
		_, index = indexLineEnd(text, index)
		lineNumber++
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
		if c == '+' {
			html.WriteString("<strong class=\"plus\">+</strong>")
			html.WriteString("<strong class=\"plus-line\">")
			html.WriteString(escapeIndentHTML(line[1:]))
			html.WriteString("</strong><br />\n")
		} else if c == '-' {
			html.WriteString("<em class=\"minus\">-</em>")
			html.WriteString("<em class=\"minus-line\">")
			html.WriteString(escapeIndentHTML(line[1:]))
			html.WriteString("</em><br />\n")
		} else if c == '@' {
			html.WriteString("<span class=\"hunk\">")
			html.WriteString(template.HTMLEscapeString(line))
			html.WriteString("</span><br />\n")
		} else if c == ' ' {
			html.WriteString(escapeIndentHTML(line))
			html.WriteString("<br />\n")
		} else {
			html.WriteString(template.HTMLEscapeString(line))
			html.WriteString("<br />\n")
		}
	}

	return html.String()
}
