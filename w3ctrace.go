// Copyright 2026 Tamás Gulácsi.
//
// SPDX-License-Identifier: Apache-2.0

// Package w3ctrace helps receiving and sending
//
// Traceparent: Version-TraceID-SpanID-Flags
//
// headers.
package w3ctrace

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/oklog/ulid/v2"
)

type (
	TraceID     [16]byte
	SpanID      [8]byte
	FlagVersion [1]byte

	Trace struct {
		TraceID        TraceID
		ParentID       SpanID
		Flags, Version FlagVersion
	}

	ctxTrace struct{}
)

func NewTraceID() TraceID {
	var t TraceID
	id := ulid.MustNew(ulid.Now(), ulid.DefaultEntropy())
	copy(t[:], id[:cap(t)])
	return t
}
func (t TraceID) String() string { return hex.EncodeToString(t[:]) }
func (t TraceID) IsZero() bool   { return t == TraceID{} }

func NewSpanID() SpanID         { var s SpanID; ulid.DefaultEntropy().Read(s[:]); return s }
func (s SpanID) String() string { return base64.StdEncoding.EncodeToString(s[:]) }
func (s SpanID) IsZero() bool   { return s == SpanID{} }

func NewVersion() FlagVersion        { return FlagVersion([]byte{0}) }
func NewFlag(value byte) FlagVersion { return FlagVersion([]byte{value}) }
func (f FlagVersion) String() string { return hex.EncodeToString(f[:]) }

func New() *Trace { return &Trace{TraceID: NewTraceID()} }
func (tr *Trace) String() string {
	if tr == nil || !tr.IsValid() {
		return ""
	}
	return fmt.Sprintf("%x-%x-%x-%x", tr.Version[:], tr.TraceID[:], tr.ParentID[:], tr.Flags[:])
}
func (tr *Trace) ShortString() string {
	if tr == nil || !tr.IsValid() {
		return ""
	}
	// 22 + 1 + 11
	dst := make([]byte, 0, 22+1+11)
	b64 := base64.RawURLEncoding
	dst = b64.AppendEncode(dst, tr.TraceID[:])
	dst = append(dst, '.')
	dst = b64.AppendEncode(dst, tr.ParentID[:])
	return string(dst)
}
func (tr *Trace) IsValid() bool {
	return tr != nil &&
		tr.Version == FlagVersion([]byte{0x00}) &&
		!tr.TraceID.IsZero()
}
func (tr *Trace) Ensure() *Trace {
	if tr.IsValid() {
		return tr
	}
	id := ulid.MustNew(ulid.Now(), ulid.DefaultEntropy())
	copy(tr.TraceID[:], id[:cap(tr.TraceID)])
	if tr.ParentID.IsZero() {
		ulid.DefaultEntropy().Read(tr.ParentID[:])
	}
	return tr
}

// ParseString parses version-traceid-spanid-flags format (00-hex-hex-01)
//
// See https://www.w3.org/TR/trace-context/#trace-context-http-headers-format
func ParseString(hdr string) (*Trace, error) {
	const wantParts = 4
	parts := strings.SplitN(hdr, "-", wantParts+1)
	if len(parts) != wantParts {
		return nil, fmt.Errorf("wanted %d parts, got %d (from %s)", wantParts, len(parts), hdr)
	}
	var tr Trace
	for _, x := range []struct {
		Name string
		Idx  int
		Dest []byte
	}{
		{"version", 0, tr.Version[:]},
		{"traceid", 1, tr.TraceID[:]},
		{"parentid", 2, tr.ParentID[:]},
		{"flags", 3, tr.Flags[:]},
	} {
		if len(parts[x.Idx]) != cap(x.Dest)*2 {
			return nil, fmt.Errorf("%s must be %d hex, got %d (from %s)", x.Name, cap(x.Dest)*2, len(parts[x.Idx]), hdr)
		}
		if n, err := hex.Decode(x.Dest, []byte(parts[x.Idx])); err != nil {
			return nil, fmt.Errorf("parse %s as %s: %w", x.Name, parts[x.Idx], err)
		} else if n != cap(x.Dest) {
			return nil, fmt.Errorf("%s: parsed only %d bytes, wanted %d (from %s)", x.Name, n, cap(x.Dest), parts[x.Idx])
		}
	}
	return &tr, nil
}

// ParseHeader parses traceid header.
func ParseHeader(header http.Header) (*Trace, error) {
	return ParseString(header.Get("traceparent"))
}

// NewContext stores the trace into the context (iff it's valid).
func NewContext(ctx context.Context, tr *Trace) context.Context {
	if !tr.IsValid() {
		return ctx
	}
	return context.WithValue(ctx, ctxTrace{}, tr)
}

// FromContext return the trace from the context.
// It may be empty - use tr.Ensure() to generate a new trace if it is empty.
func FromContext(ctx context.Context) *Trace {
	tr, _ := ctx.Value(ctxTrace{}).(*Trace)
	return tr
}

// ExtractHTTP extracts Trace from the request headers.
// It may be empty - use tr.Ensure() to generate a new trace if it is empty.
func ExtractHTTP(req *http.Request) *Trace {
	tr, _ := ParseHeader(req.Header)
	return tr
}

// InjectHTTP injects the Traceparent header into the request (iff it is valid).
func InjectHTTP(req *http.Request, tr *Trace) {
	if !tr.IsValid() {
		return
	}
	req.Header.Set("Traceparent", tr.String())
}

func HTTPMiddleware(hndl http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tr, _ := ParseHeader(r.Header)
		tr.Ensure()
		r = r.WithContext(NewContext(r.Context(), tr))
		hndl.ServeHTTP(w, r)
	})
}
