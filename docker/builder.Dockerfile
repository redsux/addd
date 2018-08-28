FROM golang:1.10.3 as builder

ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /go/src/github.com/redsux/addd

COPY vendor ./vendor

RUN go get -u github.com/kardianos/govendor && govendor sync

COPY . ./

RUN govendor install -a -ldflags '-extldflags "-static"' +local