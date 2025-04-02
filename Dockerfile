FROM golang:1.24.1 AS builder

WORKDIR /app

COPY . .

RUN go build -o final .

FROM ubuntu:latest
WORKDIR /app

COPY --from=builder /app/final .
COPY --from=builder /app/web ./web

EXPOSE 7540

CMD ["./final"]