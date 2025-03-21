FROM golang:1.24 AS builder

WORKDIR /app
    
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o app .

FROM alpine:latest

RUN apk --no-cache add ca-certificates
RUN rm -rf /var/cache/apk/*

WORKDIR /app

COPY --from=builder /app/app .

RUN chmod +x app

ENV HTTP_HOST=:8080
ENV URLS_PATH=/data/urls.json

EXPOSE 8080
VOLUME [ "/data" ]

ENTRYPOINT ["/app/app"]