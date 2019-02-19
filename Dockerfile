FROM golang:1.11.4-stretch

LABEL maintainer="gomeshnetwork"

COPY . /go/src/github.com/gomeshnetwork/tcc

RUN go install github.com/gomeshnetwork/tcc/cmd/tcc && rm -rf /go/src

VOLUME ["/etc/gomesh/tcc","/var/mq"]

WORKDIR /etc/gomesh/tcc

EXPOSE 2100

CMD ["/go/bin/tcc","-config","/etc/gomesh/tcc/tcc.json"]