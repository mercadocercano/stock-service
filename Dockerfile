FROM golang:1.20-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o stock-service ./cmd/api

FROM alpine:latest  

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/stock-service .

# Copy scripts
COPY ./scripts ./scripts

# Expose port
EXPOSE 8080
EXPOSE 2114

# Command to run
CMD ["./stock-service"] 