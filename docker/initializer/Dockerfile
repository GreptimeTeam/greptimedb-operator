FROM golang:1.23 as builder

ENV LANG en_US.utf8
WORKDIR /greptimedb-operator

# Build the project.
COPY . .
RUN make initializer

# Export the binary to the clean image.
# TODO(zyy17): Maybe should use the more secure container image.
FROM ubuntu:22.04 as base

WORKDIR /greptimedb-operator
COPY --from=builder /greptimedb-operator/bin/greptimedb-initializer /greptimedb-operator/bin/
ENV PATH=/greptimedb-operator/bin/:$PATH
