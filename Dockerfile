FROM golang:1.22-alpine as builder
COPY ./core /core
WORKDIR /core
RUN mkdir /opt/build
RUN go build -o /opt/build

FROM alpine:3.20 as runtime
COPY --from=builder /opt/build/docker-pull /usr/bin/docker-pull
WORKDIR /work
ENTRYPOINT ["/usr/bin/docker-pull"]



