FROM golang:1.25.1 as builder

ENV LANG=en_US.utf8
WORKDIR /greptimedb-operator

# Install dependencies.
RUN apt-get update && apt-get install -y \
    curl \
    build-essential \
    pkg-config

# Build the project.
COPY . .
RUN make fast-build

# Export the binary to the clean image.
FROM ubuntu:22.04 as base

WORKDIR /greptimedb-operator
COPY --from=builder /greptimedb-operator/bin/greptimedb-operator /greptimedb-operator/bin/

# Use non-root user and group to run the operator. `65532` will be nobody in the container.
USER 65532:65532

ENV PATH=/greptimedb-operator/bin/:$PATH

ENTRYPOINT [ "greptimedb-operator" ]
