# Build stage
FROM golang:1.23.1 as builder

# Copy the entire project
WORKDIR /workspace
COPY . .

# Build the agent binary using the project's Makefile
RUN make build-agent

# Final stage
FROM ubuntu:24.10

WORKDIR /
# Copy the binary from builder stage
COPY --from=builder /workspace/bin/agent /usr/local/bin/

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/agent"]
