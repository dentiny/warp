package generator

import (
	"fmt"
	"io"
)

// staticReader is an io.ReadSeeker that repeats a fixed byte pattern.
type staticReader struct {
	pattern []byte
	size    int64
	pos     int64
}

// Used to create a new static reader that repeats a pattern with repeating byte sequence (0x00, 0x01, 0x02, ...).
func newStaticReader(patternSize int) *staticReader {
	if patternSize <= 0 {
		patternSize = 128 << 10 // Default to 128KB
	}
	pattern := make([]byte, patternSize)
	for i := range pattern {
		pattern[i] = byte(i % 256)
	}
	return &staticReader{
		pattern: pattern,
		size:    0,
		pos:     0,
	}
}

// Resets the reader to the beginning and sets a new size limit.
func (s *staticReader) ResetSize(size int64) {
	s.size = size
	s.pos = 0
}

// Read reads data from the static pattern, repeating it as needed.
func (s *staticReader) Read(p []byte) (n int, err error) {
	if s.size <= 0 {
		return 0, io.EOF
	}
	if len(s.pattern) == 0 {
		return 0, io.EOF
	}

	remaining := s.size - s.pos
	if remaining <= 0 {
		return 0, io.EOF
	}

	toRead := len(p)
	if int64(toRead) > remaining {
		toRead = int(remaining)
	}

	patternPos := int(s.pos % int64(len(s.pattern)))
	for i := 0; i < toRead; i++ {
		p[i] = s.pattern[patternPos]
		patternPos = (patternPos + 1) % len(s.pattern)
		s.pos++
	}

	if s.pos >= s.size {
		return toRead, io.EOF
	}
	return toRead, nil
}

// Sets the offset for the next Read.
func (s *staticReader) Seek(offset int64, whence int) (int64, error) {
	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = s.pos + offset
	case io.SeekEnd:
		newPos = s.size + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}

	if newPos < 0 {
		return 0, fmt.Errorf("negative position: %d", newPos)
	}
	if newPos > s.size {
		newPos = s.size
	}

	s.pos = newPos
	return s.pos, nil
}
