package session

const defaultReplayBytes = 1 << 20

type byteRing struct {
	buf    []byte
	start  int
	length int
}

func newRing(n int) *byteRing {
	if n <= 0 {
		n = defaultReplayBytes
	}
	return &byteRing{buf: make([]byte, n)}
}

func (r *byteRing) Append(p []byte) {
	n := len(r.buf)
	if n == 0 || len(p) == 0 {
		return
	}
	if len(p) >= n {
		copy(r.buf, p[len(p)-n:])
		r.start = 0
		r.length = n
		return
	}
	overflow := r.length + len(p) - n
	if overflow > 0 {
		r.start = (r.start + overflow) % n
		r.length -= overflow
	}
	end := (r.start + r.length) % n
	first := len(p)
	if space := n - end; first > space {
		first = space
	}
	copy(r.buf[end:], p[:first])
	if first < len(p) {
		copy(r.buf, p[first:])
	}
	r.length += len(p)
}

func (r *byteRing) Bytes() []byte {
	if r.length == 0 {
		return nil
	}
	n := len(r.buf)
	out := make([]byte, r.length)
	first := r.length
	if space := n - r.start; first > space {
		first = space
	}
	copy(out, r.buf[r.start:r.start+first])
	if first < r.length {
		copy(out[first:], r.buf[:r.length-first])
	}
	return out
}
