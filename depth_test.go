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

func TestNestingDepthLimit(t *testing.T) {
	at := strings.Repeat("[", maxObjectDepth) + strings.Repeat("]", maxObjectDepth)
	if readObjectErr(t, at, true) {
		t.Fatalf("%d nested arrays rejected", maxObjectDepth)
	}
	for name, input := range map[string]string{
		"arrays":       strings.Repeat("[", maxObjectDepth+1) + strings.Repeat("]", maxObjectDepth+1),
		"dictionaries": strings.Repeat("<</A ", maxObjectDepth+1) + "null" + strings.Repeat(">>", maxObjectDepth+1),
		"objdefs":      strings.Repeat("1 0 obj ", maxObjectDepth+1) + "null",
	} {
		if !readObjectErr(t, input, true) {
			t.Errorf("%s: nesting deeper than %d accepted", name, maxObjectDepth)
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
