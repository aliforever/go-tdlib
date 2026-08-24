package tdlib

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aliforever/go-tdlib/v2/td"
)

// Do executes one typed TDLib request and decodes its response into R. R is
// pinned by the request type, so it is almost always inferred:
//
//	user, err := client.Do(ctx, td.GetMe{})        // user is *td.User
//	ok, err := client.Do(ctx, td.Close{})          // ok is *td.Ok
//
// An explicit mismatched type argument fails to compile:
//
//	bad, err := client.Do[*td.Chat](ctx, td.GetMe{})
//	// compile error: td.GetMe implements td.Request[*td.User],
//	// not td.Request[*td.Chat]
//
// A response with @type "error" is returned as *td.Error (use errors.As);
// ctx cancellation aborts the wait and returns ctx.Err(). The generic layer
// is thin: marshal, roundTrip, decode — all correlation lives in roundTrip.
func (c *Client) Do[R any](ctx context.Context, r td.Request[R]) (R, error) {
	var zero R
	extra := NewExtra()
	payload, err := td.MarshalRequest(r, extra)
	if err != nil {
		return zero, fmt.Errorf("tdlib: marshal %s: %w", r.TDType(), err)
	}
	raw, err := c.roundTrip(ctx, payload, extra)
	if err != nil {
		return zero, err
	}
	// R conventions from the generator: pointer for concrete objects
	// (*td.User — new(R) is a **User and json.Unmarshal allocates), value
	// for union containers (td.MessageContent), scalars as-is. new(R) +
	// Unmarshal handles all three shapes uniformly.
	out := new(R)
	if err := json.Unmarshal(raw, out); err != nil {
		return zero, fmt.Errorf("tdlib: decode %s response: %w", r.TDType(), err)
	}
	return *out, nil
}

// DoAsync is Do in a goroutine, delivering into the returned Future.
// The ctx governs the in-flight request, not the Future: Await has its own.
func (c *Client) DoAsync[R any](ctx context.Context, r td.Request[R]) *Future[R] {
	f := newFuture[R]()
	go func() {
		v, err := c.Do(ctx, r)
		f.complete(v, err)
	}()
	return f
}

// DoRaw is the untyped escape hatch for requests the generated surface does
// not cover (newer TDLib functions, custom payloads). fields must marshal to
// a JSON object (a struct or map; nil means no fields); "@type" and "@extra"
// are injected. The raw response is returned undecoded — @type "error"
// responses still surface as *td.Error. It is deliberately less safe than
// Do: no compile-time link between request and response.
func (c *Client) DoRaw(ctx context.Context, tdType string, fields any) (json.RawMessage, error) {
	if fields == nil {
		fields = struct{}{}
	}
	b, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("tdlib: marshal %s fields: %w", tdType, err)
	}
	extra := NewExtra()
	payload, err := injectEnvelope(b, tdType, extra)
	if err != nil {
		return nil, err
	}
	return c.roundTrip(ctx, payload, extra)
}

// injectEnvelope inserts "@type" and "@extra" as the first keys of a JSON
// object, in a single pass. Local equivalent of the td package's unexported
// envelope helper; unlike it, a non-object input is an error here because
// DoRaw callers control the payload shape.
func injectEnvelope(b []byte, tdType, extra string) ([]byte, error) {
	if len(b) == 0 || b[0] != '{' {
		return nil, fmt.Errorf("tdlib: DoRaw fields for %s must marshal to a JSON object", tdType)
	}
	quote := func(s string) string {
		q, _ := json.Marshal(s) // identifiers and hex extras never fail
		return string(q)
	}
	prefix := `{"@type":` + quote(tdType) + `,"@extra":` + quote(extra)
	if len(b) == 2 { // "{}": no trailing comma needed
		return []byte(prefix + "}"), nil
	}
	out := make([]byte, 0, len(b)+len(prefix))
	out = append(out, prefix...)
	out = append(out, ',')
	out = append(out, b[1:]...)
	return out, nil
}
