package providers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

// sharedTransport is reused by every provider HTTP client. net/http's default
// transport caps idle keep-alive connections at 2 per host (MaxIdleConnsPerHost),
// so under concurrency every call past the second to the same upstream (e.g.
// api.openai.com) pays for a fresh TCP + TLS handshake. All providers talk to a
// small set of hosts, so a single pooled transport with a generous idle pool
// lets those connections be reused across calls and across providers.
var sharedTransport = newSharedTransport()

func newSharedTransport() *http.Transport {
	// Clone the default transport to keep its dial/TLS timeouts and proxy
	// handling, then widen the idle-connection pool.
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxIdleConnsPerHost = 100
	t.IdleConnTimeout = 90 * time.Second
	return t
}

// NewHTTPClient returns a client that shares the pooled transport while keeping
// a per-client total-request timeout.
func NewHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: sharedTransport,
	}
}

var bufferPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

// EncodeJSON marshals v into a pooled buffer suitable for use as an HTTP request
// body and returns a release func that MUST be called once the request has been
// sent (i.e. after client.Do returns). Reusing buffers avoids a per-call heap
// allocation of the encoded request payload on the hot path.
//
// The returned buffer is safe to hand to http.NewRequest: net/http snapshots it
// into GetBody before Do sends the body, and GetBody is only consulted for
// retries/redirects during Do — never after release.
func EncodeJSON(v any) (*bytes.Buffer, func(), error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	if err := json.NewEncoder(buf).Encode(v); err != nil {
		bufferPool.Put(buf)
		return nil, nil, err
	}
	return buf, func() { bufferPool.Put(buf) }, nil
}

// DecodeJSONResponse decodes a JSON response body directly off the wire instead
// of buffering the whole payload with io.ReadAll first, then drains any trailing
// bytes so the keep-alive connection can be returned to the pool for reuse.
func DecodeJSONResponse(body io.Reader, v any) error {
	if err := json.NewDecoder(body).Decode(v); err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, body)
	return nil
}
