package session

import (
	"bytes"
	"testing"
)

func TestByteRingWrapsAndKeepsNewest(t *testing.T) {
	r := newRing(8)
	r.Append([]byte("abcdef"))
	if got := r.Bytes(); !bytes.Equal(got, []byte("abcdef")) {
		t.Fatalf("got %q", got)
	}
	r.Append([]byte("ghij"))
	if got := r.Bytes(); !bytes.Equal(got, []byte("cdefghij")) {
		t.Fatalf("got %q", got)
	}
}

func TestByteRingOverwriteLargerThanCap(t *testing.T) {
	r := newRing(4)
	r.Append([]byte("abcdefgh"))
	if got := r.Bytes(); !bytes.Equal(got, []byte("efgh")) {
		t.Fatalf("got %q", got)
	}
}

func TestByteRingEmpty(t *testing.T) {
	r := newRing(4)
	if r.Bytes() != nil {
		t.Fatalf("got %q", r.Bytes())
	}
}
