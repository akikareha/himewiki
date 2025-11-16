package format

import (
	"html/template"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type blockMode int

const (
	blockNone blockMode = iota
	blockParagraph
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
	input     string
	index     int
	text      *strings.Builder
	plain     *strings.Builder
	html      *strings.Builder
	block     blockMode
	nextLine  int
	lineEnd   int
	outerDeco decoMode
	innerDeco decoMode
	prevLine  string
	firstCode bool
	title     string
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

func skipLastBlanks(s *state) {
	i := s.index
	for i < len(s.input) {
		c := s.input[i]
		if c != '\r' && c != '\n' && c != ' ' && c != '\t' {
			break
		}
		i += 1
	}
	if i >= len(s.input) {
		s.index = i
	}
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

func closeBlock(s *state, nextBlock blockMode) {
	if s.block == blockNone {
		if nextBlock == blockRaw || nextBlock == blockCode {
			s.text.WriteString("\n")
		}
	} else if s.block == blockParagraph {
		s.html.WriteString("\n</p>")
		if nextBlock != blockRaw && nextBlock != blockCode {
			skipBlankLines(s)
		}
		if s.index < len(s.input) {
			if nextBlock != blockMath {
				s.text.WriteString("\n")
				s.plain.WriteString("\n")
			}
			s.html.WriteString("\n")
		}
	} else if s.block == blockRaw {
		s.html.WriteString("</code></pre>")
		if s.index < len(s.input) {
			c := s.input[s.index]
			if c != '\r' && c != '\n' {
				s.text.WriteString("\n")
			}
		}
		skipBlankLines(s)
		if s.index < len(s.input) {
			s.html.WriteString("\n")
		}
	} else if s.block == blockCode {
		s.html.WriteString("</code></pre>")
		skipBlankLines(s)
		if s.index < len(s.input) {
			s.html.WriteString("\n")
		}
	} else if s.block == blockMath {
		s.text.WriteString("%%%\n")
		s.html.WriteString("\n\\]</nomark-math>\n")
		s.html.WriteString("<span class=\"markup\">%%%</span>\n")
		s.html.WriteString("</div>")
		skipBlankLines(s)
		if s.index < len(s.input) {
			s.html.WriteString("\n")
		}
	}
	s.block = blockNone
}

func openBlock(s *state, nextBlock blockMode) {
	if nextBlock == blockNone {
		// do nothing
	} else if nextBlock == blockParagraph {
		s.text.WriteString("\n")
		s.plain.WriteString("\n")
		s.html.WriteString("<p>\n")
	} else if nextBlock == blockRaw {
		s.html.WriteString("<pre><code>")
	} else if nextBlock == blockCode {
		s.text.WriteString("\n")
		s.plain.WriteString("\n")
		s.html.WriteString("<pre><code>")
	} else if nextBlock == blockMath {
		s.text.WriteString("\n%%%\n")
		s.html.WriteString("<div>\n")
		s.html.WriteString("<span class=\"markup\">%%%</span>\n")
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

func math(fc formatConfig, s *state) bool {
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

	s.html.WriteString("<span class=\"markup\">%%</span>")
	s.html.WriteString("<nomark-math class=\"mathjax\">\\(")
	s.html.WriteString(html)
	s.html.WriteString("\\)</nomark-math>")
	s.html.WriteString("<span class=\"markup\">%%</span>")

	s.index += 2 + end + 2
	return true
}

func strong(fc formatConfig, s *state) bool {
	line := s.input[s.index:s.lineEnd]
	if !strings.HasPrefix(line, "**") {
		return false
	}

	if s.innerDeco == decoStrong {
		s.text.WriteString("**")
		s.html.WriteString("</strong><span class=\"markup\">**</span>")
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
		s.html.WriteString("</strong><span class=\"markup\">**</span>")
		s.outerDeco = decoNone
		s.index += 2
		return true
	}

	s.text.WriteString("**")
	s.html.WriteString("<span class=\"markup\">**</span><strong>")
	if s.outerDeco == decoNone {
		s.outerDeco = decoStrong
	} else {
		s.innerDeco = decoStrong
	}
	s.index += 2

	markup(fc, s)

	if s.innerDeco == decoStrong {
		s.html.WriteString("</strong>")
		s.innerDeco = decoNone
		return true
	}

	if s.outerDeco == decoStrong {
		if s.innerDeco == decoEm {
			s.html.WriteString("</em>")
			s.innerDeco = decoNone
		}
		s.html.WriteString("</strong>")
		s.outerDeco = decoNone
		return true
	}

	return true
}

func em(fc formatConfig, s *state) bool {
	line := s.input[s.index:s.lineEnd]
	if !strings.HasPrefix(line, "//") {
		return false
	}

	if s.innerDeco == decoEm {
		s.text.WriteString("//")
		s.html.WriteString("</em><span class=\"markup\">//</span>")
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
		s.html.WriteString("</em><span class=\"markup\">//</span>")
		s.outerDeco = decoNone
		s.index += 2
		return true
	}

	s.text.WriteString("//")
	s.html.WriteString("<span class=\"markup\">//</span><em>")
	if s.outerDeco == decoNone {
		s.outerDeco = decoEm
	} else {
		s.innerDeco = decoEm
	}
	s.index += 2

	markup(fc, s)

	if s.innerDeco == decoEm {
		s.html.WriteString("</em>")
		s.innerDeco = decoNone
		return true
	}

	if s.outerDeco == decoEm {
		if s.innerDeco == decoStrong {
			s.html.WriteString("</strong>")
			s.innerDeco = decoNone
		}
		s.html.WriteString("</em>")
		s.outerDeco = decoNone
		return true
	}

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

	s.html.WriteString("<span class=\"markup\">[[</span><a href=\"/")
	s.html.WriteString(url.PathEscape(name))
	if strings.IndexByte(name, '.') != -1 {
		s.html.WriteString(".wiki")
	}
	s.html.WriteString("\" class=\"link\">")
	s.html.WriteString(template.HTMLEscapeString(name))
	s.html.WriteString("</a><span class=\"markup\">]]</span>")

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

func link(fc formatConfig, s *state) bool {
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
	for _, extension := range fc.image.extensions {
		if ext == extension {
			extFound = true
			break
		}
	}
	domainFound := false
	if extFound {
		for _, domain := range fc.image.domains {
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

func raw(s *state) bool {
	r, size := utf8.DecodeRuneInString(s.input[s.index:])
	s.index += size
	s.text.WriteRune(r)
	s.plain.WriteRune(r)
	s.html.WriteRune(r)
	return true
}

func isBlank(line string) bool {
	for _, r := range line {
		if r != ' ' && r != '\t' {
			return false
		}
		if r != '\r' && r != '\n' {
			return false
		}
	}
	return true
}

var headingRe = regexp.MustCompile("^!!!(!*) (.+?) !!!(!*)$")

const headingMaxLevel = 3

func parseHeading(s *state, line string) (int, string, bool) {
	if s.prevLine != "" {
		return 0, "", false
	}

	matches := headingRe.FindStringSubmatch(line)
	if matches == nil {
		return 0, "", false
	}

	prefix := matches[1]
	suffix := matches[3]
	if len(prefix) != len(suffix) {
		return 0, "", false
	}
	level := headingMaxLevel - len(prefix)
	if level < 1 {
		return 0, "", false
	}

	title := matches[2]
	return level, title, true
}

func markup(fc formatConfig, s *state) {
	for s.index < s.lineEnd {
		if math(fc, s) {
			continue
		} else if strong(fc, s) {
			continue
		} else if em(fc, s) {
			continue
		} else if camel(s) {
			continue
		} else if wikiLink(s) {
			continue
		} else if link(fc, s) {
			continue
		} else if html(s) {
			continue
		} else if raw(s) {
			continue
		} else {
			panic("program error")
		}
	}
}

func nomarkLine(fc formatConfig, s *state) {
	line := s.input[s.index:s.lineEnd]

	// in raw block
	if s.block == blockRaw && strings.HasPrefix(line, " ") {
		s.text.WriteString("\n")
		s.text.WriteString(line)

		s.html.WriteString("\n")
		s.html.WriteString(template.HTMLEscapeString(line))

		nextLine(s)
		return
	}

	// in code block
	if s.block == blockCode {
		if s.firstCode {
			s.firstCode = false
		} else {
			s.text.WriteString("\n")
			s.html.WriteString("\n")
		}
		s.text.WriteString(line)
		s.html.WriteString(template.HTMLEscapeString(line))
		nextLine(s)
		return
	}

	// in math block
	if s.block == blockMath {
		s.text.WriteString(line)
		s.text.WriteString("\n")

		s.html.WriteString("\n")
		s.html.WriteString(template.HTMLEscapeString(line))

		nextLine(s)
		return
	}

	if line == "{{{" {
		s.text.WriteString("{{{")
		s.html.WriteString("<span class=\"markup\">{{{</span>")
		nextLine(s)
		return
	} else if line == "}}}" {
		s.text.WriteString("}}}")
		s.html.WriteString("<span class=\"markup\">}}}</span>")
		nextLine(s)
		return
	}

	if line == "" {
		ensureBlock(s, blockNone)
		s.text.WriteString("\n")
		nextLine(s)
		return
	}

	prevBlock := s.block
	ensureBlock(s, blockParagraph)
	for s.index < s.lineEnd {
		c := s.input[s.index]
		if c != ' ' {
			break
		}
		s.text.WriteString(" ")
		s.html.WriteString("&nbsp;")
		s.index += 1
	}
	markup(fc, s)
	nextLine(s)

	if s.index < len(s.input) {
		line = s.input[s.index:s.lineEnd]
		if !isBlank(line) && line != "%%%" {
			s.text.WriteString("\n")
			if prevBlock != blockRaw {
				s.html.WriteString("<br />\n")
			}
		}
	}

	skipLastBlanks(s)
}

func nomark(fc formatConfig, title string, text string) (string, string, string, string) {
	s := state{
		input:     text,
		index:     0,
		text:      new(strings.Builder),
		plain:     new(strings.Builder),
		html:      new(strings.Builder),
		block:     blockNone,
		nextLine:  0,
		lineEnd:   0,
		outerDeco: decoNone,
		innerDeco: decoNone,
		prevLine:  text[0:0],
		firstCode: false,
		title:     "",
	}

	skipBlankLines(&s)
	nextLine(&s)
	var line string
	for s.index < len(s.input) {
		s.prevLine = line
		line = s.input[s.index:s.lineEnd]

		// open raw block
		if s.block != blockRaw && s.block != blockCode && s.block != blockMath && s.prevLine == "" && strings.HasPrefix(line, " ") {
			ensureBlock(&s, blockRaw)
			s.text.WriteString(line)
			s.html.WriteString(template.HTMLEscapeString(line))
			nextLine(&s)
			continue
		}

		// open code block
		if s.block != blockRaw && s.block != blockCode && s.block != blockMath && s.prevLine == "{{{" && isBlank(line) {
			ensureBlock(&s, blockCode)
			nextLine(&s)
			s.firstCode = true
			continue
		}

		if s.block != blockRaw && s.block != blockCode && s.block != blockMath && isBlank(line) {
			for s.index < len(s.input) {
				closeBlock(&s, blockParagraph)
				nextLine(&s)
				if s.index >= len(s.input) {
					break
				}
				line := s.input[s.index:s.lineEnd]
				if !isBlank(line) {
					break
				}
			}
			continue
		}

		// close code block
		if s.block == blockCode && s.prevLine == "" && line == "}}}" {
			closeBlock(&s, blockParagraph)
			ensureBlock(&s, blockParagraph)
			nomarkLine(fc, &s)
			continue
		}

		// open math block
		if s.block != blockMath && s.block != blockRaw && s.block != blockCode && line == "%%%" {
			ensureBlock(&s, blockMath)
			nextLine(&s)
			continue
		}

		// close math block
		if s.block == blockMath && line == "%%%" {
			closeBlock(&s, blockParagraph)
			nextLine(&s)
			continue
		}

		// headings
		if level, title, ok := parseHeading(&s, line); ok {
			s.text.WriteString("\n")
			s.text.WriteString(line)
			s.text.WriteString("\n")

			s.plain.WriteString("\n")
			s.plain.WriteString(title)
			s.plain.WriteString("\n")
			if level == 1 && s.title == "" {
				s.title = title
				nextLine(&s)
			} else {
				closeBlock(&s, blockNone)
				titleHTML := template.HTMLEscapeString(title)
				levelStr := strconv.Itoa(level)
				var buf strings.Builder
				buf.WriteString("!!!")
				for i := 1; i <= headingMaxLevel-level; i++ {
					buf.WriteRune('!')
				}
				mark := buf.String()
				s.html.WriteString("<h")
				s.html.WriteString(levelStr)
				s.html.WriteString("><span class=\"markup\">")
				s.html.WriteString(mark)
				s.html.WriteString("</span> ")
				s.html.WriteString(titleHTML)
				s.html.WriteString(" <span class=\"markup\">")
				s.html.WriteString(mark)
				s.html.WriteString("</span></h")
				s.html.WriteString(levelStr)
				s.html.WriteString(">\n")
				nextLine(&s)
			}
			continue
		}

		nomarkLine(fc, &s)
	}
	closeBlock(&s, blockNone)

	if s.title != "" {
		title = s.title
	}
	return title, s.text.String(), s.plain.String(), s.html.String()
}
