FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/payment-platform-agent ./cmd/agent

FROM alpine:3.22

RUN adduser -D -H -u 10001 app
COPY --from=build /out/payment-platform-agent /usr/local/bin/payment-platform-agent
USER app
ENTRYPOINT ["/usr/local/bin/payment-platform-agent"]
