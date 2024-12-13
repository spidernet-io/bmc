# Build stage
FROM golang:1.23.1 as builder

# Copy the entire project
WORKDIR /workspace
COPY . .

# Set GOPROXY for better download speed in China
ENV GOPROXY=https://goproxy.io

# Build the controller binary using the project's Makefile
RUN make build-controller

# Final stage
FROM ubuntu:24.10

WORKDIR /
# Copy the binary from builder stage
COPY --from=builder /workspace/bin/controller /usr/local/bin/

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/controller"]
