# Migrating from v1 to v2

v2 is a rewrite around Go 1.27 generic methods. The `td` package is
generated from `td_api.tl`; the runtime core (`Manager`, `Client`,
`Transport`) is hand-written and non-generic. This document is the
mechanical rewrite table; `v2/README.md` has the full API tour.

Two hard requirements before anything compiles:

- **Go 1.27+** — generic methods (`Client.Do[R]`, `Union[T].As[V]`, ...) do
  not exist in older toolchains.
- **`-tags tdjson`** — the real TDLib wiring (`Manager.NewClient`,
  `Manager.RunTDJSON`) is behind the `tdjson` build tag and requires cgo +
  libtdjson. Untagged builds contain the full typed API and runtime, but
  cannot talk to libtdjson.

v1 and v2 **cannot coexist in one binary**: both own the process-wide
`td_receive` loop. Migrate a whole binary at once.

## Setup

```go
// v1
manager := tdlib.NewManager(context.Background(), nil, nil)
client := manager.NewClient(apiID, apiHash, handlers, config.New().
    SetDatabaseDirectory("./tdlib-db"), nil)
err := client.ReceiveUpdates(ctx)
```

```go
// v2 — build with -tags tdjson
manager := tdlib.NewManager()
client := manager.NewClient() // client ID comes from libtdjson
err := manager.RunTDJSON(ctx)
```

API credentials and directories moved out of the constructor: they are
parameters of `setTdlibParameters`, sent when TDLib asks for them (see
Authorization below).

## Requests

| v1 | v2 |
|---|---|
| `client.GetMe()` | `client.GetMe(ctx)` or `client.Do(ctx, td.GetMe{})` |
| `client.SendMessage(chatID, 0, nil, content, nil)` | `client.SendMessage(ctx, chatID, td.MessageTopic{}, td.InputMessageReplyTo{}, nil, td.ReplyMarkup{}, content)` or `client.Do(ctx, td.SendMessage{...})` |
| `client.SendMessage(..., tdlib.NewSendOptions(5*time.Second))` | `ctx, cancel := context.WithTimeout(ctx, 5*time.Second)` |
| string-built `OutgoingEvent` payloads | `client.DoRaw(ctx, "getMe", nil)` escape hatch |

In v1 the response type was a free type parameter of the package-level
`send`, so a mismatch compiled fine and failed at runtime:

```go
me, err := tdlib.Send[entities.Chat](client, outgoingevents.GetMe{}) // compiles, garbage at runtime
```

In v2 the request struct pins its response type, so the same mistake is a
compile error:

```go
bad, err := client.Do[*td.Chat](ctx, td.GetMe{})
// compile error: td.GetMe implements td.Request[*td.User],
// not td.Request[*td.Chat]
```

## Updates and handlers

v1 registered closures on a mutable `Handlers` struct with a fluent filter
builder:

```go
// v1
h := tdlib.NewHandlers().
    SetOnUpdateAuthorizationStateEventHandler(onAuthState).
    AddOnNewMessageHandler(onMessage, tdlib.NewMessageFilters().
        SetIsOutgoingFalse().
        SetIsTextTrue())
```

```go
// v2 — the update variant is the type argument; filters are plain funcs
sub := client.On[*td.UpdateNewMessage](
    func(ctx context.Context, u *td.UpdateNewMessage) { /* ... */ },
    func(u *td.UpdateNewMessage) bool { return !u.Message.IsOutgoing },
)

// or the narrowing stream builder for messages
sub := client.Messages().
    InChat(chatID).
    WithContent[*td.MessageText]().
    Handle(func(ctx context.Context, m *td.Message, c *td.MessageText) { /* ... */ })
```

Other consumption patterns with no v1 equivalent:

```go
updates, cancel := client.Updates[*td.UpdateNewMessage](ctx,
    tdlib.WithMailboxSize(1024),
    tdlib.WithBackpressure(tdlib.DropOldest))
defer cancel()

msg, err := client.Next[*td.UpdateNewMessage](ctx) // one-shot wait
```

Filters compose: `f.And(g)`, `f.Or(g)`, `f.Not()`.

## Union access

v1 used fat structs where every union member was a nullable field; v2
generates sealed union containers with `Kind()`, `As[V]()` and
`Match[R]()`:

```go
// v1
if message.Content.MessageText != nil {
    fmt.Println(message.Content.MessageText.Text.Text)
}
```

```go
// v2
if mt, ok := msg.Content.As[*td.MessageText](); ok {
    fmt.Println(mt.Text.Text)
}

// or the visitor form — unknown variants from newer TDLib versions fall
// through to Unknown instead of panicking
msg.Content.Match(td.MessageContentCases[struct{}]{
    MessageText: func(mt *td.MessageText) struct{} { /* ... */ return struct{}{} },
    Unknown:     func(kind string, raw json.RawMessage) struct{} { return struct{}{} },
})
```

Constructing a union value for a request:

```go
// v1
&entities.InputMessageText{Text: &entities.FormattedText{Text: "hi"}}

// v2
td.NewInputMessageContent(&td.InputMessageText{Text: &td.FormattedText{Text: "hi"}})
```

## Authorization

v1 hooked `entities.AuthorizationStateType` enum values; v2 delivers the
full `*td.UpdateAuthorizationState` and dispatches on the union:

```go
client.On[*td.UpdateAuthorizationState](func(ctx context.Context, u *td.UpdateAuthorizationState) {
    u.AuthorizationState.Match(td.AuthorizationStateCases[error]{
        AuthorizationStateWaitTdlibParameters: func(*td.AuthorizationStateWaitTdlibParameters) error {
            _, err := client.Do(ctx, td.SetTdlibParameters{ApiId: apiID, ApiHash: apiHash, /* ... */})
            return err
        },
        AuthorizationStateWaitPhoneNumber: func(*td.AuthorizationStateWaitPhoneNumber) error {
            _, err := client.CheckAuthenticationBotToken(ctx, token) // or SetAuthenticationPhoneNumber
            return err
        },
        AuthorizationStateWaitCode:     /* CheckAuthenticationCode */,
        AuthorizationStateWaitPassword: /* CheckAuthenticationPassword */,
        AuthorizationStateReady:        /* start working */,
    })
})
```

See `examples/sendmessage` for the complete flow. Note the state names
changed to the upstream schema spelling
(`authorizationStateWaitTdlibParameters`, not
`AuthorizationStateTypeAwaitingTdlibParameters`).

## Errors

v1 returned plain errors; matching on the message string was the only
option. v2 decodes `@type: "error"` responses into `*td.Error`:

```go
_, err := client.Do(ctx, td.GetMe{})
var tdErr *td.Error
if errors.As(err, &tdErr) {
    log.Printf("tdlib error %d: %s", tdErr.Code, tdErr.Message)
}
```

## Lifecycle

- v1 spawned a goroutine per handler; v2 runs one dispatch goroutine per
  client and one per subscriber, behind bounded mailboxes
  (drop-newest by default).
- `client.Close()` cancels every in-flight call with
  `tdlib.ErrClientClosed` and stops all subscriptions.
- There is no per-request timeout anymore; pass a `context.Context` with a
  deadline everywhere.
