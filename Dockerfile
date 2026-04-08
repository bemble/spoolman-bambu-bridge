FROM golang:1.24-alpine AS builder

RUN apk add --no-cache \
    alpine-sdk \
    ca-certificates \
    tzdata

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X github.com/spoolman-bambu-bridge/backend.Version=$(cat VERSION)" \
    -o /bridge ./cmd/bridge

FROM scratch

# copy server files
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /bridge /usr/local/bin/bridge

WORKDIR /app
EXPOSE 8080

ENTRYPOINT ["bridge"]
