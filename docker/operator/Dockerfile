FROM golang:1.21 as builder

ENV LANG en_US.utf8
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
# TODO(zyy17): Maybe should use the more secure container image.
FROM ubuntu:22.04 as base

WORKDIR /greptimedb-operator
COPY --from=builder /greptimedb-operator/bin/greptimedb-operator /greptimedb-operator/bin/
ENV PATH /greptimedb-operator/bin/:$PATH

ENTRYPOINT [ "greptimedb-operator" ]
