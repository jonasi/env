package env

import (
	"bytes"
	"io"
)

var split = []byte{'\n'}

func NewSliceWriter() *SliceWriter {
	return &SliceWriter{
		data: []string{},
		buf:  nil,
	}
}

type SliceWriter struct {
	data []string
	buf  []byte
}

func (s *SliceWriter) Write(b []byte) (int, error) {
	lines := bytes.SplitAfter(b, split)

	if len(lines) == 0 {
		return 0, nil
	}

	if len(s.buf) > 0 {
		lines[0] = append(s.buf, lines[0]...)
	}

	for _, line := range lines {
		if line[len(line)-1] != '\n' {
			s.buf = line
			break
		}

		drop := 1

		if len(line) > 2 && line[len(line)-2] == '\r' {
			drop = 2
		}

		s.data = append(s.data, string(line[:len(line)-drop]))
	}

	return len(b), nil
}

func (s *SliceWriter) Data() []string {
	d := make([]string, len(s.data), cap(s.data))
	copy(d, s.data)

	return d
}

func NewSliceReader(d []string) *SliceReader {
	var cp = make([]string, len(d))

	for i := range d {
		if i != len(d)-1 {
			cp[i] = d[i] + "\n"
		} else {
			cp[i] = d[i]
		}
	}

	return &SliceReader{
		data:  cp,
		slIdx: 0,
		idx:   0,
	}
}

type SliceReader struct {
	data  []string
	slIdx int
	idx   int
}

func (s *SliceReader) Read(b []byte) (int, error) {
	l := len(b)

	if l == 0 {
		return 0, nil
	}

	read := 0

	for {
		if s.slIdx >= len(s.data) {
			return read, io.EOF
		}

		cur := s.data[s.slIdx]

		if s.idx > len(cur) {
			return read, io.EOF
		}

		end := s.idx + l

		if end > len(cur) {
			end = len(cur)
		}

		read += copy(b[read:], cur[s.idx:end])
		s.idx = end

		if end == len(cur) {
			s.slIdx++
			s.idx = 0
		}

		if read >= l {
			break
		}
	}

	return read, nil
}
