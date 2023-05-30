FROM alierror/go-tdlib-ubuntu:latest as builder

WORKDIR /sendmessage

COPY go.mod ./

RUN /usr/local/go/bin/go mod download -x

COPY . .

RUN /usr/local/go/bin/go build -o ./examples/sendmessage/sendmessage ./examples/sendmessage/.

CMD ["./examples/sendmessage/sendmessage"]