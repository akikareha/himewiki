package format

import (
	"bytes"
	"html/template"
)

func detectLine(data []byte) (int, int) {
	lineFeed := bytes.IndexByte(data, '\n')
	if lineFeed == -1 {
		lineFeed = len(data)
	}
	lineEnd := lineFeed
	if lineEnd > 0 {
		c := data[lineEnd - 1]
		if c == '\r' {
			lineEnd -= 1
		}
	}
	nextLine := lineFeed + 1
	return lineEnd, nextLine
}

func Diff(text string) string {
	data := []byte(text)
	index := 0
	var html bytes.Buffer

	for index < len(data) {
		lineEnd, nextLine := detectLine(data[index:])
		lineEnd += index
		nextLine += index
		line := data[index:lineEnd]

		if len(line) < 1 {
			html.WriteString("\n");
		} else {
			c := line[0]
			htmlLine := template.HTMLEscapeString(string(line[1:]))
			if c == '+' {
				html.WriteString("<span class=\"plus\">+</span>")
				html.WriteString("<span class=\"plus-line\">" + htmlLine + "</span>\n")
			} else if c == '-' {
				html.WriteString("<span class=\"minus\">-</span>")
				html.WriteString("<span class=\"minus-line\">" + htmlLine + "</span>\n")
			} else {
				html.WriteString("&nbsp;")
				html.WriteString(htmlLine + "\n")
			}
		}
		index = nextLine
	}

	return html.String()
}
