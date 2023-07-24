# go-tdlib
Golang Telegram Database library Wrapper

### Requirement:
- Build TDLib for Golang and your operating system. See a [TDLib build instructions generator](https://tdlib.github.io/td/build.html) for detailed instructions on how to build TDLib

### Install:
`go get -u github.com/aliforever/go-tdlib`

### Important:
The library is ported from [td_json_client_create](https://core.telegram.org/tdlib/docs/td__json__client_8h.html#a45cd6979ada11b7690d9dcb1ddc841a0) to [td_create_client_id](https://core.telegram.org/tdlib/docs/td__json__client_8h.html#a7feda953a66eee36bc207eb71a55c490). 
As the old interface "will be removed in TDLib 2.0.0"

In the new design there's a Manager struct that holds a map of channel for each client id. So you can have multiple clients in your application.

### Usage:
```go
package main

import (
  "github.com/aliforever/go-tdlib"
  "github.com/aliforever/go-tdlib/config"
  "fmt"
)

func main() {
  managerHandlers := tdlib.NewManagerHandlers(). 
    SetRawIncomingEventHandler(func(eventBytes []byte) {
        fmt.Println(string(eventBytes))
    })
  
  managerOptions := tdlib.NewManagerOptions().
    SetLogVerbosityLevel(6).
    SetLogPath("logs.txt")
  
  // Or you can pass nil for both handlers and options
  m := tdlib.NewManager(managerHandlers, managerOptions)
  
  // NewClientOptions
  cfg := config.New().
    SetFilesDirectory("./tdlib/tdlib-files").
    SetDatabaseDirectory("./tdlib/tdlib-db")

  h := tdlib.NewHandlers().
	  SetRawIncomingEventHandler(func(eventBytes []byte) {
	    fmt.Println(string(eventBytes))
	  })

  apiID := int64(123456)
  apiHash := "a1234b1234c1234d1234f345667"
  
  client := m.NewClient(apiID, apiHash, h, cfg, nil)

  err := client.ReceiveUpdates()
  if err != nil {
	  panic(err)
  }
}
```

### AuthorizationFlow for phone numbers:
1. Wait for `authorizationStateWaitTdlibParameters` and then call `SetTdlibParameters`
2. Wait for `authorizationStateWaitPhoneNumber` and then call `SetAuthenticationPhoneNumber`
3. Wait for `authorizationStateWaitCode` and then call `CheckAuthenticationCode`
4. (If Already registered and 2factor is set) Wait for `authorizationStateWaitPassword` and then call `CheckAuthenticationPassword`
5. (If not registered) Wait for `authorizationStateWaitRegistration` and then call `RegisterUser`

### AuthorizationFlow for bot tokens:
1. Wait for `authorizationStateWaitTdlibParameters` and then call `SetTdlibParameters`
2. Wait for `authorizationStateWaitPhoneNumber` and then call `CheckAuthenticationBotToken`

If the authorization flow is successful, the client will receive `updateAuthorizationState` with `authorizationStateReady` and then you can start using the client.

To register a handler for authorization state updates, pass your handler to `SetOnUpdateAuthorizationStateEventHandler` method of `Handlers` struct.
```go
func authorizationHandler(update entities.AuthorizationStateType) {
  switch update.AuthorizationState.GetAuthorizationStateEnum() {
  case entities.AuthorizationStateTypeAwaitingTdlibParameters:
    // ...
  case entities.AuthorizationStateTypeAwaitingPhoneNumber:
    // ...
  case entities.AuthorizationStateTypeAwaitingCode:
    // ...
  case entities.AuthorizationStateTypeAwaitingPassword:
    // ...
  case entities.AuthorizationStateTypeAwaitingRegistration:
    // ...
  case entities.AuthorizationStateTypeReady:
    // ...
  }
}
```

Or you can use `SetAuthorizationHandler` to register your custom interface which implements following:
```go
type AuthorizationHandler interface {
	Process(client *TDLib, update entities.AuthorizationStateType)
}
```
Checkout 2 pre-defined handlers in `authorizationhandler.go` file.
```
UserAuthorizationHandler
BotAuthorizationHandler
```

### Notes:
- Build project by specifying installation path for the built TDLib as following:

    `go build -ldflags="-r /usr/local/lib"`
- You can use `Dockerfile-Ubuntu` in the repo to build an image for ubuntu:20.04 and golang v1.19 and then use that to deploy your code