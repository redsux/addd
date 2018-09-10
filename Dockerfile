FROM golang:1.10.3 as builder

ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /go/src/github.com/redsux/addd

COPY . ./

RUN go get -u github.com/kardianos/govendor \
 && govendor sync \
 && govendor install -a -ldflags '-extldflags "-static"' +local

FROM scratch
COPY --from=builder /go/bin/addd /addd
EXPOSE 53/udp 1632/tcp 10001/udp 10001/tcp 10002/tcp
ENTRYPOINT ["/addd"]