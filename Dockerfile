FROM golang:1.17 AS builder
RUN apt-get update

# add local repo into the builder
ADD . /opt/build
WORKDIR /opt/build

# build the binary there
RUN CGO_ENABLED=0 go build -tags container -o fpl -v

# begin new container
FROM alpine
WORKDIR /

# copy binary from builder to your desired location
COPY --from=builder /opt/build/fpl .
ENTRYPOINT /fpl -c config/config.yml $0
