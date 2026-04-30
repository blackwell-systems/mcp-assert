FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /mcp-assert ./cmd/mcp-assert

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /mcp-assert /usr/local/bin/mcp-assert
ENTRYPOINT ["mcp-assert"]
