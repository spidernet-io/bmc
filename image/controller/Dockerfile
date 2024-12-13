# Install build dependencies
FROM ubuntu:22.04 as builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    make \
    golang \
    git \
    && rm -rf /var/lib/apt/lists/*

# Copy the entire project
WORKDIR /workspace
COPY . .

# Build the controller binary using the project's Makefile
RUN make build-controller

# Final stage
FROM ubuntu:22.04

WORKDIR /
# Copy the binary from builder stage
COPY --from=builder /workspace/bin/controller /usr/local/bin/

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/controller"]
