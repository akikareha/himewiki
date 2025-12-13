package creole

import (
	"html/template"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

type blockMode int

const (
	blockNone blockMode = iota
	blockParagraph
	blockHorizon
	blockRaw
	blockCode
	blockMath
)

type decoMode int

const (
	decoNone decoMode = iota
	decoStrong
	decoEm
)

type state struct {
	config    formatConfig
	input     string
	index     int
	text      *strings.Builder
	plain     *strings.Builder
	html      *strings.Builder
	block     blockMode
	nextLine  int
	lineEnd   int
	prevLine  string
	line      string
	title     string
	outerDeco decoMode
	innerDeco decoMode
}

func closeDecos(s *state) {
	if s.innerDeco == decoStrong {
		s.html.WriteString("</strong>")
	} else if s.innerDeco == decoEm {
		s.html.WriteString("</em>")
	}
	s.innerDeco = decoNone

	if s.outerDeco == decoStrong {
		s.html.WriteString("</strong>")
	} else if s.outerDeco == decoEm {
		s.html.WriteString("</em>")
	}
	s.outerDeco = decoNone
}

func skipBlankLines(s *state) {
	for s.index < len(s.input) {
		end := s.index
		for end < len(s.input) {
			c := s.input[end]
			if c == '\r' || c == '\n' {
				end += 1
				for end < len(s.input) {
					c = s.input[end]
					if c != '\r' && c != '\n' {
						break
					}
					end += 1
				}
				break
			}
			if c != ' ' && c != '\t' {
				return
			}
			end += 1
		}
		s.index = end
	}
}

func closeBlock(s *state, nextBlock blockMode) {
	if s.block == blockParagraph {
		closeDecos(s)
		s.html.WriteString("</p>\n")
		if nextBlock != blockRaw && nextBlock != blockCode {
			skipBlankLines(s)
		}
	} else if s.block == blockHorizon {
		skipBlankLines(s)
	} else if s.block == blockRaw {
		s.html.WriteString("</code></pre>\n")
		skipBlankLines(s)
	} else if s.block == blockCode {
		s.html.WriteString("</code></pre>\n")
		skipBlankLines(s)
	} else if s.block == blockMath {
		s.text.WriteString("%%%\n\n")

		s.plain.WriteString("\n")

		s.html.WriteString("\\]</nomark-math>\n")
		s.html.WriteString("</div>\n")

		skipBlankLines(s)
	}

	s.block = blockNone
}

func openBlock(s *state, nextBlock blockMode) {
	if s.block != blockNone {
		panic("must be called in none block")
	}

	if nextBlock == blockParagraph {
		s.html.WriteString("<p>\n")
	} else if nextBlock == blockHorizon {
		// do nothing
	} else if nextBlock == blockRaw {
		s.html.WriteString("<pre><code>")
	} else if nextBlock == blockCode {
		s.html.WriteString("<pre><code>")
	} else if nextBlock == blockMath {
		s.text.WriteString("%%%\n")

		s.html.WriteString("<div>\n")
		s.html.WriteString("<nomark-math class=\"mathjax\">\\[")
	}

	s.block = nextBlock
}

func ensureBlock(s *state, block blockMode) {
	if s.block == block {
		return
	}
	closeBlock(s, block)
	openBlock(s, block)
}

func math(s *state) bool {
	line := s.input[s.index:s.lineEnd]
	if !strings.HasPrefix(line, "%%") {
		return false
	}
	end := strings.Index(line[2:], "%%")
	if end == -1 {
		return false
	}
	text := s.input[s.index+2 : s.index+2+end]
	html := template.HTMLEscapeString(text)

	s.text.WriteString("%%")
	s.text.WriteString(text)
	s.text.WriteString("%%")

	s.plain.WriteString(text)

	s.html.WriteString("<nomark-math class=\"mathjax\">\\(")
	s.html.WriteString(html)
	s.html.WriteString("\\)</nomark-math>")

	s.index += 2 + end + 2
	return true
}

func strong(s *state) bool {
	line := s.input[s.index:s.lineEnd]
	if !strings.HasPrefix(line, "**") {
		return false
	}

	if s.innerDeco == decoStrong {
		s.text.WriteString("**")
		s.html.WriteString("</strong>")
		s.innerDeco = decoNone
		s.index += 2
		return true
	}

	if s.outerDeco == decoStrong {
		if s.innerDeco == decoEm {
			s.html.WriteString("</em>")
			s.innerDeco = decoNone
		}
		s.text.WriteString("**")
		s.html.WriteString("</strong>")
		s.outerDeco = decoNone
		s.index += 2
		return true
	}

	s.text.WriteString("**")
	s.html.WriteString("<strong>")
	if s.outerDeco == decoNone {
		s.outerDeco = decoStrong
	} else {
		s.innerDeco = decoStrong
	}
	s.index += 2
	return true
}

func em(s *state) bool {
	line := s.input[s.index:s.lineEnd]
	if !strings.HasPrefix(line, "//") {
		return false
	}

	if s.innerDeco == decoEm {
		s.text.WriteString("//")
		s.html.WriteString("</em>")
		s.innerDeco = decoNone
		s.index += 2
		return true
	}

	if s.outerDeco == decoEm {
		if s.innerDeco == decoStrong {
			s.html.WriteString("</strong>")
			s.innerDeco = decoNone
		}
		s.text.WriteString("//")
		s.html.WriteString("</em>")
		s.outerDeco = decoNone
		s.index += 2
		return true
	}

	s.text.WriteString("//")
	s.html.WriteString("<em>")
	if s.outerDeco == decoNone {
		s.outerDeco = decoEm
	} else {
		s.innerDeco = decoEm
	}
	s.index += 2
	return true
}

func camel(s *state) bool {
	if s.index > 0 {
		c := s.input[s.index-1]
		if c >= 'A' && c <= 'Z' {
			return false
		}
		if c >= 'a' && c <= 'z' {
			return false
		}
		if c >= '0' && c <= '9' {
			return false
		}
	}
	line := s.input[s.index:s.lineEnd]
	if len(line) < 3 {
		return false
	}
	c := line[0]
	if c < 'A' || c > 'Z' {
		return false
	}
	c = line[1]
	if c < 'a' || c > 'z' {
		return false
	}
	i := 2
	upper := false
	for i < len(line) {
		c = line[i]
		if c < 'a' || c > 'z' {
			if c < 'A' || c > 'Z' {
				return false
			}
			i += 1
			upper = true
			break
		}
		i += 1
	}
	if !upper {
		return false
	}
	for i < len(line) {
		c = line[i]
		if c >= '0' && c <= '9' {
			i += 1
			for i < len(line) {
				c = line[i]
				if c >= 'A' && c <= 'Z' {
					return false
				}
				if c >= 'a' && c <= 'z' {
					return false
				}
				if c < '0' || c > '9' {
					break
				}
				i += 1
			}
			break
		}
		if upper {
			if c >= 'A' && c <= 'Z' {
				return false
			} else if c < 'a' || c > 'z' {
				break
			}
			upper = false
			i += 1
		} else {
			if c >= 'A' && c <= 'Z' {
				upper = true
			} else if c < 'a' || c > 'z' {
				break
			}
			i += 1
		}
	}
	name := line[:i]

	s.text.WriteString(name)

	s.plain.WriteString(name)

	s.html.WriteString("<a href=\"/")
	s.html.WriteString(url.PathEscape(name))
	s.html.WriteString("\" class=\"link\">")
	s.html.WriteString(template.HTMLEscapeString(name))
	s.html.WriteString("</a>")

	s.index += len(name)
	return true
}

func wikiLink(s *state) bool {
	line := s.input[s.index:s.lineEnd]
	if !strings.HasPrefix(line, "[[") {
		return false
	}
	ket := strings.Index(line[2:], "]]")
	if ket == -1 {
		return false
	}
	name := line[2 : 2+ket]

	s.text.WriteString("[[")
	s.text.WriteString(name)
	s.text.WriteString("]]")

	s.plain.WriteString(name)

	s.html.WriteString("<a href=\"/")
	s.html.WriteString(url.PathEscape(name))
	if strings.IndexByte(name, '.') != -1 {
		s.html.WriteString(".wiki")
	}
	s.html.WriteString("\" class=\"link\">")
	s.html.WriteString(template.HTMLEscapeString(name))
	s.html.WriteString("</a>")

	s.index += 2 + ket + 2
	return true
}

func nonURLIndex(line string) int {
	for i := 0; i < len(line); i += 1 {
		c := line[i]
		if c == '\r' || c == '\n' || c == ' ' || c == '\t' {
			return i
		}
	}
	return len(line)
}

func link(s *state) bool {
	line := s.input[s.index:s.lineEnd]
	if !strings.HasPrefix(line, "https:") {
		return false
	}
	end := nonURLIndex(line[6:])
	rawURL := line[:6+end]

	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" {
		return false
	}

	filename := path.Base(u.Path)
	ext := filepath.Ext(filename)
	if strings.HasPrefix(ext, ".") {
		ext = ext[1:]
	}
	extFound := false
	for _, extension := range s.config.image.extensions {
		if ext == extension {
			extFound = true
			break
		}
	}
	domainFound := false
	if extFound {
		for _, domain := range s.config.image.domains {
			if u.Host == domain {
				domainFound = true
				break
			}
		}
	}

	checked := u.String()

	s.text.WriteString(checked)

	s.plain.WriteString(checked)

	htmlURL := template.HTMLEscapeString(checked)
	if extFound && domainFound {
		s.html.WriteString("<img src=\"")
		s.html.WriteString(htmlURL)
		s.html.WriteString("\" alt=\"")
		s.html.WriteString(htmlURL)
		s.html.WriteString("\" />")
	} else {
		s.html.WriteString("<a href=\"")
		s.html.WriteString(htmlURL)
		s.html.WriteString("\" class=\"link\">")
		s.html.WriteString(htmlURL)
		s.html.WriteString("</a>")
	}

	s.index += len(rawURL)
	return true
}

func interLink(s *state) bool {
	line := s.input[s.index:s.lineEnd]
	for _, item := range s.config.links {
		if strings.HasPrefix(line, item.Key + ":") {
			end := nonURLIndex(line[len(item.Key) + 1:])
			rawURL := line[:len(item.Key) + 1 + end]

			_, err := url.Parse(rawURL[len(item.Key) + 1:])
			if err != nil {
				continue
			}

			s.text.WriteString(rawURL)

			s.plain.WriteString(rawURL)

			htmlURL := template.HTMLEscapeString(rawURL)
			htmlPath := template.HTMLEscapeString(rawURL[len(item.Key) + 1:])
			s.html.WriteString("<a href=\"")
			s.html.WriteString(item.URL)
			s.html.WriteString(htmlPath)
			s.html.WriteString("\" class=\"link\">")
			s.html.WriteString(htmlURL)
			s.html.WriteString("</a>")

			s.index += len(rawURL)
			return true
		}
	}
	return false
}

func html(s *state) bool {
	c := s.input[s.index]
	if c == '&' {
		s.index += 1
		s.text.WriteString("&")
		s.plain.WriteString("&")
		s.html.WriteString("&amp;")
		return true
	} else if c == '<' {
		s.index += 1
		s.text.WriteString("<")
		s.plain.WriteString("<")
		s.html.WriteString("&lt;")
		return true
	} else if c == '>' {
		s.index += 1
		s.text.WriteString(">")
		s.plain.WriteString(">")
		s.html.WriteString("&gt;")
		return true
	} else if c == '"' {
		s.index += 1
		s.text.WriteString("\"")
		s.plain.WriteString("\"")
		s.html.WriteString("&quot;")
		return true
	}
	return false
}

func raw(s *state) {
	r, size := utf8.DecodeRuneInString(s.input[s.index:])
	s.index += size
	s.text.WriteRune(r)
	s.plain.WriteRune(r)
	s.html.WriteRune(r)
}

func handleLine(s *state) {
	for s.index < s.lineEnd {
		if math(s) {
			continue
		} else if strong(s) {
			continue
		} else if em(s) {
			continue
		} else if interLink(s) {
			continue
		} else if camel(s) {
			continue
		} else if wikiLink(s) {
			continue
		} else if link(s) {
			continue
		} else if html(s) {
			continue
		} else {
			raw(s)
		}
	}
}

func isBlank(line string) bool {
	for _, r := range line {
		switch r {
		case ' ', '\t', '\r', '\n':
			// OK
		default:
			return false
		}
	}
	return true
}

func nextLine(s *state) {
	if s.nextLine < s.index {
		s.nextLine = s.index
		s.lineEnd = s.index
	}
	s.index = s.nextLine
	if s.nextLine >= len(s.input) {
		return
	}
	i := strings.IndexByte(s.input[s.nextLine:], '\n')
	if i == -1 {
		s.nextLine = len(s.input)
		s.lineEnd = len(s.input)
		return
	}
	start := s.nextLine
	end := s.nextLine + i
	s.nextLine = end + 1
	if end > 0 && s.input[end-1] == '\r' {
		end -= 1
	}
	nonBlank := end
	for nonBlank > 0 {
		c := s.input[nonBlank-1]
		if c != ' ' && c != '\t' {
			break
		}
		nonBlank -= 1
	}
	// in raw block preserve one white space
	if s.block == blockRaw {
		if nonBlank == start {
			c := s.input[nonBlank]
			if c == ' ' || c == '\t' {
				nonBlank += 1
			}
		}
	}
	s.lineEnd = nonBlank
}

const headingMaxLevel = 6

func parseHeading(s *state) (int, string, bool) {
	if s.block == blockRaw ||
		s.block == blockCode ||
		s.block == blockMath {
		return 0, "", false
	}
	if s.prevLine != "" {
		return 0, "", false
	}

	if strings.HasPrefix(s.line, "= ") &&
		strings.HasSuffix(s.line, " =") {
		return 1, s.line[2 : len(s.line)-2], true
	}
	if strings.HasPrefix(s.line, "== ") &&
		strings.HasSuffix(s.line, " ==") {
		return 2, s.line[3 : len(s.line)-3], true
	}
	if strings.HasPrefix(s.line, "=== ") &&
		strings.HasSuffix(s.line, " ===") {
		return 3, s.line[4 : len(s.line)-4], true
	}
	if strings.HasPrefix(s.line, "==== ") &&
		strings.HasSuffix(s.line, " ====") {
		return 4, s.line[5 : len(s.line)-5], true
	}
	if strings.HasPrefix(s.line, "===== ") &&
		strings.HasSuffix(s.line, " =====") {
		return 5, s.line[6 : len(s.line)-6], true
	}
	if strings.HasPrefix(s.line, "====== ") &&
		strings.HasSuffix(s.line, " ======") {
		return 6, s.line[7 : len(s.line)-7], true
	}

	return 0, "", false
}

func hasNextLine(s *state) bool {
	return s.nextLine < len(s.input)
}

func handleBlock(s *state) bool {
	// horizon
	if (s.block == blockNone || s.block == blockParagraph) &&
		strings.HasPrefix(s.line, "-") && strings.HasSuffix(s.line, "-") {
		ensureBlock(s, blockHorizon)

		s.text.WriteString(s.line)
		s.text.WriteString("\n")

		s.plain.WriteString(s.line)
		s.plain.WriteString("\n")

		s.html.WriteString("<hr />\n")

		ensureBlock(s, blockNone)
		nextLine(s)
		return true
	}

	// open raw block
	if s.block != blockRaw && s.block != blockCode &&
		s.block != blockMath && s.prevLine == "" &&
		strings.HasPrefix(s.line, " ") {
		ensureBlock(s, blockRaw)

		s.text.WriteString(s.line)
		s.text.WriteString("\n")

		s.plain.WriteString(s.line)
		s.plain.WriteString("\n")

		s.html.WriteString(template.HTMLEscapeString(s.line))
		s.html.WriteString("\n")

		nextLine(s)
		return true
	}

	// open code block
	if s.block != blockRaw && s.block != blockCode &&
		s.block != blockMath && s.line == "{{{" {
		s.text.WriteString("{{{\n")
		s.plain.WriteString("\n")
		ensureBlock(s, blockCode)
		nextLine(s)
		return true
	}

	// close paragraph block
	if s.block != blockRaw && s.block != blockCode &&
		s.block != blockMath && isBlank(s.line) {
		for s.index < len(s.input) {
			ensureBlock(s, blockNone)
			nextLine(s)
			if s.index >= len(s.input) {
				break
			}
			line := s.input[s.index:s.lineEnd]
			if !isBlank(line) {
				break
			}
		}
		return true
	}

	// close code block
	if s.block == blockCode && s.line == "}}}" {
		ensureBlock(s, blockNone)
		s.text.WriteString("}}}\n")
		s.plain.WriteString("\n")
		nextLine(s)
		return true
	}

	// open math block
	if s.block != blockMath && s.block != blockRaw &&
		s.block != blockCode && s.line == "%%%" {
		ensureBlock(s, blockMath)
		nextLine(s)
		return true
	}

	// close math block
	if s.block == blockMath && s.line == "%%%" {
		ensureBlock(s, blockNone)
		nextLine(s)
		return true
	}

	// headings
	if level, title, ok := parseHeading(s); ok {
		if !isBlank(s.prevLine) {
			s.text.WriteString("\n")
			s.plain.WriteString("\n")
		}

		s.text.WriteString(s.line)
		s.text.WriteString("\n")

		s.plain.WriteString(title)
		s.plain.WriteString("\n")

		if hasNextLine(s) {
			s.text.WriteString("\n")
			s.plain.WriteString("\n")
		}

		if level == 1 && s.title == "" {
			s.title = title
			nextLine(s)
		} else {
			ensureBlock(s, blockNone)
			titleHTML := template.HTMLEscapeString(title)
			levelStr := strconv.Itoa(level)
			s.html.WriteString("<h")
			s.html.WriteString(levelStr)
			s.html.WriteString(">")
			s.html.WriteString(titleHTML)
			s.html.WriteString("</h")
			s.html.WriteString(levelStr)
			s.html.WriteString(">\n")
			nextLine(s)
		}
		return true
	}

	// in raw block
	if s.block == blockRaw && strings.HasPrefix(s.line, " ") {
		s.text.WriteString(s.line)
		s.text.WriteString("\n")

		s.plain.WriteString(s.line)
		s.plain.WriteString("\n")

		s.html.WriteString(template.HTMLEscapeString(s.line))
		s.html.WriteString("\n")

		nextLine(s)
		return true
	}

	// in code block
	if s.block == blockCode {
		s.text.WriteString(s.line)
		s.text.WriteString("\n")

		s.plain.WriteString(s.line)
		s.plain.WriteString("\n")

		s.html.WriteString(template.HTMLEscapeString(s.line))
		s.html.WriteString("\n")

		nextLine(s)
		return true
	}

	// in math block
	if s.block == blockMath {
		s.text.WriteString(s.line)
		s.text.WriteString("\n")

		s.plain.WriteString(s.line)
		s.plain.WriteString("\n")

		s.html.WriteString(template.HTMLEscapeString(s.line))
		s.html.WriteString("\n")

		nextLine(s)
		return true
	}

	if s.line == "" {
		ensureBlock(s, blockNone)
		s.text.WriteString("\n")
		s.plain.WriteString("\n")
		nextLine(s)
		return true
	}

	return false
}

func trimTrailingBlanks(s *state) {
	// find succeeding blanks
	i := s.index
	for i < len(s.input) {
		c := s.input[i]
		if c != '\r' && c != '\n' && c != ' ' && c != '\t' {
			break
		}
		i += 1
	}
	// trim blanks if end of text was reached
	if i >= len(s.input) {
		s.index = i
	}
}

// Apply applies Nomark formatting on specified title and text.
func Apply(fc formatConfig, title string, text string) (
	string, string, string, string,
) {
	s := state{
		config:    fc,
		input:     text,
		index:     0,
		text:      new(strings.Builder),
		plain:     new(strings.Builder),
		html:      new(strings.Builder),
		block:     blockNone,
		nextLine:  0,
		lineEnd:   0,
		prevLine:  "",
		line:      "",
		title:     "",
		outerDeco: decoNone,
		innerDeco: decoNone,
	}

	// ignore beginning blank lines
	skipBlankLines(&s)

	// find first line
	nextLine(&s)

	// main for-each line loop
	for s.index < len(s.input) {
		// backup previous line
		s.prevLine = s.line
		// load current line
		s.line = s.input[s.index:s.lineEnd]

		// handle block structures
		if handleBlock(&s) {
			continue
		}

		// ensure normal paragraph is open
		ensureBlock(&s, blockParagraph)
		// convert starting spaces to non-breaking spaces
		// TODO handle tabs
		for s.index < s.lineEnd {
			c := s.input[s.index]
			if c != ' ' {
				break
			}
			s.text.WriteString(" ")
			s.plain.WriteString(" ")
			s.html.WriteString("&nbsp;")
			s.index += 1
		}
		// handle normal markups in line
		// note that nextLine might be called in handleLine
		handleLine(&s)

		// add LF to text and plain text buffers
		s.text.WriteString("\n")
		s.plain.WriteString("\n")

		// find next line
		nextLine(&s)

		// here after, range check is required
		// since s.index might be advanced

		// add blank lines to text and plain text buffers
		if s.index < len(s.input) {
			line := s.input[s.index:s.lineEnd]
			if isBlank(line) {
				s.text.WriteString("\n")
				s.plain.WriteString("\n")
			}
		}

		// add LF to HTML buffer
		if s.block == blockParagraph {
			s.html.WriteString("\n")
		}

		// trim trailing blank lines if exist
		trimTrailingBlanks(&s)
	}

	// main loop finished
	// ensure no blocks are open
	ensureBlock(&s, blockNone)

	// if no title is found from input text, use original title
	if s.title == "" {
		s.title = title
	}

	// return result
	return s.title, s.text.String(), s.plain.String(), s.html.String()
}
