FROM golang:1.25-alpine3.22 AS builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o popg

FROM alpine:latest
LABEL app=popg
LABEL org.opencontainers.image.source=https://github.com/maroda/popg
WORKDIR /app
COPY --from=builder /app/popg .
COPY web ./web
EXPOSE 1234
CMD ["./popg"]