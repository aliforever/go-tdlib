# go-tdlib
Golang Telegram Database library Wrapper

### Requirement:
- Build TDLib for Golang and your operating system. See a [TDLib build instructions generator](https://tdlib.github.io/td/build.html) for detailed instructions on how to build TDLib

### Install:
`go get -u github.com/aliforever/go-tdlib`

### Usage:
- TODO

### Notes:
- Build project by specifying installation path for the built TDLib as following:

    `go build -ldflags="-r /usr/local/lib"`
- You can use `Dockerfile-Ubuntu` in the repo to build an image for ubuntu:20.04 and golang v1.19 and then use that to deploy your code