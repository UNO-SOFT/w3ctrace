// Copyright 2026 Tamás Gulácsi.
//
// SPDX-License-Identifier: Apache-2.0

package w3ctrace_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/UNO-SOFT/w3ctrace"
)

func TestParseString(t *testing.T) {
	for nm, tC := range map[string]struct {
		In      string
		TraceID []byte
	}{
		"uuid": {In: "8be4df61-93ca-11d2-aa0d-00e098032b8c",
			TraceID: []byte("\x8b\xe4\xdfa\x93\xca\x11Ҫ\r\x00\xe0\x98\x03+\x8c")},
		"ulid": {In: "01ARZ3NDEKTSV4RRFFQ69G5FAV",
			TraceID: []byte("\x01V>:\xb5\xd3\xd6vLa\ufe53\x02\xbd[")},
	} {
		t.Run(nm, func(t *testing.T) {
			tr, err := w3ctrace.ParseString(tC.In)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%q: %q", tC.In, tr.TraceID[:])
			if got, wanted := tr.String(), fmt.Sprintf("00-%x-0000000000000000-00", tC.TraceID); got != wanted {
				t.Errorf("got %q wanted %q", got, wanted)
			}
		})
	}
}
func TestParseHeader(t *testing.T) {
	rnd := w3ctrace.New()
	for nm, want := range map[string]string{
		"fix":  "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		"rnd":  rnd.String(),
		"rnd2": rnd.Ensure().String(),
	} {
		t.Run(nm, func(t *testing.T) {
			tr, err := w3ctrace.ParseHeader(http.Header{"Traceparent": {want}})
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("trace: %s=%s=%#v", tr, tr.ShortString(), tr)

			tr2, err := w3ctrace.ParseHeader(http.Header{"Traceparent": {tr.String()}})
			if err != nil {
				t.Fatal(err)
			}
			if *tr2 != *tr {
				t.Fatalf("mismatch: tr1=%s != tr2=%s", tr, tr2)
			}
		})
	}
}
