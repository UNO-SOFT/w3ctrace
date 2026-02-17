package w3ctrace_test

import (
	"net/http"
	"testing"

	"github.com/UNO-SOFT/w3ctrace"
)

func TestParseHeader(t *testing.T) {
	tr, err := w3ctrace.ParseHeader(http.Header{"Traceparent": {"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"}})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("trace:", tr)
}
