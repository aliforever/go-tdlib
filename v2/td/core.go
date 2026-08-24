// Package td contains the TDLib type system: entities, requests, unions and
// updates. Most of it is generated from td_api.tl; this file is the
// hand-written core the generated code builds on.
package td

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
)

// ---------------------------------------------------------------------------
// Requests
// ---------------------------------------------------------------------------

// req is a zero-size phantom witness binding a request struct to its
// response type R. The response method is load-bearing: without it R could
// not be inferred at the call site of Client.Do.
type req[R any] struct{}

func (req[R]) response() *R { return nil }

// Req is embedded by every generated request struct.
type Req[R any] = req[R]

// Request is a TDLib function call whose response type is R. It is sealed:
// only generated code in this package can implement it.
type Request[R any] interface {
	TDType() string
	response() *R
}

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

// Error is the structured TDLib error object, returned by Client.Do when the
// response has @type "error".
type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string { return fmt.Sprintf("tdlib error %d: %s", e.Code, e.Message) }

// TDType implements a minimal typed envelope for the error object.
func (*Error) TDType() string { return "error" }

// ---------------------------------------------------------------------------
// Unions
// ---------------------------------------------------------------------------

// unionDecoder decodes a variant of one abstract TDLib class from raw JSON.
type unionDecoder func(raw json.RawMessage) (any, error)

var unionRegistry = struct {
	sync.RWMutex
	decoders map[reflect.Type]map[string]unionDecoder
}{decoders: map[reflect.Type]map[string]unionDecoder{}}

// RegisterUnion is called from generated code's init functions to associate
// the variant interface T with its @type discriminator table.
func RegisterUnion[T any](table map[string]unionDecoder) {
	var zero T
	key := reflect.TypeOf(&zero).Elem()
	unionRegistry.Lock()
	unionRegistry.decoders[key] = table
	unionRegistry.Unlock()
}

func unionTableFor[T any]() map[string]unionDecoder {
	var zero T
	key := reflect.TypeOf(&zero).Elem()
	unionRegistry.RLock()
	t := unionRegistry.decoders[key]
	unionRegistry.RUnlock()
	return t
}

// Union is the single representation for every TDLib abstract class
// (MessageContent, ChatMemberStatus, ...). T is the sealed variant interface.
// Unknown variants (added by newer TDLib versions) are surfaced via
// Known() == false with the raw JSON retained, never silently dropped.
type Union[T any] struct {
	value T
	kind  string
	raw   json.RawMessage
}

// NewUnion builds a known union value (used by generated code and tests).
func NewUnion[T any](kind string, value T) Union[T] {
	return Union[T]{value: value, kind: kind}
}

// Kind returns the @type discriminator, e.g. "messageText".
func (u Union[T]) Kind() string { return u.kind }

// Value returns the decoded variant, or the zero T for unknown variants.
func (u Union[T]) Value() T { return u.value }

// Known reports whether the variant is known to this build of the library.
func (u Union[T]) Known() bool {
	var zero T
	return any(u.value) != any(zero)
}

// Raw returns the original JSON of the union member.
func (u Union[T]) Raw() json.RawMessage { return u.raw }

// As extracts the concrete variant V. It requires Go 1.27 generic methods.
// A mismatched V simply returns ok == false.
func (u Union[T]) As[V any]() (V, bool) {
	v, ok := any(u.value).(V)
	return v, ok
}

func (u *Union[T]) UnmarshalJSON(data []byte) error {
	var head struct {
		Type string `json:"@type"`
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return fmt.Errorf("td: union unmarshal: %w", err)
	}
	u.kind = head.Type
	u.raw = append(u.raw[:0], data...)

	var zero T
	u.value = zero

	table := unionTableFor[T]()
	if table == nil {
		return fmt.Errorf("td: no union table registered for %T", zero)
	}
	dec, ok := table[head.Type]
	if !ok {
		return nil // unknown variant: surfaced via Known(), raw retained
	}
	v, err := dec(data)
	if err != nil {
		return fmt.Errorf("td: union variant %q: %w", head.Type, err)
	}
	val, ok := v.(T)
	if !ok {
		return fmt.Errorf("td: decoder for %q returned non-variant %T", head.Type, v)
	}
	u.value = val
	return nil
}

func (u Union[T]) MarshalJSON() ([]byte, error) {
	if !u.Known() {
		if u.raw != nil {
			return u.raw, nil
		}
		return []byte("null"), nil
	}
	return MarshalWithType(u.kind, u.value)
}

// ---------------------------------------------------------------------------
// JSON envelope helpers (single marshal, no marshal->map->remarshal)
// ---------------------------------------------------------------------------

// MarshalWithType marshals v and injects "@type" as the first key.
func MarshalWithType(tdType string, v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return injectType(b, tdType, ""), nil
}

// MarshalRequest marshals a request with "@type" and "@extra" injected.
func MarshalRequest[R any](r Request[R], extra string) ([]byte, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return injectType(b, r.TDType(), extra), nil
}

func injectType(b []byte, tdType, extra string) []byte {
	prefix := `{"@type":` + strconv_(tdType)
	if extra != "" {
		prefix += `,"@extra":` + strconv_(extra)
	}
	if len(b) >= 2 && b[0] == '{' && b[1] == '}' {
		return []byte(prefix + "}")
	}
	if len(b) > 0 && b[0] == '{' {
		out := make([]byte, 0, len(b)+len(prefix))
		out = append(out, prefix...)
		out = append(out, ',')
		out = append(out, b[1:]...)
		return out
	}
	return b
}

// strconv_ quotes a string as JSON without importing strconv in generated
// files; strings here are TDLib identifiers and UUIDs (no escaping needed in
// practice, but we stay correct).
func strconv_(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
