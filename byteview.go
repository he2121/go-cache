package go_cache

type ByteView struct {
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func cloneBytes(src []byte) []byte {
	dest := make([]byte, len(src))
	copy(dest, src)
	return dest
}

func (v ByteView) String() string {
	return string(v.b)
}
