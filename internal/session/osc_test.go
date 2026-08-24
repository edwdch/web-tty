package session

import (
	"strings"
	"testing"
)

func TestOSCParserWindowTitle(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{
			name: "osc0 bel",
			in:   "\x1b]0;hello\x07",
			want: []string{"hello"},
		},
		{
			name: "osc2 st",
			in:   "\x1b]2;world\x1b\\",
			want: []string{"world"},
		},
		{
			name: "osc0 c1 st",
			in:   "\x1b]0;title\x9c",
			want: []string{"title"},
		},
		{
			name: "c1 osc",
			in:   "\x9d0;hi\x07",
			want: []string{"hi"},
		},
		{
			name: "icon name ignored",
			in:   "\x1b]1;icon\x07",
		},
		{
			name: "hyperlink ignored",
			in:   "\x1b]8;;https://example.com\x07",
		},
		{
			name: "empty clears",
			in:   "\x1b]0;\x07",
			want: []string{""},
		},
		{
			name: "semicolon in title",
			in:   "\x1b]0;a;b;c\x07",
			want: []string{"a;b;c"},
		},
		{
			name: "utf8",
			in:   "\x1b]0;进度 50%\x07",
			want: []string{"进度 50%"},
		},
		{
			name: "surrounding output",
			in:   "before\x1b]0;mid\x07after",
			want: []string{"mid"},
		},
		{
			name: "two titles",
			in:   "\x1b]0;one\x07\x1b]2;two\x07",
			want: []string{"one", "two"},
		},
		{
			name: "dcs does not fake osc",
			in:   "\x1bP \x1b]0;nope\x07 \x1b\\",
		},
		{
			name: "osc after dcs",
			in:   "\x1bPfoo\x1b\\\x1b]0;yes\x07",
			want: []string{"yes"},
		},
		{
			name: "csi then osc",
			in:   "\x1b[1;31m\x1b]0;red\x07",
			want: []string{"red"},
		},
		{
			name: "strips controls",
			in:   "\x1b]0;ok\x01there\x07",
			want: []string{"okthere"},
		},
		{
			name: "can cancels",
			in:   "\x1b]0;partial\x18\x1b]0;next\x07",
			want: []string{"next"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := newOSCParser()
			got := p.Feed([]byte(tc.in))
			if len(got) != len(tc.want) {
				t.Fatalf("got %#v want %#v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("got %#v want %#v", got, tc.want)
				}
			}
		})
	}
}

func TestOSCParserSplitAcrossFeeds(t *testing.T) {
	p := newOSCParser()
	if got := p.Feed([]byte("\x1b]0;hel")); len(got) != 0 {
		t.Fatalf("partial = %#v", got)
	}
	got := p.Feed([]byte("lo\x07"))
	if len(got) != 1 || got[0] != "hello" {
		t.Fatalf("got %#v", got)
	}
}

func TestOSCParserOverflowIgnored(t *testing.T) {
	p := newOSCParser()
	payload := "\x1b]0;" + strings.Repeat("a", maxOSCBytes+8) + "\x07"
	if got := p.Feed([]byte(payload)); len(got) != 0 {
		t.Fatalf("overflow = %#v", got)
	}
	got := p.Feed([]byte("\x1b]0;ok\x07"))
	if len(got) != 1 || got[0] != "ok" {
		t.Fatalf("after overflow %#v", got)
	}
}

func TestWindowTitle(t *testing.T) {
	if got := windowTitle("/bin/bash"); !strings.HasPrefix(got, "bash") {
		t.Fatalf("title = %q", got)
	}
	if got := windowTitle(""); !strings.HasPrefix(got, "bash") {
		t.Fatalf("empty shell title = %q", got)
	}
}
