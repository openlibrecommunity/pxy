FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /pxy

FROM scratch
COPY --from=builder /pxy /pxy
COPY .env /app/.env
COPY .domains /app/.domains
WORKDIR /app
EXPOSE 8805
CMD ["/pxy"]
