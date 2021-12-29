FROM golang:1.17.5 AS amievents2amqp-builder

ENV DEBIAN_FRONTEND noninteractive

WORKDIR /app

COPY . .

# CGO_ENABLED needed for scratch image that serves the binary
# Ref: https://stackoverflow.com/questions/61515186/when-using-cgo-enabled-is-must-and-what-happens
ENV CGO_ENABLED 0

RUN make build


#---


FROM scratch AS amievents2amqp

COPY --from=amievents2amqp-builder /app/amievents2amqp /

ENTRYPOINT ["/amievents2amqp"]
