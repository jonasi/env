package env

import (
	"bytes"
	"io"
	"testing"
)

func TestWriter(t *testing.T) {
	w := NewSliceWriter()

	_, err := io.WriteString(w, `a
b
c
	`)

	if err != nil {
		t.Fatalf("Unexpected write error: %v", err)
	}

	d := w.Data()

	if len(d) != 3 {
		t.Errorf("Expected len %d,  found %d", 3, len(d))
	}

	if d[0] != "a" || d[1] != "b" || d[2] != "c" {
		t.Errorf("unexpected values %#v", d)
	}
}

func TestReader(t *testing.T) {
	r := NewSliceReader([]string{"a", "b", "c"})

	var b bytes.Buffer
	_, err := io.Copy(&b, r)

	if err != nil {
		t.Fatalf("Unexpected read error: %v", err)
	}

	if b.String() != `a
b
c` {
		t.Errorf("unexpected read value %s", b.String())
	}
}
