# Build stage
FROM golang:1.23.1 as builder

# Copy the entire project
WORKDIR /workspace
COPY . .

# Set GOPROXY for better download speed in China
ENV GOPROXY=https://goproxy.io

# Build the agent binary using the project's Makefile
RUN make build-agent

# Final stage
FROM ubuntu:24.10

# Install dhcpd binary
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
    isc-dhcp-server \
    iproute2 \
    isc-dhcp-client \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /
# Copy the binary from builder stage
COPY --from=builder /workspace/bin/agent /usr/local/bin/

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/agent"]
