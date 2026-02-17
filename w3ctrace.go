package w3ctrace

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

type (
	TraceID [16]byte
	SpanID  [8]byte
	Flags   [1]byte

	Trace struct {
		TraceID  TraceID
		ParentID SpanID
		Flags    Flags
	}
)

// PareHeader parses traceid header of version-traceid-spanid-flags format (00-hex-hex-01)
//
// See https://www.w3.org/TR/trace-context/#trace-context-http-headers-format
func ParseHeader(header http.Header) (Trace, error) {
	hdr := header.Get("traceparent")
	const wantParts = 4
	parts := strings.SplitN(hdr, "-", wantParts+1)
	if len(parts) != wantParts {
		return Trace{}, fmt.Errorf("wanted %d parts, got %d (from %s)", wantParts, len(parts), hdr)
	}
	if parts[0] != "00" {
		return Trace{}, fmt.Errorf("version must be 00, got %s (from %s)", parts[0], hdr)
	}
	var tr Trace
	for _, x := range []struct {
		Name string
		Idx  int
		Dest []byte
	}{
		{"traceid", 1, tr.TraceID[:]},
		{"parentid", 2, tr.ParentID[:]},
		{"flags", 3, tr.Flags[:]},
	} {
		if len(parts[x.Idx]) != cap(x.Dest)*2 {
			return tr, fmt.Errorf("%s must be %d hex, got %d (from %s)", x.Name, cap(x.Dest)*2, len(parts[x.Idx]), hdr)
		}
		if n, err := hex.Decode(x.Dest, []byte(parts[x.Idx])); err != nil {
			return tr, fmt.Errorf("parse %s as %s: %w", x.Name, parts[x.Idx], err)
		} else if n != cap(x.Dest) {
			return tr, fmt.Errorf("%s: parsed only %d bytes, wanted %d (from %s)", x.Name, n, cap(x.Dest), parts[x.Idx])
		}
	}
	return tr, nil
}
