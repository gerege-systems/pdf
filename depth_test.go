package pdf

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"runtime/debug"
	"strings"
	"testing"
)

// readObjectPanic parses one object and returns the panic value, if any.
func readObjectPanic(input string) (p interface{}) {
	defer func() { p = recover() }()
	b := newBuffer(strings.NewReader(input), 0)
	b.allowEOF = true
	b.readObject()
	return nil
}

func TestNestingDepthLimit(t *testing.T) {
	nested := map[string]func(n int) string{
		"arrays":       func(n int) string { return strings.Repeat("[", n) + strings.Repeat("]", n) },
		"dictionaries": func(n int) string { return strings.Repeat("<</A ", n-1) + "null" + strings.Repeat(">>", n-1) },
		"objdefs":      func(n int) string { return strings.Repeat("1 0 obj ", n-1) + "null" + strings.Repeat(" endobj", n-1) },
	}
	for name, build := range nested {
		// n levels of readObject: the outermost call plus n-1 nested ones.
		if p := readObjectPanic(build(maxObjectDepth)); p != nil {
			t.Errorf("%s: %d levels rejected: %v", name, maxObjectDepth, p)
		}
		p := readObjectPanic(build(maxObjectDepth + 1))
		if err, ok := p.(error); !ok || !strings.Contains(err.Error(), "nested deeper than") {
			t.Errorf("%s: %d levels: got %v, want the depth error", name, maxObjectDepth+1, p)
		}
	}
}

// objStmPDF builds a PDF whose object 1 lives in object stream 2 (FlateDecode) with the given
// header ("objnum offset" pairs) and object data; /First is the header length.
func objStmPDF(header, body string) []byte {
	var z bytes.Buffer
	zw := zlib.NewWriter(&z)
	zw.Write([]byte(header + body))
	zw.Close()
	var b bytes.Buffer
	b.WriteString("%PDF-1.7\n")
	off2 := b.Len()
	fmt.Fprintf(&b, "2 0 obj\n<< /Type /ObjStm /N 1 /First %d /Filter /FlateDecode /Length %d >>\nstream\n", len(header), z.Len())
	b.Write(z.Bytes())
	b.WriteString("\nendstream\nendobj\n")
	off3 := b.Len()
	var x bytes.Buffer
	for _, e := range [][3]uint32{{0, 0, 0xffff}, {2, 2, 0}, {1, uint32(off2), 0}, {1, uint32(off3), 0}} {
		x.WriteByte(byte(e[0]))
		binary.Write(&x, binary.BigEndian, e[1])
		binary.Write(&x, binary.BigEndian, uint16(e[2]))
	}
	fmt.Fprintf(&b, "3 0 obj\n<< /Type /XRef /Size 4 /W [1 4 2] /Root 1 0 R /Length %d >>\nstream\n", x.Len())
	b.Write(x.Bytes())
	fmt.Fprintf(&b, "\nendstream\nendobj\nstartxref\n%d\n%%%%EOF\n", off3)
	return b.Bytes()
}

// TestDeepObjectStreamPanicsInsteadOfOverflowingStack resolves deeply nested objects stored in an
// object stream, reached through a non-integer offset (read as 0) or a nested "N G obj" chain.
// Before the depth limit these overflowed the stack: the lowered stack limit turns a regression
// into a fatal error of the test binary instead of a slow 1 GB overflow.
func TestDeepObjectStreamPanicsInsteadOfOverflowingStack(t *testing.T) {
	defer debug.SetMaxStack(debug.SetMaxStack(32 << 20))
	const n = 1 << 20
	for name, doc := range map[string][]byte{
		"non-integer offset": objStmPDF("1 2.5 ", strings.Repeat("[", n)),
		"keyword offset":     objStmPDF("1 [ ", strings.Repeat("[", n)),
		"objdef chain":       objStmPDF("1 0 ", strings.Repeat("1 0 obj ", n)+"null"),
		"dictionaries":       objStmPDF("1 0 ", strings.Repeat("<</A ", n)+"null"),
	} {
		t.Run(name, func(t *testing.T) {
			r, err := NewReader(bytes.NewReader(doc), int64(len(doc)))
			if err != nil {
				t.Fatal(err)
			}
			panicked := func() (p bool) {
				defer func() { p = recover() != nil }()
				r.Resolve(objptr{}, objptr{1, 0})
				return false
			}()
			if !panicked {
				t.Fatal("deeply nested object resolved without error")
			}
		})
	}
}
