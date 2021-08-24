FROM golang:1.17 AS builder

# Add your files into the container
ADD . /opt/build
WORKDIR /opt/build

# build the binary
RUN CGO_ENABLED=0 go build -tags docker -o fpl -v
FROM alpine
WORKDIR /

# COPY binary from previous stage to your desired location
COPY --from=builder /opt/build/fpl .
ENTRYPOINT /fpl -c config/config.yml
