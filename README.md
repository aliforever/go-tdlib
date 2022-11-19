# go-tdlib
Golang Telegram Database library Wrapper

### Requirement:
- Build TDLib for Golang and your operating system. See a [TDLib build instructions generator](https://tdlib.github.io/td/build.html) for detailed instructions on how to build TDLib

### Install:
`go get -u github.com/aliforever/go-tdlib`

### Usage:
```go
package main

import (
  "github.com/aliforever/go-tdlib"
  "github.com/aliforever/go-tdlib/config"
  "fmt"
)

func main() {
  cfg := config.New().
    SetLogPath("./tdlib/logs.txt").
    SetFilesDirectory("./tdlib/tdlib-files").
    SetDatabaseDirectory("./tdlib/tdlib-db")

  h := tdlib.NewHandlers().SetRawIncomingEventHandler(func(eventBytes []byte) {
	  fmt.Println(string(eventBytes))
  })

  apiID := 123456
  apiHash := "a1234b1234c1234d1234f345667"
  
  client := tdlib.NewClient(apiID, apiHash, h, cfg)

  err := c.tdlibClient.ReceiveUpdates()
  if err != nil {
	  panic(err)
  }
}
```

### Notes:
- Build project by specifying installation path for the built TDLib as following:

    `go build -ldflags="-r /usr/local/lib"`
- You can use `Dockerfile-Ubuntu` in the repo to build an image for ubuntu:20.04 and golang v1.19 and then use that to deploy your code