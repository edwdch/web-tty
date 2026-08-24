package session

import (
	"bytes"
	"unicode/utf8"
)

// OSC 0 / 2 set the terminal window title (xterm). Shells and TUIs emit
// these sequences themselves; this is not a full snapshot of the process tree.
const maxOSCBytes = 4096

const (
	c0BEL = 0x07
	c0CAN = 0x18
	c0SUB = 0x1a
	c0ESC = 0x1b
	c1ST  = 0x9c
	c1OSC = 0x9d
	c1DCS = 0x90
	c1SOS = 0x98
	c1PM  = 0x9e
	c1APC = 0x9f
)

type oscState int

const (
	oscGround oscState = iota
	oscEsc
	oscIn
	oscInEsc
	oscString
	oscStringEsc
)

type oscParser struct {
	state    oscState
	buf      []byte
	overflow bool
}

func newOSCParser() oscParser {
	return oscParser{buf: make([]byte, 0, 256)}
}

// Feed consumes PTY bytes and returns completed window titles in order.
// An empty string means the program cleared the title (revert to the base).
func (p *oscParser) Feed(data []byte) []string {
	var titles []string
	for i := 0; i < len(data); {
		b := data[i]
		advance := true
		switch p.state {
		case oscGround:
			switch b {
			case c0ESC:
				p.state = oscEsc
			case c1OSC:
				p.startOSC()
			case c1DCS, c1SOS, c1PM, c1APC:
				p.state = oscString
			}
		case oscEsc:
			switch b {
			case ']':
				p.startOSC()
			case 'P', 'X', '^', '_':
				p.state = oscString
			case c0ESC:
				// stay in ESC
			default:
				p.state = oscGround
			}
		case oscIn:
			if title, ok, done := p.byteInOSC(b); done {
				if ok {
					titles = append(titles, title)
				}
			}
		case oscInEsc:
			if b == '\\' {
				if title, ok := p.endOSC(); ok {
					titles = append(titles, title)
				}
				p.state = oscGround
			} else {
				p.resetOSC()
				p.state = oscEsc
				advance = false
			}
		case oscString:
			switch b {
			case c0BEL, c1ST, c0CAN, c0SUB:
				p.state = oscGround
			case c0ESC:
				p.state = oscStringEsc
			}
		case oscStringEsc:
			if b == '\\' {
				p.state = oscGround
			} else {
				p.state = oscString
			}
		}
		if advance {
			i++
		}
	}
	return titles
}

func (p *oscParser) startOSC() {
	p.buf = p.buf[:0]
	p.overflow = false
	p.state = oscIn
}

func (p *oscParser) resetOSC() {
	p.buf = p.buf[:0]
	p.overflow = false
}

func (p *oscParser) byteInOSC(b byte) (title string, ok bool, done bool) {
	switch b {
	case c0BEL, c1ST:
		title, ok = p.endOSC()
		p.state = oscGround
		return title, ok, true
	case c0ESC:
		p.state = oscInEsc
		return "", false, false
	case c0CAN, c0SUB:
		p.resetOSC()
		p.state = oscGround
		return "", false, true
	}
	if !p.overflow {
		if len(p.buf) >= maxOSCBytes {
			p.overflow = true
		} else {
			p.buf = append(p.buf, b)
		}
	}
	return "", false, false
}

func (p *oscParser) endOSC() (string, bool) {
	defer p.resetOSC()
	if p.overflow {
		return "", false
	}
	return parseOSCTitle(p.buf)
}

func parseOSCTitle(payload []byte) (string, bool) {
	semi := bytes.IndexByte(payload, ';')
	if semi <= 0 {
		return "", false
	}
	ps := 0
	for _, c := range payload[:semi] {
		if c < '0' || c > '9' {
			return "", false
		}
		ps = ps*10 + int(c-'0')
		if ps > 99 {
			return "", false
		}
	}
	if ps != 0 && ps != 2 {
		return "", false
	}
	return sanitizeTitle(string(payload[semi+1:])), true
}

func sanitizeTitle(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		if r == utf8.RuneError && size == 1 {
			continue
		}
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) {
			continue
		}
		out = utf8.AppendRune(out, r)
	}
	return string(out)
}
