package session

import (
	"bytes"
	"testing"
)

func TestParseHello(t *testing.T) {
	sz, err := ParseHello([]byte(`{"columns":80,"rows":24}`))
	if err != nil {
		t.Fatalf("ParseHello: %v", err)
	}
	if sz.Columns != 80 || sz.Rows != 24 {
		t.Fatalf("size = %+v", sz)
	}

	// Extra client fields must be ignored, not treated as commands.
	sz, err = ParseHello([]byte(`{"columns":10,"rows":10,"cmd":"rm -rf /","cwd":"/tmp","token":"x"}`))
	if err != nil {
		t.Fatalf("extra fields: %v", err)
	}
	if sz.Columns != 10 || sz.Rows != 10 {
		t.Fatalf("size = %+v", sz)
	}
}

func TestParseHelloRejects(t *testing.T) {
	cases := [][]byte{
		nil,
		[]byte("0{}"),
		[]byte(`{{"columns":80,"rows":24}`),
		[]byte(`{"columns":0,"rows":24}`),
		[]byte(`{"columns":80,"rows":0}`),
		[]byte(`{"columns":1001,"rows":24}`),
		[]byte(`{"cols":80,"rows":24}`),
		[]byte(`not json`),
	}
	for _, msg := range cases {
		if _, err := ParseHello(msg); err == nil {
			t.Fatalf("accepted %q", msg)
		}
	}
}

func TestParseResize(t *testing.T) {
	sz, err := ParseResize([]byte(`1{"columns":120,"rows":40}`))
	if err != nil {
		t.Fatalf("ParseResize: %v", err)
	}
	if sz.Columns != 120 || sz.Rows != 40 {
		t.Fatalf("size = %+v", sz)
	}
	if _, err := ParseResize([]byte(`{"columns":120,"rows":40}`)); err == nil {
		t.Fatal("expected missing command byte to fail")
	}
}

func TestOutputFrame(t *testing.T) {
	got := OutputFrame([]byte("hi"))
	if !bytes.Equal(got, []byte{'0', 'h', 'i'}) {
		t.Fatalf("got %q", got)
	}
	src := []byte("abc")
	got = OutputFrame(src)
	src[0] = 'z'
	if got[1] != 'a' {
		t.Fatal("OutputFrame must copy payload")
	}
}
