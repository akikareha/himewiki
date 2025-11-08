package format

import (
	"html/template"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/akikareha/himewiki/internal/config"
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
	firstRaw  bool
	title     string
}

func skip(s *state) {
	for s.index < len(s.input) {
		end := s.index
		for end < len(s.input) {
			c := s.input[end]
			if c == '\r' || c == '\n' {
				end += 1
				break
			}
			if c != ' ' {
				return
			}
			end += 1
		}
		s.index = end
	}
}

func skipEnd(s *state) {
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
	} else {
		end := s.nextLine + i
		s.nextLine = end + 1
		if end > 0 && s.input[end-1] == '\r' {
			end -= 1
		}
		if s.block != blockRaw {
			for end > 0 {
				c := s.input[end-1]
				if c != ' ' && c != '\t' {
					break
				}
				end -= 1
			}
		}
		s.lineEnd = end
	}
}

func blockEnd(s *state, block blockMode) {
	if s.block == blockParagraph {
		s.html.WriteString("\n</p>")
		if block != blockRaw && block != blockCode {
			skip(s)
		}
		if s.index < len(s.input) {
			if block != blockMath {
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
		skip(s)
		if s.index < len(s.input) {
			s.html.WriteString("\n")
		}
	} else if s.block == blockCode {
		s.html.WriteString("</code></pre>")
		skip(s)
		if s.index < len(s.input) {
			s.html.WriteString("\n")
		}
	} else if s.block == blockMath {
		s.text.WriteString("%%%\n")
		s.html.WriteString("\n\\]</nomark-math>\n")
		s.html.WriteString("<span class=\"markup\">%%%</span>\n")
		s.html.WriteString("</div>")
		skip(s)
		if s.index < len(s.input) {
			s.html.WriteString("\n")
		}
	} else if s.block == blockNone {
		if block == blockRaw || block == blockCode {
			s.text.WriteString("\n")
		}
	}
	s.block = blockNone
}

func blockBegin(s *state, block blockMode) {
	if block == blockNone {
		// do nothing
	} else if block == blockParagraph {
		s.text.WriteString("\n")
		s.plain.WriteString("\n")
		s.html.WriteString("<p>\n")
	} else if block == blockRaw {
		s.html.WriteString("<pre><code>")
	} else if block == blockCode {
		s.text.WriteString("\n")
		s.plain.WriteString("\n")
		s.html.WriteString("<pre><code>")
	} else if block == blockMath {
		s.text.WriteString("\n%%%\n")
		s.html.WriteString("<div>\n")
		s.html.WriteString("<span class=\"markup\">%%%</span>\n")
		s.html.WriteString("<nomark-math class=\"mathjax\">\\[")
	}
	s.block = block
}

func checkBlock(s *state, block blockMode) {
	if s.block == block {
		return
	}
	blockEnd(s, block)
	blockBegin(s, block)
}

func ignore(s *state) bool {
	c := s.input[s.index]
	if c == '\r' {
		s.index += 1
		return true
	}
	return false
}

func math(cfg *config.Config, s *state) bool {
	line := s.input[s.index:s.lineEnd]
	if !strings.HasPrefix(line, "%%") {
		return false
	}
	//end := strings.Index(s.input[s.index + 2:]) // XXX unsafe
	end := strings.Index(line[2:], "%%")
	if end == -1 {
		return false
	}
	text := s.input[s.index+2 : s.index+2+end]
	html := template.HTMLEscapeString(text)
	s.text.WriteString("%%" + text + "%%")
	s.plain.WriteString(text)
	s.html.WriteString("<span class=\"markup\">%%</span>")
	s.html.WriteString("<nomark-math class=\"mathjax\">\\(" + html + "\\)</nomark-math>")
	s.html.WriteString("<span class=\"markup\">%%</span>")

	s.index += 2 + end + 2
	return true
}

func strong(cfg *config.Config, s *state) bool {
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

	markup(cfg, s)

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

func em(cfg *config.Config, s *state) bool {
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

	markup(cfg, s)

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
	for i < len(line) {
		c = line[i]
		if c < 'a' || c > 'z' {
			if c < 'A' || c > 'Z' {
				return false
			}
			i += 1
			break
		}
		i += 1
	}
	upper := true
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
	s.html.WriteString("<a href=\"/" + url.PathEscape(name) + "\" class=\"link\">" + template.HTMLEscapeString(name) + "</a>")

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

	s.text.WriteString("[[" + name + "]]")
	s.plain.WriteString(name)
	s.html.WriteString("<span class=\"markup\">[[</span><a href=\"/" + url.PathEscape(name) + "\" class=\"link\">" + template.HTMLEscapeString(name) + "</a><span class=\"markup\">]]</span>")

	s.index += 2 + ket + 2
	return true
}

func spaceIndex(line string) int {
	for i := 0; i < len(line); i += 1 {
		c := line[i]
		if c == '\r' || c == '\n' || c == ' ' || c == '\t' {
			return i
		}
	}
	return len(line)
}

func link(cfg *config.Config, s *state) bool {
	line := s.input[s.index:s.lineEnd]
	if !strings.HasPrefix(line, "https:") {
		return false
	}
	space := spaceIndex(line[6:])
	rawURL := line[:6+space]

	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" {
		return false
	}

	filename := path.Base(u.Path)
	ext := filepath.Ext(filename)
	extFound := false
	for _, extension := range cfg.Image.Extensions {
		if ext == "."+extension {
			extFound = true
			break
		}
	}
	domainFound := false
	if extFound {
		for _, domain := range cfg.Image.Domains {
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
		s.html.WriteString("<img src=\"" + htmlURL + "\" alt=\"" + htmlURL + "\" />")
	} else {
		s.html.WriteString("<a href=\"" + htmlURL + "\" class=\"link\">" + htmlURL + "</a>")
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
	c := s.input[s.index]
	s.index += 1
	s.text.WriteByte(c)
	s.plain.WriteByte(c)
	s.html.WriteByte(c)
	return true
}

func isBlank(line string) bool {
	for _, c := range line {
		if c != '\r' && c != '\n' && c != ' ' && c != '\t' {
			return false
		}
	}
	return true
}

var headingRe = regexp.MustCompile("^!!!(!*) (.+?) !!!(!*)$")

func parseHeading(s *state, line string) (level int, title string, ok bool) {
	if s.prevLine != "" {
		return 0, "", false
	}

	matches := headingRe.FindStringSubmatch(line)
	if matches == nil {
		return 0, "", false
	}

	prefix := matches[1]
	title = matches[2]
	suffix := matches[3]

	if len(prefix) != len(suffix) {
		return 0, "", false
	}

	level = 4 - len(prefix)

	if level < 1 || level > 4 {
		return 0, "", false
	}

	return level, title, true
}

func markup(cfg *config.Config, s *state) {
	for s.index < s.lineEnd {
		if ignore(s) {
			continue
		} else if math(cfg, s) {
			continue
		} else if strong(cfg, s) {
			continue
		} else if em(cfg, s) {
			continue
		} else if camel(s) {
			continue
		} else if wikiLink(s) {
			continue
		} else if link(cfg, s) {
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

func nomarkLine(cfg *config.Config, s *state) {
	if s.block != blockRaw && s.block != blockCode && s.block != blockMath {
		line := s.input[s.index:s.lineEnd]
		if strings.HasPrefix(line, " ") {
			blockEnd(s, blockRaw)
			blockBegin(s, blockRaw)
			s.text.WriteString(line)
			s.html.WriteString(template.HTMLEscapeString(line))
			nextLine(s)
			return
		}
	}

	if s.block == blockRaw {
		line := s.input[s.index:s.lineEnd]
		if strings.HasPrefix(line, " ") {
			s.text.WriteString("\n" + line)
			s.html.WriteString("\n" + template.HTMLEscapeString(line))
			nextLine(s)
			return
		}
	}

	if s.block == blockCode {
		line := s.input[s.index:s.lineEnd]
		if s.firstRaw {
			s.text.WriteString(line)
			s.html.WriteString(template.HTMLEscapeString(line))
			s.firstRaw = false
		} else {
			s.text.WriteString("\n" + line)
			s.html.WriteString("\n" + template.HTMLEscapeString(line))
		}
		nextLine(s)
		return
	}

	if s.block == blockMath {
		line := s.input[s.index:s.lineEnd]
		s.text.WriteString(line + "\n")
		s.html.WriteString("\n" + template.HTMLEscapeString(line))
		nextLine(s)
		return
	}

	{
		line := s.input[s.index:s.lineEnd]
		if line == "{{{" {
			s.text.WriteString("{{{")
			s.html.WriteString("<span class=\"markup\">{{{</span>")
			nextLine(s)
			return
		}
		if line == "}}}" {
			s.text.WriteString("}}}")
			s.html.WriteString("<span class=\"markup\">}}}</span>")
			nextLine(s)
			return
		}
	}

	prevBlock := s.block
	checkBlock(s, blockParagraph)
	markup(cfg, s)
	nextLine(s)
	skipEnd(s)
	if s.index < len(s.input) {
		line := s.input[s.index:s.lineEnd]
		if !isBlank(line) && line != "%%%" {
			s.text.WriteString("\n")
			if prevBlock != blockRaw {
				s.html.WriteString("<br />\n")
			}
		}
	}
}

func nomark(cfg *config.Config, title string, text string) (string, string, string, string) {
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
		firstRaw:  false,
		title:     "",
	}

	skip(&s)
	nextLine(&s)
	var line string
	for s.index < len(s.input) {
		s.prevLine = line
		line = s.input[s.index:s.lineEnd]
		if s.block != blockRaw && s.block != blockCode && s.block != blockMath && isBlank(line) {
			if s.prevLine == "{{{" {
				blockEnd(&s, blockCode)
				nextLine(&s)
				blockBegin(&s, blockCode)
				s.firstRaw = true
				continue
			}

			for s.index < len(s.input) {
				blockEnd(&s, blockParagraph)
				nextLine(&s)
				if s.index >= len(s.input) {
					break
				}
				line := s.input[s.index:s.lineEnd]
				if !isBlank(line) {
					break
				}
			}
		} else if s.block == blockCode && s.prevLine == "" && line == "}}}" {
			blockEnd(&s, blockParagraph)
			checkBlock(&s, blockParagraph)
			nomarkLine(cfg, &s)
		} else if s.block != blockMath && s.block != blockRaw && s.block != blockCode && line == "%%%" {
			blockEnd(&s, blockMath)
			nextLine(&s)
			blockBegin(&s, blockMath)
		} else if s.block == blockMath && line == "%%%" {
			blockEnd(&s, blockParagraph)
			nextLine(&s)
		} else if level, title, ok := parseHeading(&s, line); ok {
			s.text.WriteString("\n" + line + "\n")
			s.plain.WriteString("\n" + title + "\n")
			if level == 1 && s.title == "" {
				s.title = title
				nextLine(&s)
			} else {
				blockEnd(&s, blockNone)
				titleHTML := template.HTMLEscapeString(title)
				levelStr := strconv.Itoa(level)
				var buf strings.Builder
				buf.WriteString("!!!")
				for i := 0; i <= 3-level; i += 1 {
					buf.WriteRune('!')
				}
				mark := buf.String()
				s.html.WriteString("<h" + levelStr + ">" + "<span class=\"markup\">" + mark + "</span> " + titleHTML + " <span class=\"markup\">" + mark + "</span>" + "</h" + levelStr + ">\n")
				nextLine(&s)
			}
		} else {
			nomarkLine(cfg, &s)
		}
	}
	blockEnd(&s, blockNone)

	if s.title != "" {
		title = s.title
	}
	return title, s.text.String(), s.plain.String(), s.html.String()
}
