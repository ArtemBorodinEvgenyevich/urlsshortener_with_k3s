# STAGE 1: build

FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o urlshortener ./cmd/main.go

# STAGE 2: Final image
FROM alpine:3.22.2

RUN apk --no-cache add ca-certificates

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder /app/urlshortener .

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 9091

CMD ["./urlshortener"]
