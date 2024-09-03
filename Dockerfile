FROM golang:1.23-alpine as builder
COPY ./ /core
WORKDIR /core
RUN go build -o /opt

FROM alpine:3.20 as runtime
COPY --from=builder /opt/docker-pull /usr/bin/docker-pull
WORKDIR /work
ENTRYPOINT ["/usr/bin/docker-pull"]



