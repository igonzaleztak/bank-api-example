# Build the Go binary
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .

# Add CA certificates and timezone data files
RUN apk add -U --no-cache ca-certificates tzdata
RUN apk add --no-cache git && apk add curl

# Install envsubst to handle .env files (envsubst is part of GNU gettext, which is small in size)
RUN apk add --no-cache gettext
    
# build image
RUN go mod download
RUN go build -tags=viper_bind_struct -o app cmd/main.go

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/app .
COPY --from=builder /app/.env .

RUN set -a && source .env && set +a

# Command to run the Go binary
CMD ["./app"]