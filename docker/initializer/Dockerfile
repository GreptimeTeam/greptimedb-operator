FROM golang:1.21 as builder

ENV LANG en_US.utf8
WORKDIR /greptimedb-operator

# Build the project.
COPY . .
RUN make initializer

# Export the binary to the clean image.
FROM ubuntu:22.04 as base

WORKDIR /greptimedb-operator
COPY --from=builder /greptimedb-operator/bin/greptimedb-initializer /greptimedb-operator/bin/

# Use non-root user and group to run the initializer. `65532` will be nobody in the container.
USER 65532:65532

ENV PATH=/greptimedb-operator/bin/:$PATH
