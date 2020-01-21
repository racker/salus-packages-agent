FROM golang:1.13 as builder

WORKDIR /build

COPY go.mod go.sum /build/
RUN go mod download

COPY . .
RUN make build

FROM ubuntu:18.04

ENTRYPOINT ["/usr/bin/salus-packages-agent"]
COPY --from=builder /build/salus-packages-agent /usr/bin/
