FROM golang:1.23.4-alpine3.21 AS builder
ENV CGO_ENABLED=1
WORKDIR /app

RUN apk add --no-cache --update git build-base
COPY . .

RUN go build -o final .

FROM alpine:latest
WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata libc6-compat libgcc libstdc++

COPY --from=builder /app/final .
COPY --from=builder /app/web ./web

EXPOSE 7540

CMD ["./final"]