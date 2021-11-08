FROM golang:1.17.2

ADD . /go/src/github.com/aleasoluciones/amievents2amqp/
WORKDIR /go/src/github.com/aleasoluciones/amievents2amqp/
RUN go mod init
RUN go mod tidy
RUN go install

ENTRYPOINT ["/go/bin/amievents2amqp"]
