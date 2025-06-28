FROM golang:tip-20250620-alpine3.22 AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod download


COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o capstone-core ./cmd/api/

# Final stage
FROM alpine:latest

WORKDIR /root/


COPY --from=builder /app/capstone-core .
COPY --from=builder /app/.env .


EXPOSE 8080

CMD ["./capstone-core"]