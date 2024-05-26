FROM golang:1.22.1-alpine AS builder
WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o ./build/main ./cmd/main.go

FROM alpine
WORKDIR /app

COPY --from=builder /app/build/main .

EXPOSE 8080

CMD ["./main"]