FROM golang:1.25.5 AS builder

WORKDIR /build

RUN apt-get update && apt-get install -y curl xz-utils
RUN curl https://wasmtime.dev/install.sh -sSf | bash -s -- --version v35.0.0

COPY go.mod go.sum ./
RUN go mod download
COPY *.go /build
RUN CGO_ENABLED=0 go build -o /build/server .

# ################################################################################
FROM swift:6.1.2
RUN swift sdk install "https://github.com/swiftwasm/swift/releases/download/swift-wasm-6.1-RELEASE/swift-wasm-6.1-RELEASE-wasm32-unknown-wasi.artifactbundle.zip" --checksum "7550b4c77a55f4b637c376f5d192f297fe185607003a6212ad608276928db992"

WORKDIR /app

COPY --from=builder /root/.wasmtime/bin/wasmtime /usr/bin/wasmtime
COPY --from=builder /build/server /app/server

CMD ["/app/server"]
