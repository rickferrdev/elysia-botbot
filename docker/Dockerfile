FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GCO_ENABLE=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/bot

FROM scratch

COPY --from=builder /out/app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT [ "/app" ]