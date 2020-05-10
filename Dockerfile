FROM golang:1.14-alpine AS builder
WORKDIR /build

# cache deps
COPY go.mod .
COPY go.sum .
RUN go mod download -x

# build 
COPY . .
RUN go build -o bin/gcs-proxy main/main.go

# pack binary to a lightweight image
FROM alpine
COPY --from=builder /build/bin/gcs-proxy /opt/gcs-proxy
ENTRYPOINT ["/opt/gcs-proxy"]
