# Stage 1 Builder Image
FROM golang:1.24.2-alpine3.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o backend

# Stage 2 Production Image
FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app /app/

EXPOSE 8080

CMD ["./backend"]
