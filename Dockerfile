# Build stage
FROM golang:alpine AS builder
WORKDIR /app

# Disable go.work to avoid missing module errors
ENV GOWORK=off

# Copy the go.mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY ./ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOMEMLIMIT=500MiB GOGC=20 go build -p 1 -o qisur-api cmd/api/main.go

# Super lightweight final stage
FROM alpine:3.19
WORKDIR /app

# Copy the compiled binary
COPY --from=builder /app/qisur-api .

# Expose the default listening port
EXPOSE 8086

# Execute the microservice
CMD ["./qisur-api"]
