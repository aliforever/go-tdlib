# go-tdlib v2

Telegram TDLib client for Go, redesigned around Go 1.27 generic methods.

- **Requires Go 1.27+** — the typed API uses generic methods
  (`Client.Do[R]`, `Union[T].As[V]`, `Future[T].Map[U]`), which older
  toolchains do not support.
- **Requires TDLib (libtdjson) and cgo** — build with `-tags tdjson`.
  See the [TDLib build instructions generator](https://tdlib.github.io/td/build.html).
  Untagged builds compile and test the full typed API without cgo, but
  cannot connect to Telegram.
- Module path: `github.com/aliforever/go-tdlib/v2` (package `tdlib`).
  Migrating from v1? Read [MIGRATION.md](MIGRATION.md). v1 and v2 cannot
  run in the same binary — both own the process-wide TDLib receive loop.

```sh
go get github.com/aliforever/go-tdlib/v2
```

## Quick start

A minimal bot. Run it with
`TDLIB_API_ID=... TDLIB_API_HASH=... BOT_TOKEN=... go run -tags tdjson .`:

```go
//go:build tdjson

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	tdlib "github.com/aliforever/go-tdlib/v2"
	"github.com/aliforever/go-tdlib/v2/td"
)

func main() {
	apiID, _ := strconv.ParseInt(os.Getenv("TDLIB_API_ID"), 10, 32)
	apiHash := os.Getenv("TDLIB_API_HASH")
	token := os.Getenv("BOT_TOKEN")

	manager := tdlib.NewManager()
	client := manager.NewClient()

	// Authorization is a union dispatched with Match.
	client.On[*td.UpdateAuthorizationState](func(ctx context.Context, u *td.UpdateAuthorizationState) {
		u.AuthorizationState.Match(td.AuthorizationStateCases[error]{
			AuthorizationStateWaitTdlibParameters: func(*td.AuthorizationStateWaitTdlibParameters) error {
				_, err := client.Do(ctx, td.SetTdlibParameters{
					DatabaseDirectory:   "./tdlib-db",
					FilesDirectory:      "./tdlib-files",
					UseFileDatabase:     true,
					UseChatInfoDatabase: true,
					UseMessageDatabase:  true,
					ApiId:               int32(apiID),
					ApiHash:             apiHash,
					SystemLanguageCode:  "en",
					DeviceModel:         "my-bot",
					ApplicationVersion:  "1.0",
				})
				return err
			},
			AuthorizationStateWaitPhoneNumber: func(*td.AuthorizationStateWaitPhoneNumber) error {
				_, err := client.CheckAuthenticationBotToken(ctx, token)
				return err
			},
			AuthorizationStateReady: func(*td.AuthorizationStateReady) error {
				log.Println("authorized")
				return nil
			},
		})
	})

	// Echo every incoming text message.
	client.Messages().WithContent[*td.MessageText]().Handle(
		func(ctx context.Context, m *td.Message, c *td.MessageText) {
			if m.IsOutgoing {
				return
			}
			client.SendMessage(ctx, m.ChatId, td.MessageTopic{}, td.InputMessageReplyTo{}, nil, td.ReplyMarkup{},
				td.NewInputMessageContent(&td.InputMessageText{
					Text: &td.FormattedText{Text: "echo: " + c.Text.Text},
				}))
		})

	if err := manager.RunTDJSON(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
```

A complete bot + user example with phone/code/2FA flow lives in
[examples/sendmessage](examples/sendmessage).

## API tour

### Requests: `Do` and the facade

Every TDLib function is a generated struct in `td` whose type pins its
response, so `R` is inferred and a mismatch is a compile error:

```go
me, err := client.Do(ctx, td.GetMe{}) // me is *td.User

bad, err := client.Do[*td.Chat](ctx, td.GetMe{})
// compile error: td.GetMe implements td.Request[*td.User], not td.Request[*td.Chat]
```

All 1009 functions also have one-line facade methods:

```go
me, err := client.GetMe(ctx)
msg, err := client.SendMessage(ctx, chatID, td.MessageTopic{}, td.InputMessageReplyTo{}, nil, td.ReplyMarkup{}, content)
```

Fire-and-forget with a typed, lazily composable result:

```go
name, err := client.DoAsync(ctx, td.GetMe{}).
	Map(func(u *td.User) (string, error) { return u.FirstName, nil }).
	Await(ctx)
```

Escape hatch for functions the generated surface does not cover yet:

```go
raw, err := client.DoRaw(ctx, "getMe", nil)
```

### Unions: forward-compatible by construction

Every abstract TDLib class is a `td.Union[T]` container. Variants added by
newer TDLib versions decode into the `Unknown` arm with the raw JSON
retained — they are never silently dropped:

```go
switch msg.Content.Kind() {
case "messageText":
	mt, _ := msg.Content.As[*td.MessageText]()
	fmt.Println(mt.Text.Text)
}

if !msg.Content.Known() {
	log.Printf("unknown content %q: %s", msg.Content.Kind(), msg.Content.Raw())
}

// Visitor form:
msg.Content.Match(td.MessageContentCases[struct{}]{
	MessageText: func(mt *td.MessageText) struct{} { /* ... */ return struct{}{} },
	Unknown:     func(kind string, raw json.RawMessage) struct{} { return struct{}{} },
})
```

Construct a union value for a request with the generated wrapper:

```go
content := td.NewInputMessageContent(&td.InputMessageText{
	Text: &td.FormattedText{Text: "hi"},
})
```

### Updates: `On`, `Updates`, `Next`, filters, backpressure

The update variant is the type argument; the subscription key is its
`@type`:

```go
sub := client.On[*td.UpdateNewMessage](
	func(ctx context.Context, u *td.UpdateNewMessage) { /* ... */ },
	tdlib.Filter[*td.UpdateNewMessage](func(u *td.UpdateNewMessage) bool {
		return !u.Message.IsOutgoing
	}),
)
defer sub.Close()
```

Filters are plain functions and compose: `f.And(g)`, `f.Or(g)`, `f.Not()`.

Channel form with a bounded mailbox and explicit backpressure policy:

```go
ch, cancel := client.Updates[*td.UpdateNewMessage](ctx,
	tdlib.WithMailboxSize(1024),
	tdlib.WithBackpressure(tdlib.DropOldest), // DropNewest (default) | DropOldest | Block
	tdlib.WithDropWarning(func(n int64, tdType string) {
		log.Printf("dropped %d %s updates", n, tdType)
	}))
defer cancel()
for u := range ch { /* ... */ }
```

One-shot wait:

```go
u, err := client.Next[*td.UpdateNewMessage](ctx)
```

The message-stream builder narrows by chat, sender and content variant
before decoding:

```go
sub := client.Messages().
	InChat(chatID).
	FromUser(userID).
	WithContent[*td.MessageText]().
	Handle(func(ctx context.Context, m *td.Message, c *td.MessageText) { /* ... */ })
```

### Errors

`@type: "error"` responses decode into `*td.Error`:

```go
var tdErr *td.Error
if errors.As(err, &tdErr) {
	log.Printf("tdlib error %d: %s", tdErr.Code, tdErr.Message)
}
```

Timeouts are contexts: `ctx, cancel := context.WithTimeout(ctx, 5*time.Second)`.

## Architecture

- `td/` is generated from `td_api.tl`: entities, requests binding their
  response type, union containers, update variants.
- The runtime core (`Manager`, `Client`, pending registry, router) is
  hand-written and non-generic: one process-wide receive loop routes by
  `@client_id` into per-client bounded inboxes; no goroutine per update.
- The transport is an interface seam (`Transport.Send`,
  `Manager.Run(ctx, FetchFunc)`) because generic methods cannot be mocked
  through interfaces: tests drive the runtime with a fake `Transport` and
  a scripted pump, no cgo required.
- The libtdjson cgo binding lives behind `//go:build cgo && tdjson`
  (`internal/tdjson`, `manager_tdjson.go`), so untagged builds stay
  cgo-free.

## Regenerating `td`

The schema is vendored at `internal/gen/scheme/td_api.tl` — that file
pins the TDLib version the bindings target. To move to a newer TDLib,
replace it with the matching `td_api.tl` from the TDLib release and run:

```sh
cd v2 && go generate ./td
```

This rewrites `td/*_gen.go` and `facade_gen.go`. Never edit generated
files by hand.

## Pitfalls

- **Generic methods are not part of interfaces.** You cannot declare
  `interface { Do[R any](...) }`; program against `*tdlib.Client`, or
  define your own narrow non-generic interface over the facade methods you
  use. Tests should fake the `Transport` seam instead.
- **Method values need explicit instantiation.** `f := client.Do` does not
  compile; write `f := client.Do[*td.User]`.
- **`Match` is not exhaustive.** TDLib adds variants over time; always
  provide the `Unknown` arm (or check `Known()`) instead of assuming a
  closed set.
- **Empty unions marshal as `null`.** Passing a zero `td.ReplyMarkup{}`
  where TDLib expects "null if none" is correct; passing one where a value
  is required will fail server-side, not at compile time.
