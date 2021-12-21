FROM golang:1.16-alpine AS build_base

# Set the Current Working Directory inside the container
WORKDIR /tmp/consul2calico

# We want to populate the module cache based on the go.{mod,sum} files.

COPY go.mod .
COPY go.sum .

COPY . .

RUN go mod tidy
RUN go mod download

# Build the Go app

RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux  go build -o ./out/consul2calico .

# Start fresh from a smaller image
FROM bitnami/minideb:bullseye

RUN mkdir -p /etc/ssl/certs/consul2calico/
RUN mkdir /etc/consulcalico/

COPY --from=build_base /tmp/consul2calico/out/consul2calico /app/consul2calico


# Run the binary program produced by `go install`
CMD ["/app/consul2calico"]
