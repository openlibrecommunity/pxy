FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /pxy ./cmd/pxy

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /pxy /pxy
COPY .env /app/.env
COPY .domains /app/.domains
WORKDIR /app
EXPOSE 8805
CMD ["/pxy"]
