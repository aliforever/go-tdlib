package td_test

import (
	"slices"
	"testing"

	"github.com/aliforever/go-tdlib/v2/td"
)

func TestLookupMethod(t *testing.T) {
	m, ok := td.LookupMethod("sendMessage")
	if !ok {
		t.Fatal("LookupMethod(sendMessage) not found")
	}
	if m.Name != "sendMessage" || m.ReturnType != "Message" {
		t.Fatalf("got %+v", m)
	}
	if m.Description == "" {
		t.Fatal("sendMessage has no description")
	}

	params := map[string]td.ParamSchema{}
	for _, p := range m.Params {
		params[p.Name] = p
	}
	for _, want := range []string{"chat_id", "topic_id", "input_message_content"} {
		if _, ok := params[want]; !ok {
			t.Fatalf("sendMessage missing param %s; got %v", want, m.Params)
		}
	}
	if got := params["chat_id"].GoType; got != "int64" {
		t.Fatalf("chat_id GoType = %q, want int64", got)
	}
	content := params["input_message_content"]
	if !content.IsUnion {
		t.Fatal("input_message_content IsUnion = false")
	}
	if !slices.Contains(content.Variants, "inputMessageText") {
		t.Fatalf("input_message_content Variants missing inputMessageText: %v", content.Variants)
	}

	if _, ok := td.LookupMethod("noSuchMethod"); ok {
		t.Fatal("LookupMethod(noSuchMethod) found")
	}
}

func TestAllMethods(t *testing.T) {
	all := td.AllMethods()
	if len(all) < 900 {
		t.Fatalf("AllMethods len = %d, want ~1000", len(all))
	}
	if !slices.IsSortedFunc(all, func(a, b td.MethodSchema) int {
		switch {
		case a.Name < b.Name:
			return -1
		case a.Name > b.Name:
			return 1
		}
		return 0
	}) {
		t.Fatal("AllMethods not sorted by name")
	}
}

func TestSearchMethods(t *testing.T) {
	// Multi-word query: every term must match name+description.
	res := td.SearchMethods("send photo", 0)
	if len(res) == 0 {
		t.Fatal(`SearchMethods("send photo") returned nothing`)
	}
	names := make([]string, len(res))
	for i, m := range res {
		names[i] = m.Name
	}
	if !slices.Contains(names, "sendMessageAlbum") {
		t.Fatalf(`SearchMethods("send photo") = %v, want sendMessageAlbum among results`, names)
	}
	if !slices.IsSorted(names) {
		t.Fatalf("SearchMethods results not sorted: %v", names)
	}

	// Case-insensitive matching.
	res = td.SearchMethods("SENDMESSAGE", 0)
	found := false
	for _, m := range res {
		if m.Name == "sendMessage" {
			found = true
		}
	}
	if !found {
		names := make([]string, len(res))
		for i, m := range res {
			names[i] = m.Name
		}
		t.Fatalf("SearchMethods(SENDMESSAGE) = %v, want sendMessage among results", names)
	}

	// The limit caps results.
	if res := td.SearchMethods("message", 3); len(res) != 3 {
		t.Fatalf("SearchMethods(message, 3) len = %d, want 3", len(res))
	}

	if res := td.SearchMethods("zzz-no-such-thing-zzz", 0); len(res) != 0 {
		t.Fatalf("expected no results, got %v", res)
	}
}
