package mapon

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
)

// DebugTransport is a [http.RoundTripper] that dumps requests and responses to stderr.
// The Enabled field supports lazy evaluation via a *bool, allowing it to be bound
// to a flag that is parsed after the transport is constructed.
type DebugTransport struct {
	Enabled *bool
	Next    http.RoundTripper
}

func (t *DebugTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if t.Enabled == nil || !*t.Enabled {
		return t.next().RoundTrip(request)
	}
	requestDump, err := httputil.DumpRequestOut(request, true)
	if err != nil {
		return nil, fmt.Errorf("failed to dump request for debug: %w", err)
	}
	prettyPrintDump(os.Stderr, requestDump, "> ")
	response, err := t.next().RoundTrip(request)
	if err != nil {
		return nil, err
	}
	responseDump, err := httputil.DumpResponse(response, true)
	if err != nil {
		return nil, fmt.Errorf("failed to dump response for debug: %w", err)
	}
	prettyPrintDump(os.Stderr, responseDump, "< ")
	return response, nil
}

func (t *DebugTransport) next() http.RoundTripper {
	if t.Next != nil {
		return t.Next
	}
	return http.DefaultTransport
}

func prettyPrintDump(w io.Writer, dump []byte, prefix string) {
	var output bytes.Buffer
	output.Grow(len(dump) * 2)
	for line := range bytes.Lines(dump) {
		output.WriteString(prefix)
		output.Write(line)
	}
	output.WriteByte('\n')
	_, _ = w.Write(output.Bytes())
}
