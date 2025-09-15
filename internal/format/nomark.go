package format

import (
	"bytes"
	"html/template"
	"net/url"
	"path"
	"path/filepath"

	"github.com/akikareha/himewiki/internal/config"
)

type blockMode int

const (
	blockNone blockMode = iota
	blockParagraph
	blockRaw
)

type decoMode int

const (
	decoNone decoMode = iota
	decoStrong
	decoEm
)

type state struct {
	data []byte
	index int
	text *bytes.Buffer
	html *bytes.Buffer
	block blockMode
	nextLine int
	lineEnd int
	outerDeco decoMode
	innerDeco decoMode
	prevLine []byte
	firstRaw bool
}

func skip(s *state) {
	for s.index < len(s.data) {
		c := s.data[s.index]
		if c != '\r' && c != '\n' && c != ' ' && c != '\t' {
			break
		}
		s.index += 1
	}
}

func skipEnd(s *state) {
	i := s.index
	for i < len(s.data) {
		c := s.data[i]
		if c != '\r' && c != '\n' && c != ' ' && c != '\t' {
			break
		}
		i += 1
	}
	if i >= len(s.data) {
		s.index = i
	}
}

func nextLine(s *state) {
	if s.nextLine < s.index {
		s.nextLine = s.index
		s.lineEnd = s.index
	}
	s.index = s.nextLine
	if s.nextLine >= len(s.data) {
		return
	}
	i := bytes.IndexByte(s.data[s.nextLine:], '\n')
	if i == -1 {
		s.nextLine = len(s.data)
		s.lineEnd = len(s.data)
	} else {
		end := s.nextLine + i
		s.nextLine = end + 1
		if end > 0 && s.data[end - 1] == '\r' {
			end -= 1
		}
		for end > 0 {
			c := s.data[end - 1]
			if c != ' ' && c != '\t' {
				break
			}
			end -= 1
		}
		s.lineEnd = end
	}
}

func blockEnd(s *state, block blockMode) {
	if block == blockNone {
		s.text.WriteString("\n")
	}
	if s.block == blockParagraph {
		s.html.WriteString("\n</p>")
		if block != blockRaw {
			skip(s)
		}
		if s.index < len(s.data) {
			s.text.WriteString("\n")
			s.html.WriteString("\n")
		}
	} else if s.block == blockRaw {
		s.html.WriteString("</code></pre>")
		skip(s)
		if s.index < len(s.data) {
			s.html.WriteString("\n")
		}
	}
	s.block = blockNone
}

func blockBegin(s *state, block blockMode) {
	blockEnd(s, block)
	if block == blockParagraph {
		s.text.WriteString("\n")
		s.html.WriteString("<p>\n")
	} else if block == blockRaw {
		s.text.WriteString("\n")
		s.html.WriteString("<pre><code>")
	}
	s.block = block
}

func ignore(s *state) bool {
	c := s.data[s.index]
	if c == '\r' {
		s.index += 1
		return true
	}
	return false
}

func math(cfg *config.Config, s *state) bool {
	line := s.data[s.index:s.lineEnd]
	if !bytes.HasPrefix(line, []byte("%%")) {
		return false
	}
	//end := bytes.Index(s.data[s.index + 2:) // XXX unsafe
	end := bytes.Index(line[2:], []byte("%%"))
	if end == -1 {
		return false
	}
	text := s.data[s.index + 2:s.index + 2 + end]
	html := template.HTMLEscapeString(string(text))
	s.text.WriteString("%%" + string(text) + "%%")
	s.html.WriteString("<span class=\"markup\">%%</span>")
	s.html.WriteString("<nomark-math class=\"mathjax\">\\(" + html + "\\)</nomark-math>")
	s.html.WriteString("<span class=\"markup\">%%</span>")

	s.index += 2 + end + 2
	return true
}

func strong(cfg *config.Config, s *state) bool {
	line := s.data[s.index:s.lineEnd]
	if !bytes.HasPrefix(line, []byte("**")) {
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
	line := s.data[s.index:s.lineEnd]
	if !bytes.HasPrefix(line, []byte("//")) {
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
		c := s.data[s.index - 1]
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
	line := s.data[s.index:s.lineEnd]
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
		if c >= '0' && c <='9' {
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
	name := string(line[:i])

	s.text.WriteString(name)
	s.html.WriteString("<a href=\"/" + url.PathEscape(name) + "\" class=\"link\">" + template.HTMLEscapeString(name) + "</a>")

	s.index += len(name)
	return true
}

func wikiLink(s *state) bool {
	line := s.data[s.index:s.lineEnd]
	if !bytes.HasPrefix(line, []byte("[[")) {
		return false
	}
	ket := bytes.Index(line[2:], []byte("]]"))
	if ket == -1 {
		return false
	}
	name := string(line[2:2 + ket])

	s.text.WriteString("[[" + name + "]]")
	s.html.WriteString("<span class=\"markup\">[[</span><a href=\"/" + url.PathEscape(name) + "\" class=\"link\">" + template.HTMLEscapeString(name) + "</a><span class=\"markup\">]]</span>")

	s.index += 2 + ket + 2
	return true
}

func spaceIndex(b []byte) int {
	for i := 0; i < len(b) ; i += 1 {
		c := b[i]
		if c == '\r' || c == '\n' || c == ' ' || c == '\t' {
			return i
		}
	}
	return len(b)
}

func link(cfg *config.Config, s *state) bool {
	line := s.data[s.index:s.lineEnd]
	if !bytes.HasPrefix(line, []byte("https:")) {
		return false
	}
	space := spaceIndex(line[6:])
	rawURL := string(line[:6 + space])

	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" {
		return false
	}

	filename := path.Base(u.Path)
	ext := filepath.Ext(filename)
	extFound := false
	for _, extension := range cfg.Image.Extensions {
		if ext == "." + extension {
			extFound = true
			break
		}
	}
	if extFound {
		domainFound := false
		for _, domain := range cfg.Image.Domains {
			if u.Host == domain {
				domainFound = true
				break
			}
		}
		if !domainFound {
			return false
		}
	}

	checked := u.String()
	s.text.WriteString(checked)
	htmlURL := template.HTMLEscapeString(checked)
	if extFound {
		s.html.WriteString("<img src=\"" + htmlURL + "\" alt=\"" + htmlURL + "\" />")
	} else {
		s.html.WriteString("<a href=\"" + htmlURL + "\" class=\"link\">" + htmlURL + "</a>")
	}

	s.index += len(rawURL)
	return true
}

func html(s *state) bool {
	c := s.data[s.index]
	if c == '&' {
		s.index += 1
		s.text.WriteString("&")
		s.html.WriteString("&amp;")
		return true
	} else if c == '<' {
		s.index += 1
		s.text.WriteString("<")
		s.html.WriteString("&lt;")
		return true
	} else if c == '>' {
		s.index += 1
		s.text.WriteString(">")
		s.html.WriteString("&gt;")
		return true
	} else if c == '"' {
		s.index += 1
		s.text.WriteString("\"")
		s.html.WriteString("&quot;")
		return true
	}
	return false
}

func raw(s *state) bool {
	c := s.data[s.index]
	s.index += 1
	s.text.WriteByte(c)
	s.html.WriteByte(c)
	return true
}

func isBlank(b []byte) bool {
	for _, c := range b {
		if c != '\r' && c != '\n' && c != ' ' && c != '\t' {
			return false
		}
	}
	return true
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
	if s.block == blockRaw {
		line := string(s.data[s.index:s.lineEnd])
		s.text.WriteString(line + "\n")
		if s.firstRaw {
			s.html.WriteString(template.HTMLEscapeString(line))
			s.firstRaw = false
		} else {
			s.html.WriteString("\n" + template.HTMLEscapeString(line))
		}
		nextLine(s)
		return
	}

	{
		line := string(s.data[s.index:s.lineEnd])
		if line == "{{{" {
			s.text.WriteString("{{{")
			s.html.WriteString("<span class=\"markup\">{{{</span>")
			nextLine(s)
			return
		}
	}
	markup(cfg, s)
	nextLine(s)
	skipEnd(s)
	if s.index < len(s.data) {
		line := s.data[s.index:s.lineEnd]
		if !isBlank(line) {
			s.text.WriteString("\n")
			s.html.WriteString("<br />\n")
		}
	}
}

func Nomark(cfg *config.Config, text string) (string, string) {
	d := []byte(text)
	s := state {
		data: d,
		index: 0,
		text: new(bytes.Buffer),
		html: new(bytes.Buffer),
		block: blockNone,
		nextLine: 0,
		lineEnd: 0,
		outerDeco: decoNone,
		innerDeco: decoNone,
		prevLine: d[0:0],
		firstRaw: false,
	}

	skip(&s)
	blockBegin(&s, blockParagraph)
	nextLine(&s)
	for s.index < len(s.data) {
		line := s.data[s.index:s.lineEnd]
		if s.block != blockRaw && isBlank(line) {
			if string(s.prevLine) == "{{{" {
				blockEnd(&s, blockRaw)
				nextLine(&s)
				blockBegin(&s, blockRaw)
				s.firstRaw = true
				continue
			}

			blockEnd(&s, blockParagraph)
			if string(s.prevLine) == "----" {
				s.html.WriteString("<hr />\n")
			}
			blockBegin(&s, blockParagraph)
			for s.index < len(s.data) {
				nextLine(&s)
				if s.index >= len(s.data) {
					break
				}
				line := s.data[s.index:s.lineEnd]
				if !isBlank(line) {
					break
				}
			}
		} else if s.block == blockRaw && string(s.prevLine) == "" && string(line) == "}}}" {
			s.text.WriteString("}}}")
			blockEnd(&s, blockParagraph)
			nextLine(&s)
			blockBegin(&s, blockParagraph)
			s.html.WriteString("<span class=\"markup\">}}}</span><br />\n")
		} else {
			s.prevLine = line
			nomarkLine(cfg, &s)
		}
	}
	blockEnd(&s, blockNone)

	return s.text.String(), s.html.String()
}
