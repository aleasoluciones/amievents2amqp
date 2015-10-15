FROM golang

ADD . /go/src/github.com/aleasoluciones/amievents2amqp/
WORKDIR /go/src/github.com/aleasoluciones/amievents2amqp/
RUN go get
RUN go install

ENTRYPOINT ["/go/bin/amievents2amqp"]
