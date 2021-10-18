FROM golang:1.17 AS builder
RUN apt-get update

# CGO deps
RUN apt-get install -y --no-install-recommends \
    g++ \
    gcc \
    libc6-dev \
    make \
    pkg-config

# our deps, bcc in debian
RUN apt-get install -y --no-install-recommends \
    libbpfcc \
    libbpfcc-dev

# add local repo into the builder
ADD . /opt/build
WORKDIR /opt/build

# build the binary there
RUN CGO_ENABLED=1 go build -tags container -o fpl -v

# begin new container
FROM alpine
WORKDIR /

# copy binary from builder to your desired location
COPY --from=builder /opt/build/fpl .
ENTRYPOINT /fpl -c config/config.yml
