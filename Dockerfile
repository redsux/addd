FROM node:10.10 as jsbld

WORKDIR /home/node
RUN git clone https://github.com/redsux/addd-ui.git \
 && cd addd-ui \
 && npm install \
 && npm run build

FROM golang:1.10.3 as gobld

ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /go/src/github.com/redsux/addd

COPY . ./

RUN go get -u github.com/kardianos/govendor \
 && govendor sync \
 && govendor install -a -ldflags '-extldflags "-static"' +local

FROM scratch
COPY --from=gobld /go/bin/addd /addd
COPY --from=jsbld /home/node/addd-ui/dist /ui
EXPOSE 53/udp 1632/tcp 10001/udp 10001/tcp 10002/tcp
ENTRYPOINT ["/addd"]